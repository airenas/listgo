package zoom

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/airenas/listgo/internal/app/manager"
	"github.com/airenas/listgo/internal/app/result"
	stapi "github.com/airenas/listgo/internal/app/status/api"
	"github.com/airenas/listgo/internal/app/upload"
	"github.com/airenas/listgo/internal/pkg/messages"
	"github.com/airenas/listgo/internal/pkg/persistence"
	resultConst "github.com/airenas/listgo/internal/pkg/result"
	"github.com/airenas/listgo/internal/pkg/status"
	"github.com/airenas/listgo/internal/pkg/utils"

	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type (
	// StatusProvider returns status by ID
	StatusProvider interface {
		Get(ID string) (*stapi.TranscriptionResult, error)
	}
	// FilesGetter retrieves file for ID
	FilesGetter interface {
		List(ID string) ([]string, error)
	}
	// WorkPersistence save, gets data to DB
	WorkPersistence interface {
		Save(*persistence.WorkData) error
		Get(ID string) (*persistence.WorkData, error)
	}
	// AudioDuration provides audio len for file/io.Reader
	AudioDuration interface {
		Get(string, io.Reader) (time.Duration, error)
	}
	// ServiceData keeps data required for service work
	ServiceData struct {
		MessageSender       messages.Sender
		InformMessageSender messages.Sender
		Publisher           messages.Publisher
		StatusSaver         status.Saver
		StatusProvider      StatusProvider
		ResultSaver         manager.ResultSaver
		FilesGetter         FilesGetter
		Loader              result.FileLoader
		AudioLen            AudioDuration
		FileSaver           upload.FileSaver
		RequestSaver        upload.RequestSaver
		DB                  WorkPersistence
		DecodeMultiCh       <-chan amqp.Delivery
		JoinAudioCh         <-chan amqp.Delivery
		JoinResultsCh       <-chan amqp.Delivery
		OneCompletedCh      <-chan amqp.Delivery
		OneStatusCh         <-chan amqp.Delivery
		fc                  *utils.MultiCloseChannel
	}
)

// return true if it can be redelivered
type prFunc func(d *amqp.Delivery, data *ServiceData) (bool, error)

// StartWorkerService starts the event queue listener service to listen for events
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
		return errors.New("filesGetter not provided")
	}
	if data.Loader == nil {
		return errors.New("loader not provided")
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
	if data.DB == nil {
		return errors.New("db not provided")
	}
	if data.StatusProvider == nil {
		return errors.New("status provider not provided")
	}

	cmdapp.Log.Infof("Starting listen for messages")

	go listenQueue(data.DecodeMultiCh, decode, data)
	go listenQueue(data.OneStatusCh, gotStatus, data)
	go listenQueue(data.OneCompletedCh, completed, data)
	go listenQueue(data.JoinAudioCh, mergedAudio, data)
	go listenQueue(data.JoinResultsCh, joinResultsFinish, data)

	return nil
}

func listenQueue(q <-chan amqp.Delivery, f prFunc, data *ServiceData) {
	for d := range q {
		redeliver, err := f(&d, data)
		if err != nil {
			cmdapp.Log.Errorf("can't process message %s\n%s", d.MessageId, string(d.Body))
			cmdapp.Log.Error(err)
			d.Nack(false, redeliver && !d.Redelivered) // redeliver for first time
		} else {
			d.Ack(false)
		}
	}
	cmdapp.Log.Infof("Stopped listening queue")
	data.fc.Close()
}

