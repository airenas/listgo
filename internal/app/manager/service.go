package manager

import (
	"encoding/json"
	"sync"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/status"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type multiCloseChannel struct {
	c    chan struct{}
	once sync.Once
}

func newMultiCloseChannel() *multiCloseChannel {
	return &multiCloseChannel{c: make(chan struct{})}
}

func (mc *multiCloseChannel) close() {
	mc.once.Do(func() {
		close(mc.c)
	})
}

// ServiceData keeps data required for service work
type ServiceData struct {
	MessageSender   messages.Sender
	Publisher       messages.Publisher
	StatusSaver     status.Saver
	ResultSaver     ResultSaver
	DecodeCh        <-chan amqp.Delivery
	AudioConvertCh  <-chan amqp.Delivery
	DiarizationCh   <-chan amqp.Delivery
	TranscriptionCh <-chan amqp.Delivery
	ResultMakeCh    <-chan amqp.Delivery
}

//return true if it can be redelivered
type prFunc func(d *amqp.Delivery, data *ServiceData) (bool, error)

//StartWorkerService starts the event queue listener service to listen for events
func StartWorkerService(data *ServiceData) (<-chan struct{}, error) {
	if data.ResultSaver == nil {
		return nil, errors.New("Result saver not provided")
	}
	if data.Publisher == nil {
		return nil, errors.New("Publisher not provided")
	}

	cmdapp.Log.Infof("Starting listen for messages")

	fc := newMultiCloseChannel()

	go listenQueue(data.DecodeCh, decode, data, fc)
	go listenQueue(data.AudioConvertCh, audioConvertFinish, data, fc)
	go listenQueue(data.DiarizationCh, diarizationFinish, data, fc)
	go listenQueue(data.TranscriptionCh, transcriptionFinish, data, fc)
	go listenQueue(data.ResultMakeCh, resultMakeFinish, data, fc)

	return fc.c, nil
}

func listenQueue(q <-chan amqp.Delivery, f prFunc, data *ServiceData, fc *multiCloseChannel) {
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
	fc.close()
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

	cmdapp.Log.Infof("Got %s msg :%s", messages.Decode, message.ID)
	err := data.StatusSaver.Save(message.ID, status.AudioConvert)
	if err != nil {
		cmdapp.Log.Error(err)
		return true, err
	}
	publishStatusChange(&message, data)
	err = data.MessageSender.Send(newInformMessage(message.ID, messages.InformType_Started),
		messages.Inform, "")
	if err != nil {
		return true, err
	}
	return true, data.MessageSender.Send(messages.NewQueueMessage(message.ID),
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
	return true, data.MessageSender.Send(messages.NewQueueMessage(message.ID),
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
	return true, data.MessageSender.Send(messages.NewQueueMessage(message.ID),
		messages.Transcription, messages.ResultQueueFor(messages.Transcription))
}

//transcriptionFinish processes transcription result messages
// 1. logs status
// 2. sends 'ResultMake' message
func transcriptionFinish(d *amqp.Delivery, data *ServiceData) (bool, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "Can't unmarshal message "+string(d.Body))
	}
	c, err := processStatus(&message, data, messages.Transcription, status.ResultMake)
	if !c {
		if err != nil {
			cmdapp.Log.Error(err)
		}
		return true, err
	}
	return true, data.MessageSender.Send(messages.NewQueueMessage(message.ID),
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
	publishStatusChange(&message.QueueMessage, data)

	return true, data.MessageSender.Send(newInformMessage(message.ID, messages.InformType_Finished),
		messages.Inform, "")
}

func processStatus(message *messages.QueueMessage, data *ServiceData, from string, to status.Status) (bool, error) {
	cmdapp.Log.Infof("Got %s msg :%s", from, message.ID)
	if message.Error != "" {
		err := data.StatusSaver.SaveError(message.ID, message.Error)
		if err != nil {
			cmdapp.Log.Error(err)
			return false, err
		}
		publishStatusChange(message, data)
		return false, nil
	}
	err := data.StatusSaver.Save(message.ID, to)
	if err != nil {
		cmdapp.Log.Error(err)
		return false, err
	}
	publishStatusChange(message, data)
	return true, nil
}

func publishStatusChange(message *messages.QueueMessage, data *ServiceData) {
	cmdapp.Log.Infof("Publishing status change %s", message.ID)
	err := data.Publisher.Publish(message.ID, messages.TopicStatusChange)
	cmdapp.LogIf(err)
}

func newInformMessage(id string, it string) *messages.InformMessage {
	return &messages.InformMessage{QueueMessage: messages.QueueMessage{ID: id}, Type: it, At: time.Now().UTC()}
}
