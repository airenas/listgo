package cmdworker

import (
	"encoding/json"
	"os/exec"
	"strings"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// ServiceData keeps data required for service work
type ServiceData struct {
	TaskName   string
	Command    string
	WorkingDir string

	MessageSender messages.Sender
	WorkCh        <-chan amqp.Delivery
}

//StartWorkerService starts the event queue listener service to listen for configured events
// return channel to track the finish event
//
// to wait sync for the service to finish:
// fc, err := StartWorkerService(data)
// handle err
// <-fc // waits for finish
func StartWorkerService(data *ServiceData) (<-chan bool, error) {
	cmdapp.Log.Infof("Starting listen for messages")
	if data.TaskName == "" {
		return nil, errors.New("No Task Name")
	}
	if data.Command == "" {
		return nil, errors.New("No command")
	}

	fc := make(chan bool)

	go listenQueue(data, fc)
	return fc, nil
}

//work is main method to process of the worker
func work(data *ServiceData, id string) error {
	cmdapp.Log.Infof("Got task %s for ID: %s", data.TaskName, id)
	err := RunCommand(data.Command, data.WorkingDir, id)
	if err != nil {
		cmdapp.Log.Error(err)
		return err
	}
	return nil
}

func listenQueue(data *ServiceData, fc chan<- bool) {
	for d := range data.WorkCh {
		msg, err := processMsg(&d, data)
		if err != nil {
			cmdapp.Log.Error("Message error", err)
			d.Nack(false, false)
			continue
		}
		if d.ReplyTo != "" {
			err = data.MessageSender.Send(msg, d.ReplyTo, "")
			if err != nil {
				cmdapp.Log.Error("Can't reply result", err)
				d.Nack(false, !d.Redelivered) // try redeliver for first time
				continue
			}
		}
		d.Ack(false)
	}
	cmdapp.Log.Infof("Stopped listening queue")
	fc <- true
}

func processMsg(d *amqp.Delivery, data *ServiceData) (*messages.QueueMessage, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return nil, errors.Wrap(err, "Can't unmarshal message "+string(d.Body))
	}
	err := work(data, message.ID)
	cmdapp.Log.Infof("Msg processed")
	result := messages.NewQueueMessage(message.ID)
	if err != nil {
		result.Error = err.Error()
	}
	return result, nil
}

//RunCommand executes system comman end return error if any
func RunCommand(command string, workingDir string, id string) error {
	realCommand := strings.Replace(command, "{ID}", id, -1)
	cmdapp.Log.Infof("Running command: %s", realCommand)
	cmdapp.Log.Infof("Working Dir: %s", workingDir)
	cmdArr := strings.Split(realCommand, " ")
	if len(cmdArr) < 2 {
		return errors.New("Wrong command. No parameter " + realCommand)
	}

	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	cmd.Dir = workingDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		errR := errors.Wrap(err, "Output: "+string(output))
		cmdapp.Log.Error(errR.Error())
		return errR
	}
	return nil
}