// decode starts the transcription process
// workflow:
// 1. send "Started" event
// 2. validate file lengths
// 3. copy each file for transcription
// 4. send 'Decode' event for each file
func decode(d *amqp.Delivery, data *ServiceData) (bool, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "can't unmarshal message "+string(d.Body))
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
		err := data.StatusSaver.SaveF(message.ID, map[string]interface{}{
			persistence.StError:     "Files length differ",
			persistence.StErrorCode: "LEN_DIFFER",
		}, nil)
		if err != nil {
			cmdapp.Log.Error(err)
			return true, err
		}
		publishStatusChange(&message, data)
		sendInformFailure(&message, data)
		return false, nil
	}

	err = data.InformMessageSender.Send(newInformMessage(&message, messages.InformTypeStarted),
		messages.Inform, "")
	if err != nil {
		return true, err
	}

	if sepCh, ok := messages.GetTag(message.Tags, messages.TagSepSpeakersOnChannel); ok && utils.ParamTrue(sepCh) {
		cmdapp.Log.Infof("Separate speakers on channels")
	} else {
		if err := data.MessageSender.Send(messages.NewQueueMessageFromM(&message), messages.JoinAudio,
			messages.ResultQueueFor(messages.JoinAudio)); err != nil {
			return true, err
		}
	}

	ids, fNames, err := startTranscriptions(data, files, &message)
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
	if err := data.DB.Save(&persistence.WorkData{ID: message.ID, Related: ids, FileNames: fNames}); err != nil {
		return true, errors.Wrapf(err, "can't save related ids")
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
		defer bData.Close()
		fl, err := data.AudioLen.Get(f, bData)
		if err != nil {
			return false, err
		}
		if i == 0 {
			len = fl
		}
		if !cmpDur(len, fl) {
			cmdapp.Log.Infof("File len differs %s vs %s", len.String(), fl.String())
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

func startTranscriptions(data *ServiceData, files []string, message *messages.QueueMessage) ([]string, []string, error) {
	res := make([]string, 0)
	resF := make([]string, 0)
	for _, f := range files {
		id, err := startTranscription(data, f, message)
		if err != nil {
			return nil, nil, err
		}
		res = append(res, id)
		resF = append(resF, filepath.Base(f))
	}
	return res, resF, nil
}

func startTranscription(data *ServiceData, file string, message *messages.QueueMessage) (string, error) {
	bData, err := data.Loader.Load(file)
	if err != nil {
		return "", err
	}
	defer bData.Close()

	id := uuid.New().String()
	ext := filepath.Ext(file)
	fileName := id + ext

	err = data.RequestSaver.Save(&persistence.Request{ID: id, File: fileName, RecognizerID: message.Recognizer})
	if err != nil {
		return "", errors.Wrapf(err, "can't save request")
	}

	err = data.StatusSaver.SaveF(id, map[string]interface{}{"status": status.Name(status.Uploaded),
		persistence.StAudioReady: true}, nil)
	if err != nil {
		return "", errors.Wrapf(err, "can't save status")
	}

	err = data.FileSaver.Save(fileName, bData)
	if err != nil {
		return "", errors.Wrapf(err, "can't save file")
	}

	tags := make([]messages.Tag, 0)
	for _, t := range message.Tags {
		if t.Key == messages.TagNumberOfSpeakers || t.Key == messages.TagTimestamp ||
			t.Key == messages.TagSepSpeakersOnChannel {
			continue
		}
		tags = append(tags, t)
	}
	tags = append(tags, messages.NewTag(messages.TagNumberOfSpeakers, "1"),
		messages.NewTag(messages.TagTimestamp, strconv.FormatInt(time.Now().Unix(), 10)),
		messages.NewTag(messages.TagParentID, message.ID),
		messages.NewTag(messages.TagStatusQueue, messages.OneStatus),
		messages.NewTag(messages.TagResultQueue, messages.OneCompleted),
	)

	return id, data.MessageSender.Send(messages.NewQueueMessage(id, message.Recognizer, tags), messages.Decode, "")
}

func sendInformFailure(message *messages.QueueMessage, data *ServiceData) {
	cmdapp.Log.Infof("Trying send inform msg about failure %s", message.ID)
	err := data.InformMessageSender.Send(newInformMessage(message, messages.InformTypeFailed), messages.Inform, "")
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

// gotStatus process status msgs from child transcriptions
func gotStatus(d *amqp.Delivery, data *ServiceData) (bool, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "can't unmarshal message "+string(d.Body))
	}

	cmdapp.Log.Infof("Got %s msg :%s (%s)", messages.OneStatus, message.ID, message.Recognizer)
	pID, ok := messages.GetTag(message.Tags, messages.TagParentID)
	if !ok {
		return false, errors.New("no parent ID")
	}
	cmdapp.Log.Infof("Parent ID %s", pID)
	wd, err := data.DB.Get(pID)
	if err != nil {
		return true, errors.Wrapf(err, "can't load work data")
	}
	st, err := data.StatusProvider.Get(pID)
	if err != nil {
		return true, errors.Wrapf(err, "can't load status")
	}
	if st.Error != "" || st.ErrorCode != "" { // already failed
		cmdapp.Log.Infof("Skip ID %s - already failed", pID)
		return false, nil
	}

	nStatus := status.Completed
	msg := messages.NewQueueMessage(pID, message.Recognizer, message.Tags)
	for _, id := range wd.Related {
		cSt, err := data.StatusProvider.Get(id)
		cStatus := status.From(cSt.Status)
		if err != nil {
			return true, errors.Wrapf(err, "can't load status")
		}
		if cSt.Error != "" || cSt.ErrorCode != "" {
			msg.Error = cSt.Error
			if msg.Error == "" {
				msg.Error = cSt.ErrorCode
			}
			c, err := processStatus(msg, data, messages.OneStatus, status.JoinResults)
			if !c {
				if err != nil {
					cmdapp.Log.Error(err)
				}
				return true, err
			}
			break
		}
		nStatus = status.Min(nStatus, cStatus)
	}
	if nStatus != status.From(st.Status) && nStatus != status.Completed {
		c, err := processStatus(msg, data, messages.OneStatus, nStatus)
		if !c {
			if err != nil {
				cmdapp.Log.Error(err)
			}
			return true, err
		}
	}
	return false, nil
}

// completed process result msgs from child transcriptions
func completed(d *amqp.Delivery, data *ServiceData) (bool, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "can't unmarshal message "+string(d.Body))
	}

	cmdapp.Log.Infof("Got %s msg :%s (%s)", messages.OneCompleted, message.ID, message.Recognizer)
	pID, ok := messages.GetTag(message.Tags, messages.TagParentID)
	if !ok {
		return false, errors.New("no parent ID")
	}
	cmdapp.Log.Infof("Parent ID %s", pID)
	wd, err := data.DB.Get(pID)
	if err != nil {
		return true, errors.Wrapf(err, "can't load work data")
	}
	st, err := data.StatusProvider.Get(pID)
	if err != nil {
		return true, errors.Wrapf(err, "can't load status")
	}
	if st.Error != "" || st.ErrorCode != "" { // already failed
		cmdapp.Log.Infof("Skip ID %s - already failed", pID)
		return false, nil
	}

	done := len(wd.Related) > 0
	for _, id := range wd.Related {
		cSt, err := data.StatusProvider.Get(id)
		if err != nil {
			return true, errors.Wrapf(err, "can't load status")
		}
		if cSt.Error != "" || cSt.ErrorCode != "" {
			msg := messages.NewQueueMessage(pID, message.Recognizer, message.Tags)
			msg.Error = cSt.Error
			if msg.Error == "" {
				msg.Error = cSt.ErrorCode
			}
			c, err := processStatus(msg, data, messages.OneCompleted, status.JoinResults)
			if !c {
				if err != nil {
					cmdapp.Log.Error(err)
				}
				return true, err
			}
			done = false
			break
		}
		if status.From(cSt.Status) != status.Completed {
			cmdapp.Log.Debugf("Not finished ID %s", id)
			done = false
		}
	}

	if done {
		msg := messages.NewQueueMessage(pID, message.Recognizer, message.Tags)
		c, err := processStatus(msg, data, messages.OneCompleted, status.JoinResults)
		if !c {
			if err != nil {
				cmdapp.Log.Error(err)
			}
			return true, err
		}

		tags := message.Tags
		tags = append(tags, messages.NewTag(messages.TagChildIDS, strings.Join(wd.Related, " ")))
		ifm, err := makeIDsFnMap(wd.Related, wd.FileNames)
		if err != nil {
			cmdapp.Log.Error(err)
		}
		tags = append(tags, messages.NewTag(messages.TagChildIDSFileNames, ifm))
		err = data.MessageSender.Send(messages.NewQueueMessage(pID, message.Recognizer, tags), messages.JoinResults,
			messages.ResultQueueFor(messages.JoinResults))
		if err != nil {
			return true, err
		}
	}

	return false, nil
}

