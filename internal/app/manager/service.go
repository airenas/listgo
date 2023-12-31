package manager

import (
	"encoding/json"
	"time"

	"github.com/airenas/listgo/internal/pkg/messages"
	"github.com/airenas/listgo/internal/pkg/persistence"
	"github.com/airenas/listgo/internal/pkg/result"
	"github.com/airenas/listgo/internal/pkg/status"
	"github.com/airenas/listgo/internal/pkg/utils"

	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// ServiceData keeps data required for service work
type ServiceData struct {
	MessageSender       messages.Sender
	InformMessageSender messages.Sender
	Publisher           messages.Publisher
	StatusSaver         status.Saver
	ResultSaver         ResultSaver
	DecodeCh            <-chan amqp.Delivery
	SplitChannelsCh     <-chan amqp.Delivery
	AudioConvertCh      <-chan amqp.Delivery
	DiarizationCh       <-chan amqp.Delivery
	TranscriptionCh     <-chan amqp.Delivery
	RescoreCh           <-chan amqp.Delivery
	ResultMakeCh        <-chan amqp.Delivery
	fc                  *utils.MultiCloseChannel
	speechIndicator     SpeechIndicator
}

// SpeechIndicator looks if request audio has speech
type SpeechIndicator interface {
	Test(string) (bool, error)
}

// return true if it can be redelivered
type prFunc func(d *amqp.Delivery, data *ServiceData) (bool, error)

// StartWorkerService starts the event queue listener service to listen for events
func StartWorkerService(data *ServiceData) error {
	if data.ResultSaver == nil {
		return errors.New("Result saver not provided")
	}
	if data.Publisher == nil {
		return errors.New("Publisher not provided")
	}
	if data.MessageSender == nil {
		return errors.New("MessageSender not provided")
	}
	if data.InformMessageSender == nil {
		return errors.New("InformMessageSender not provided")
	}
	if data.speechIndicator == nil {
		return errors.New("speechIndicator not provided")
	}

	cmdapp.Log.Infof("Starting listen for messages")

	go listenQueue(data.DecodeCh, decode, data)
	go listenQueue(data.SplitChannelsCh, splitChannelsFinish, data)
	go listenQueue(data.AudioConvertCh, audioConvertFinish, data)
	go listenQueue(data.DiarizationCh, diarizationFinish, data)
	go listenQueue(data.TranscriptionCh, transcriptionFinish, data)
	go listenQueue(data.RescoreCh, rescoreFinish, data)
	go listenQueue(data.ResultMakeCh, resultMakeFinish, data)

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

// decode starts the transcription process
// workflow:
// 1. set status to STARTED
// 2. send 'Started' event (async)
// 3.1. send and wait for 'AudioConvert' to finish
// or
// 3.2  send msg to 'SplitChannels'
func decode(d *amqp.Delivery, data *ServiceData) (bool, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "can't unmarshal message "+string(d.Body))
	}

	cmdapp.Log.Infof("Got %s msg :%s (%s)", messages.Decode, message.ID, message.Recognizer)

	st := status.AudioConvert
	target := messages.AudioConvert
	if sepCh, ok := messages.GetTag(message.Tags, messages.TagSepSpeakersOnChannel); ok && utils.ParamTrue(sepCh) {
		cmdapp.Log.Infof("Separate speakers on channels")
		st = status.SplitChannels
		target = messages.SplitChannels
	}

	if err := data.StatusSaver.Save(message.ID, st); err != nil {
		cmdapp.Log.Error(err)
		return true, err
	}
	publishStatusChange(&message, data)
	err := data.InformMessageSender.Send(newInformMessage(&message, messages.InformTypeStarted), messages.Inform, "")
	if err != nil {
		return true, err
	}
	return true, data.MessageSender.Send(messages.NewQueueMessageFromM(&message), target, messages.ResultQueueFor(target))
}

// splitChannelsFinish processes split channel finish message
// 1. logs status
// 2. sends 'DecodeMultiple' message
func splitChannelsFinish(d *amqp.Delivery, data *ServiceData) (bool, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "can't unmarshal message "+string(d.Body))
	}
	c, err := processStatus(&message, data, messages.SplitChannels, status.AudioConvert) // there is no DecodeMultiple status
	if !c {
		if err != nil {
			cmdapp.Log.Error(err)
		}
		return true, err
	}
	return true, data.MessageSender.Send(messages.NewQueueMessageFromM(&message), messages.DecodeMultiple, "")
}

// audioConvertFinish processes audio convert result messages
// 1. logs status
// 2. sends 'Diarization' message
func audioConvertFinish(d *amqp.Delivery, data *ServiceData) (bool, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "can't unmarshal message "+string(d.Body))
	}
	c, err := processStatus(&message, data, messages.AudioConvert, status.Diarization)
	if !c {
		if err != nil {
			cmdapp.Log.Error(err)
		}
		return true, err
	}
	return true, data.MessageSender.Send(messages.NewQueueMessageFromM(&message),
		messages.Diarization, messages.ResultQueueFor(messages.Diarization))
}

