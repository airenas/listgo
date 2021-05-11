package zoom

import (
	"bytes"
	"encoding/json"
	"io"
	"path/filepath"
	"strconv"
	"time"

	"bitbucket.org/airenas/listgo/internal/app/manager"
	"bitbucket.org/airenas/listgo/internal/app/upload"
	"bitbucket.org/airenas/listgo/internal/app/upload/api"
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/status"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type (
	FilesGetter interface {
		List(ID string) ([]string, error)
	}
	FileLoader interface {
		Load(file string) ([]byte, error)
	}
	FileLen interface {
		Get(rd io.Reader) (time.Duration, error)
	}
	// ServiceData keeps data required for service work
	ServiceData struct {
		MessageSender       messages.Sender
		InformMessageSender messages.Sender
		Publisher           messages.Publisher
		StatusSaver         status.Saver
		ResultSaver         manager.ResultSaver
		FilesGetter         FilesGetter
		Loader              FileLoader
		AudioLen            FileLen
		FileSaver           upload.FileSaver
		RequestSaver        upload.RequestSaver
		DecodeMultiCh       <-chan amqp.Delivery
		JoinAudioCh         <-chan amqp.Delivery
		JoinResultsCh       <-chan amqp.Delivery
		OneCompletedCh      <-chan amqp.Delivery
		fc                  *utils.MultiCloseChannel
	}
)

//return true if it can be redelivered
type prFunc func(d *amqp.Delivery, data *ServiceData) (bool, error)

//StartWorkerService starts the event queue listener service to listen for events
func StartWorkerService(data *ServiceData) error {
	if data.ResultSaver == nil {
		return errors.New("result saver not provided")
	}
	if data.Publisher == nil {
		return errors.New("publisher not provided")
	}
	if data.MessageSender == nil {
		return errors.New("messageSender not provided")
	}
	if data.InformMessageSender == nil {
		return errors.New("informMessageSender not provided")
	}
	if data.StatusSaver == nil {
		return errors.New("statusSaver not provided")
	}
	if data.FilesGetter == nil {
		return errors.New("FilesGetter not provided")
	}
	if data.Loader == nil {
		return errors.New("Loader not provided")
	}
	if data.AudioLen == nil {
		return errors.New("audio len service not provided")
	}
	if data.FileSaver == nil {
		return errors.New("fileSaver not provided")
	}
	if data.RequestSaver == nil {
		return errors.New("requestSaver not provided")
	}

	cmdapp.Log.Infof("Starting listen for messages")

	go listenQueue(data.DecodeMultiCh, decode, data)

	return nil
}

func listenQueue(q <-chan amqp.Delivery, f prFunc, data *ServiceData) {
	for d := range q {
		redeliver, err := f(&d, data)
		if err != nil {
			cmdapp.Log.Errorf("Can't process message %s\n%s", d.MessageId, string(d.Body))
			cmdapp.Log.Error(err)
			d.Nack(false, redeliver && !d.Redelivered) // redeliver for first time
		} else {
			d.Ack(false)
		}
	}
	cmdapp.Log.Infof("Stopped listening queue")
	data.fc.Close()
}

//decode starts the transcription process
// workflow:
// 1. send "Started" event
// 2. validate file lengths
// 3. copy each file for transcription
// 4. send 'Decode' event for each file

func decode(d *amqp.Delivery, data *ServiceData) (bool, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "Can't unmarshal message "+string(d.Body))
	}

	cmdapp.Log.Infof("Got %s msg :%s (%s)", messages.DecodeMultiple, message.ID, message.Recognizer)

	files, err := data.FilesGetter.List(message.ID)
	if err != nil {
		return true, err
	}
	ok, err := validateLen(data, files)
	if err != nil {
		return true, err
	}
	if !ok {
		err := data.StatusSaver.SaveError(message.ID, "Files len differ")
		if err != nil {
			cmdapp.Log.Error(err)
			return true, err
		}
		publishStatusChange(&message, data)
		sendInformFailure(&message, data)
		return false, nil
	}

	err = data.InformMessageSender.Send(newInformMessage(&message, messages.InformType_Started),
		messages.Inform, "")
	if err != nil {
		return true, err
	}

	err = startTranscriptions(data, files, &message)
	if err != nil {
		if d.Redelivered {
			if err := data.StatusSaver.SaveError(message.ID, "Can't start transcription. "+err.Error()); err != nil {
				cmdapp.Log.Error(err)
				return true, err
			}
			publishStatusChange(&message, data)
			sendInformFailure(&message, data)
			return false, err
		}
		return true, err
	}
	return false, nil
}

// validateLen returns false if file len differs
func validateLen(data *ServiceData, files []string) (bool, error) {
	var len time.Duration
	for i, f := range files {
		bData, err := data.Loader.Load(f)
		if err != nil {
			return false, err
		}
		fl, err := data.AudioLen.Get(bytes.NewBuffer(bData))
		if err != nil {
			return false, err
		}
		if i == 0 {
			len = fl
		}
		if !cmpDur(len, fl) {
			cmdapp.Log.Info("File len differs %s vs %s", len.String(), fl.String())
			return false, nil
		}
	}
	return true, nil
}

func cmpDur(d1, d2 time.Duration) bool {
	diff := d1 - d2
	if diff < 0 {
		diff = -diff
	}
	return diff < time.Second
}

func startTranscriptions(data *ServiceData, files []string, message *messages.QueueMessage) error {
	for _, f := range files {
		err := startTranscription(data, f, message)
		if err != nil {
			return err
		}
	}
	return nil
}

func startTranscription(data *ServiceData, file string, message *messages.QueueMessage) error {
	bData, err := data.Loader.Load(file)
	if err != nil {
		return err
	}

	id := uuid.New().String()
	ext := filepath.Ext(file)
	fileName := id + ext

	err = data.RequestSaver.Save(api.RequestData{ID: id, File: fileName, RecognizerID: message.Recognizer})
	if err != nil {
		return err
	}

	err = data.StatusSaver.Save(id, status.Uploaded)
	if err != nil {
		return err
	}

	err = data.FileSaver.Save(fileName, bytes.NewBuffer(bData))
	if err != nil {
		return err
	}

	tags := make([]messages.Tag, 0)
	for _, t := range message.Tags {
		if t.Key == messages.TagNumberOfSpeakers || t.Key == messages.TagTimestamp {
			continue
		}
		tags = append(tags, t)
	}
	tags = append(tags, messages.NewTag(messages.TagNumberOfSpeakers, "1"),
		messages.NewTag(messages.TagTimestamp, strconv.FormatInt(time.Now().Unix(), 10)))

	return data.MessageSender.Send(messages.NewQueueMessage(id, message.Recognizer, tags), messages.Decode, "")
}

func sendInformFailure(message *messages.QueueMessage, data *ServiceData) {
	cmdapp.Log.Infof("Trying send inform msg about failure %s", message.ID)
	err := data.InformMessageSender.Send(newInformMessage(message, messages.InformType_Failed), messages.Inform, "")
	cmdapp.LogIf(err)
}

func publishStatusChange(message *messages.QueueMessage, data *ServiceData) {
	cmdapp.Log.Infof("Publishing status change %s", message.ID)
	err := data.Publisher.Publish(message.ID, messages.TopicStatusChange)
	cmdapp.LogIf(err)
}

func newInformMessage(msg *messages.QueueMessage, it string) *messages.InformMessage {
	return &messages.InformMessage{QueueMessage: messages.QueueMessage{ID: msg.ID, Recognizer: msg.Recognizer, Tags: msg.Tags},
		Type: it, At: time.Now().UTC()}
}
