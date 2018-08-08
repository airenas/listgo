package manager

import (
	"time"

	"bitbucket.org/airenas/listgo/internal/app/upload"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/msgsender"
	"github.com/pkg/errors"
)

// ServiceData keeps data required for service work
type ServiceData struct {
	MessageSender   MessageSender
	MessageListener MessageListener
	StatusSaver     upload.StatusSaver
}

//StartWorkerService starts the event queue listener service tp listen for Decode events
func StartWorkerService(data *ServiceData) error {
	cmdapp.Log.Infof("Starting listen for messages")

	err := data.MessageListener.RegisterTask("Decode", func(id string) error {
		go decode(data, id)
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "Can't register decode task")
	}

	err = data.MessageListener.Listen("decode_manager")

	if err != nil {
		return errors.Wrap(err, "Can't listen for message queue")
	}
	return nil
}

//decode is main method to lead the transcription process
// workflow:
// 1. set status to STARTED
// 2. send 'Started' event (async)
// 3. send and wait for 'AudioConvert' to finish
// 4. send and wait for 'Diarization' to finish

// send 'Finished' event (async)
// set status to COMPLETE
func decode(data *ServiceData, id string) error {
	cmdapp.Log.Infof("Got decode msg :%s", id)
	err := data.StatusSaver.Save(id, "STARTED", "")
	if err != nil {
		cmdapp.Log.Error(err)
		return err
	}
	sendAsyncMessage("Started", data, id)

	err = doTask("AudioConvert", data, id)
	if err != nil {
		cmdapp.Log.Error(err)
		return err
	}
	err = doTask("Diarization", data, id)
	if err != nil {
		cmdapp.Log.Error(err)
		return err
	}
	// err = doTask("Decode", data, id)
	// if err != nil {
	// 	cmdapp.Log.Error(err)
	// 	return err
	// }

	sendAsyncMessage("Finished", data, id)

	err = data.StatusSaver.Save(id, "COMPLETE", "")
	if err != nil {
		cmdapp.Log.Error(err)
		return err
	}
	return nil
}

func doTask(name string, data *ServiceData, id string) error {
	cmdapp.Log.Infof("Doing task %s, id %s", name, id)
	err := data.StatusSaver.Save(id, name, "")
	if err != nil {
		return errors.Wrap(err, "Can't save status")
	}
	asyncResult, err := data.MessageSender.Send(newMsg(id, name))
	if err != nil {
		data.StatusSaver.Save(id, name, err.Error())
		return errors.Wrap(err, "Can't send message")
	}
	cmdapp.Log.Infof("Message sent. Waiting for completion %s, id %s", name, id)
	_, err = asyncResult.Get(time.Duration(time.Second))
	if err != nil {
		data.StatusSaver.Save(id, name, err.Error())
		return errors.Wrap(err, "Can't get result")
	}
	return nil
}

func sendAsyncMessage(name string, data *ServiceData, id string) {
	_, err := data.MessageSender.Send(newMsg(id, name))
	if err != nil {
		cmdapp.Log.Errorf("Can't send message %s (%s). Cause: %s", name, id, err.Error())
	}
}

func newMsg(id string, queue string) *msgsender.Message {
	return &msgsender.Message{ID: id, Queue: queue}
}