// diarizationFinish processes audio diarization result messages
// 1. logs status
// 2. sends 'Transctiption' message
func diarizationFinish(d *amqp.Delivery, data *ServiceData) (bool, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "Can't unmarshal message "+string(d.Body))
	}
	nextTask := messages.Transcription
	nextStatus := status.Transcription
	if noSpeech(message.ID, data) {
		cmdapp.Log.Info("No speech detected. Skip Transcription and Rescore steps")
		message.Tags = append(message.Tags, messages.NewTag("NO_SPEECH", "true"))
		nextTask = messages.ResultMake
		nextStatus = status.ResultMake
	}
	c, err := processStatus(&message, data, messages.Diarization, nextStatus)
	if !c {
		if err != nil {
			cmdapp.Log.Error(err)
		}
		return true, err
	}
	return true, data.MessageSender.Send(messages.NewQueueMessageFromM(&message), nextTask, messages.ResultQueueFor(nextTask))
}

// transcriptionFinish processes transcription result messages
// 1. logs status
// 2. sends 'Rescore' message
func transcriptionFinish(d *amqp.Delivery, data *ServiceData) (bool, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "Can't unmarshal message "+string(d.Body))
	}
	c, err := processStatus(&message, data, messages.Transcription, status.Rescore)
	if !c {
		if err != nil {
			cmdapp.Log.Error(err)
		}
		return true, err
	}
	return true, data.MessageSender.Send(messages.NewQueueMessageFromM(&message),
		messages.Rescore, messages.ResultQueueFor(messages.Rescore))
}

// rescoreFinish processes rescore result messages
// 1. logs status
// 2. sends 'ResultMake' message
func rescoreFinish(d *amqp.Delivery, data *ServiceData) (bool, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "Can't unmarshal message "+string(d.Body))
	}
	c, err := processStatus(&message, data, messages.Rescore, status.ResultMake)
	if !c {
		if err != nil {
			cmdapp.Log.Error(err)
		}
		return true, err
	}
	return true, data.MessageSender.Send(messages.NewQueueMessageFromM(&message),
		messages.ResultMake, messages.ResultQueueFor(messages.ResultMake))
}

// transcriptionFinish processes transcription result messages
// 1. logs status
// 2. sends 'FinishDecode' message
func resultMakeFinish(d *amqp.Delivery, data *ServiceData) (bool, error) {
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
			persistence.StAvailableResults: []string{result.Txt, result.TxtFinal,
				result.Lat, result.LatGz,
				result.LatRestored, result.LatRestoredGz, result.WebVTT}}, nil)
		if err != nil {
			cmdapp.Log.Error(err)
			return true, err
		}
	}
	c, err := processStatus(&message.QueueMessage, data, messages.ResultMake, status.Completed)
	if !c {
		if err != nil {
			cmdapp.Log.Error(err)
		}
		return true, err
	}
	if tq, ok := messages.GetTag(message.Tags, messages.TagResultQueue); ok {
		err := data.MessageSender.Send(&message.QueueMessage, tq, "")
		cmdapp.LogIf(err)
	}
	return true, data.InformMessageSender.Send(newInformMessage(&message.QueueMessage, messages.InformTypeFinished),
		messages.Inform, "")
}

// processStatus analyzes message response and saves status
// returns false if no futher processing is needed
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

func noSpeech(ID string, data *ServiceData) bool {
	fileNonEmpty, err := data.speechIndicator.Test(ID)
	if err != nil {
		cmdapp.Log.Error(err)
		return false
	}
	return !fileNonEmpty
}

func sendInformFailure(message *messages.QueueMessage, data *ServiceData) {
	cmdapp.Log.Infof("Trying send inform msg about failure %s", message.ID)
	err := data.InformMessageSender.Send(newInformMessage(message, messages.InformTypeFailed), messages.Inform, "")
	cmdapp.LogIf(err)
	if tq, ok := messages.GetTag(message.Tags, messages.TagResultQueue); ok {
		msg := messages.NewQueueMessageFromM(message)
		err := data.MessageSender.Send(msg, tq, "")
		cmdapp.LogIf(err)
	}
}

func publishStatusChange(message *messages.QueueMessage, data *ServiceData) {
	cmdapp.Log.Infof("Publishing status change %s", message.ID)
	err := data.Publisher.Publish(message.ID, messages.TopicStatusChange)
	cmdapp.LogIf(err)
	if tq, ok := messages.GetTag(message.Tags, messages.TagStatusQueue); ok {
		msg := messages.NewQueueMessageFromM(message)
		err := data.MessageSender.Send(msg, tq, "")
		cmdapp.LogIf(err)
	}
}

func newInformMessage(msg *messages.QueueMessage, it string) *messages.InformMessage {
	return &messages.InformMessage{QueueMessage: messages.QueueMessage{ID: msg.ID, Recognizer: msg.Recognizer, Tags: msg.Tags},
		Type: it, At: time.Now().UTC()}
}