func makeIDsFnMap(ids, fns []string) (string, error) {
	res := strings.Builder{}
	if len(ids) != len(fns) {
		return "", errors.New("ids and file names len does not match")
	}
	for i, id := range ids {
		if i > 0 {
			res.WriteString(";")
		}
		res.WriteString(fmt.Sprintf("%s=%s", id, fns[i]))
	}
	return res.String(), nil
}

func processStatus(message *messages.QueueMessage, data *ServiceData, from string, to status.Status) (bool, error) {
	cmdapp.Log.Infof("Got %s msg :%s (%s)", from, message.ID, message.Recognizer)
	if message.Error != "" {
		err := data.StatusSaver.SaveError(message.ID, message.Error)
		if err != nil {
			cmdapp.Log.Error(err)
			return false, err
		}
		publishStatusChange(message, data)
		sendInformFailure(message, data)
		return false, nil
	}
	err := data.StatusSaver.Save(message.ID, to)
	if err != nil {
		cmdapp.Log.Error(err)
		sendInformFailure(message, data)
		return false, err
	}
	publishStatusChange(message, data)
	return true, nil
}

// mergedAudio process audio merge event result
func mergedAudio(d *amqp.Delivery, data *ServiceData) (bool, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "can't unmarshal message "+string(d.Body))
	}

	cmdapp.Log.Infof("Got %s msg :%s (%s)", messages.ResultQueueFor(messages.JoinAudio), message.ID, message.Recognizer)
	st, err := data.StatusProvider.Get(message.ID)
	if err != nil {
		return true, errors.Wrapf(err, "can't load status")
	}
	if st.Error != "" || st.ErrorCode != "" { // already failed
		cmdapp.Log.Infof("Skip ID %s - already failed", message.ID)
		return false, nil
	}

	if message.Error != "" {
		err := data.StatusSaver.SaveError(message.ID, message.Error)
		if err != nil {
			cmdapp.Log.Error(err)
			return false, err
		}
		publishStatusChange(&message, data)
		sendInformFailure(&message, data)
		return false, nil
	}
	err = data.StatusSaver.SaveF(message.ID, map[string]interface{}{persistence.StAudioReady: true}, nil)
	if err != nil {
		cmdapp.Log.Error(err)
		sendInformFailure(&message, data)
		return false, err
	}
	publishStatusChange(&message, data)
	return true, nil
}

// transcriptionFinish processes transcription result messages
// 1. logs status
// 2. sends 'FinishDecode' message
func joinResultsFinish(d *amqp.Delivery, data *ServiceData) (bool, error) {
	var message messages.ResultMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "Can't unmarshal message "+string(d.Body))
	}
	if message.Error == "" {
		err := data.ResultSaver.Save(message.ID, message.Result)
		if err != nil {
			cmdapp.Log.Error(err)
			return true, err
		}
		err = data.StatusSaver.SaveF(message.ID, map[string]interface{}{
			persistence.StAvailableResults: []string{resultConst.TxtFinal, resultConst.WebVTT}}, nil)
		if err != nil {
			cmdapp.Log.Error(err)
			return true, err
		}
	}
	c, err := processStatus(&message.QueueMessage, data, messages.JoinAudio, status.Completed)
	if !c {
		if err != nil {
			cmdapp.Log.Error(err)
		}
		return true, err
	}
	return true, data.InformMessageSender.Send(newInformMessage(&message.QueueMessage, messages.InformTypeFinished),
		messages.Inform, "")
}
