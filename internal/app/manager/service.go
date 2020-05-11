package manager

import (
	"encoding/json"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/status"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
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
	AudioConvertCh      <-chan amqp.Delivery
	DiarizationCh       <-chan amqp.Delivery
	TranscriptionCh     <-chan amqp.Delivery
	RescoreCh           <-chan amqp.Delivery
	ResultMakeCh        <-chan amqp.Delivery
	fc                  *utils.MultiCloseChannel
}

//return true if it can be redelivered
type prFunc func(d *amqp.Delivery, data *ServiceData) (bool, error)

//StartWorkerService starts the event queue listener service to listen for events
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

	cmdapp.Log.Infof("Starting listen for messages")

	go listenQueue(data.DecodeCh, decode, data)
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

//decode starts the transcription process
// workflow:
// 1. set status to STARTED
// 2. send 'Started' event (async)
// 3. send and wait for 'AudioConvert' to finish
func decode(d *amqp.Delivery, data *ServiceData) (bool, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "Can't unmarshal message "+string(d.Body))
	}

	cmdapp.Log.Infof("Got %s msg :%s (%s)", messages.Decode, message.ID, message.Recognizer)
	err := data.StatusSaver.Save(message.ID, status.AudioConvert)
	if err != nil {
		cmdapp.Log.Error(err)
		return true, err
	}
	publishStatusChange(&message, data)
	err = data.InformMessageSender.Send(newInformMessage(&message, messages.InformType_Started),
		messages.Inform, "")
	if err != nil {
		return true, err
	}
	return true, data.MessageSender.Send(messages.NewQueueMessageFromM(&message),
		messages.AudioConvert, messages.ResultQueueFor(messages.AudioConvert))
}

//audioConvertFinish processes audio convert result messages
// 1. logs status
// 2. sends 'Diarization' message
func audioConvertFinish(d *amqp.Delivery, data *ServiceData) (bool, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "Can't unmarshal message "+string(d.Body))
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

//diarizationFinish processes audio diarization result messages
// 1. logs status
// 2. sends 'Transctiption' message
func diarizationFinish(d *amqp.Delivery, data *ServiceData) (bool, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "Can't unmarshal message "+string(d.Body))
	}
	c, err := processStatus(&message, data, messages.Diarization, status.Transcription)
	if !c {
		if err != nil {
			cmdapp.Log.Error(err)
		}
		return true, err
	}
	return true, data.MessageSender.Send(messages.NewQueueMessageFromM(&message),
		messages.Transcription, messages.ResultQueueFor(messages.Transcription))
}

//transcriptionFinish processes transcription result messages
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

//rescoreFinish processes rescore result messages
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

//transcriptionFinish processes transcription result messages
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
	}
	c, err := processStatus(&message.QueueMessage, data, messages.ResultMake, status.Completed)
	if !c {
		if err != nil {
			cmdapp.Log.Error(err)
		}
		return true, err
	}
	return true, data.InformMessageSender.Send(newInformMessage(&message.QueueMessage, messages.InformType_Finished),
		messages.Inform, "")
}

//processStatus analyzes message response and saves status
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
