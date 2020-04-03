package cmdworker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/recognizer"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type readFunc func(file string, id string) (string, error)

//RecInfoLoader loads recognizer information
type RecInfoLoader interface {
	Get(key string) (*recognizer.Info, error)
}

//PreloadTaskManager manages long running process, loaded by key before processing task
type PreloadTaskManager interface {
	EnsureRunning(map[string]string) error
}

// ServiceData keeps data required for service work
type ServiceData struct {
	TaskName   string
	Command    string
	WorkingDir string
	//ResultFile if non empty then tries to pass result to reply message from the file
	// changes {ID} in the file with message id
	ResultFile string
	//File to log into the cmd output
	LogFile        string
	ReadFunc       readFunc
	RecInfoLoader  RecInfoLoader
	PreloadManager PreloadTaskManager

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
	if data.ResultFile != "" && data.ReadFunc == nil {
		return nil, errors.New("No command")
	}
	if data.RecInfoLoader == nil {
		return nil, errors.New("No recognizer info loader")
	}
	if data.PreloadManager == nil {
		return nil, errors.New("No Preload manager set")
	}

	fc := make(chan bool)

	go listenQueue(data, fc)
	return fc, nil
}

//work is main method to process of the worker
func work(data *ServiceData, msg *messages.QueueMessage) error {
	cmdapp.Log.Infof("Got task %s for ID: %s, rec: %s", data.TaskName, msg.ID, msg.Recognizer)
	rp, err := data.RecInfoLoader.Get(msg.Recognizer)
	if err != nil {
		return errors.Wrap(err, "Can't load description")
	}
	envs, err := collectEnvParams(rp, msg)
	if err != nil {
		return err
	}
	err = data.PreloadManager.EnsureRunning(rp.Settings)
	if err != nil {
		return errors.Wrap(err, "Can't init preload task")
	}
	logOutput := ioutil.Discard
	if data.LogFile != "" {
		lf := strings.Replace(data.LogFile, "{ID}", msg.ID, -1)
		f, err := os.OpenFile(lf, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			cmdapp.Log.Warn(errors.Wrapf(err, "Can't open file %s", lf))
		} else {
			defer f.Close()
			logOutput = f
		}
	}
	return RunCommand(data.Command, data.WorkingDir, msg.ID, envs, logOutput)
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

func processMsg(d *amqp.Delivery, data *ServiceData) (messages.Message, error) {
	var message messages.QueueMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return nil, errors.Wrap(err, "Can't unmarshal message "+string(d.Body))
	}
	err := work(data, &message)
	cmdapp.Log.Infof("Msg processed")
	result := messages.NewQueueMessageFromM(&message)
	var res string
	if err != nil {
		cmdapp.Log.Error(err)
		result.Error = err.Error()
	} else {
		if data.ResultFile != "" && d.ReplyTo != "" {
			res, err = data.ReadFunc(data.ResultFile, message.ID)
			if err != nil {
				cmdapp.Log.Error(err)
				result.Error = err.Error()
			}
		}
	}
	if data.ResultFile != "" {
		return &messages.ResultMessage{QueueMessage: *result, Result: res}, nil
	}
	return result, nil
}

//ReadFile reads content as string
func ReadFile(file string, id string) (string, error) {
	realFile := strings.Replace(file, "{ID}", id, -1)
	cmdapp.Log.Infof("Reading file: %s", realFile)
	bytes, err := ioutil.ReadFile(realFile)
	if err != nil {
		return "", errors.Wrap(err, "Can't read file "+realFile)
	}
	return string(bytes), nil
}

func collectEnvParams(rp *recognizer.Info, msg *messages.QueueMessage) ([]string, error) {
	var res []string
	for _, t := range msg.Tags {
		res = append(res, fmt.Sprintf("%s=%s", strings.ToUpper(t.Key), t.Value))
	}

	for k, v := range rp.Settings {
		res = append(res, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
	}
	return res, nil
}
