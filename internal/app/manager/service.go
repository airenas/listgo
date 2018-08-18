package manager

import (
	"encoding/json"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"

	"bitbucket.org/airenas/listgo/internal/app/upload"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// ServiceData keeps data required for service work
type ServiceData struct {
	MessageSender  messages.Sender
	StatusSaver    upload.StatusSaver
	DecodeCh       <-chan amqp.Delivery
	AudioConvertCh <-chan amqp.Delivery
}

type prFunc func(message *messages.QueueMessage, data *ServiceData, d *amqp.Delivery) error

//StartWorkerService starts the event queue listener service to listen for events
func StartWorkerService(data *ServiceData) error {
	cmdapp.Log.Infof("Starting listen for messages")

	fc := make(chan bool)

	go listenQueue(data.DecodeCh, decode, data, fc)
	go listenQueue(data.AudioConvertCh, audioConvertFinish, data, fc)

	<-fc
	cmdapp.Log.Infof("Exiting service")
	return nil
}

func listenQueue(q <-chan amqp.Delivery, f prFunc, data *ServiceData, fc chan<- bool) {
	for d := range q {
		redeliver, err := processMsg(&d, f, data)
		if err != nil {
			cmdapp.Log.Errorf("Can't process message %s\n%s", d.MessageId, string(d.Body))
			cmdapp.Log.Error(err)
			d.Nack(false, redeliver && !d.Redelivered) // redeliver for first time
		} else {
			d.Ack(false)
		}
	}
	cmdapp.Log.Infof("Stopped listening queue")
	fc <- true
}

//processMsg return true if message can be retried
func processMsg(d *amqp.Delivery, f prFunc, data *ServiceData) (bool, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "Can't unmarshal message "+string(d.Body))
	}
	return true, f(&message, data, d)
}

//decode starts the transcription process
// workflow:
// 1. set status to STARTED
// 2. send 'Started' event (async)
// 3. send and wait for 'AudioConvert' to finish
func decode(message *messages.QueueMessage, data *ServiceData, d *amqp.Delivery) error {
	cmdapp.Log.Infof("Got decode msg :%s", message.ID)
	err := data.StatusSaver.Save(message.ID, messages.AudioConvert, "")
	if err != nil {
		cmdapp.Log.Error(err)
		return err
	}
	err = data.MessageSender.Send(messages.NewQueueMessage(message.ID),
		messages.StartedDecode, "")
	if err != nil {
		return err
	}
	return data.MessageSender.Send(messages.NewQueueMessage(message.ID),
		messages.AudioConvert, messages.ResultQueueFor(messages.AudioConvert))
}

//audioConvertFinish processes audio convert result messages
// 1. logs status
// 2. sends 'Diarization' message
func audioConvertFinish(message *messages.QueueMessage, data *ServiceData, d *amqp.Delivery) error {
	cmdapp.Log.Infof("Got audioConvertFinish msg :%s", message.ID)
	if message.Error != "" {
		err := data.StatusSaver.Save(message.ID, messages.AudioConvert, message.Error)
		if err != nil {
			cmdapp.Log.Error(err)
			return err
		}
		return nil
	}
	err := data.StatusSaver.Save(message.ID, messages.Diarization, "")
	if err != nil {
		cmdapp.Log.Error(err)
		return err
	}
	return data.MessageSender.Send(messages.NewQueueMessage(message.ID),
		messages.Diarization, messages.ResultQueueFor(messages.Diarization))
}

//decode is main method to lead the transcription process
// workflow:
// 1. set status to STARTED
// 2. send 'Started' event (async)
// 3. send and wait for 'AudioConvert' to finish
// 4. send and wait for 'Diarization' to finish
// 5. send and wait for 'Transcription' to finish
// 6. send and wait for 'MakeResult' to finish

// send 'Finished' event (async)
// set status to COMPLETE
// func decode1(data *ServiceData, id string) error {
// 	cmdapp.Log.Infof("Got decode msg :%s", id)
// 	err := data.StatusSaver.Save(id, "STARTED", "")
// 	if err != nil {
// 		cmdapp.Log.Error(err)
// 		return err
// 	}
// 	sendAsyncMessage("Started", data, id)

// 	err = doTask("AudioConvert", data, id)
// 	if err != nil {
// 		cmdapp.Log.Error(err)
// 		return err
// 	}
// 	err = doTask("Diarization", data, id)
// 	if err != nil {
// 		cmdapp.Log.Error(err)
// 		return err
// 	}
// 	err = doTask("Transcription", data, id)
// 	if err != nil {
// 		cmdapp.Log.Error(err)
// 		return err
// 	}
// 	err = doTask("MakeResult", data, id)
// 	if err != nil {
// 		cmdapp.Log.Error(err)
// 		return err
// 	}

// 	sendAsyncMessage("Finished", data, id)

// 	err = data.StatusSaver.Save(id, "COMPLETE", "")
// 	if err != nil {
// 		cmdapp.Log.Error(err)
// 		return err
// 	}
// 	return nil
// }

// func doTask(name string, data *ServiceData, id string) error {
// 	cmdapp.Log.Infof("Doing task %s, id %s", name, id)
// 	err := data.StatusSaver.Save(id, name, "")
// 	if err != nil {
// 		return errors.Wrap(err, "Can't save status")
// 	}
// 	asyncResult, err := data.MessageSender.Send(newMsg(id, name))
// 	if err != nil {
// 		data.StatusSaver.Save(id, name, err.Error())
// 		return errors.Wrap(err, "Can't send message")
// 	}
// 	cmdapp.Log.Infof("Message sent. Waiting for completion %s, id %s", name, id)
// 	_, err = asyncResult.Get(time.Duration(time.Second))
// 	if err != nil {
// 		data.StatusSaver.Save(id, name, err.Error())
// 		return errors.Wrap(err, "Can't get result")
// 	}
// 	return nil
// }

// func sendMessage(name string, data *ServiceData, id string) {
// 	// _, err := data.MessageSender.Send(newMsg(id, name))
// 	// if err != nil {
// 	// 	cmdapp.Log.Errorf("Can't send message %s (%s). Cause: %s", name, id, err.Error())
// 	// }
// }

// func newMsg(id string, queue string) *msgsender.Message {
// 	return &msgsender.Message{ID: id, Queue: queue}
// }
