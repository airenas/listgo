package cmdworker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/recognizer"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"
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
	Close() error
}

// ServiceData keeps data required for service work
type ServiceData struct {
	Name       string
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

	MessageSender messages.SenderWithCorr
	WorkCh        <-chan amqp.Delivery
	reapLock      *sync.RWMutex

	skipAck     bool
	quitChannel *utils.MultiCloseChannel
}

//StartWorkerService starts the event queue listener service to listen for configured events
// return channel to track the finish event
//
// to wait sync for the service to finish:
// fc, err := StartWorkerService(data)
// handle err
// <-fc // waits for finish
func StartWorkerService(data *ServiceData) error {
	cmdapp.Log.Infof("Starting listen for messages")
	if data.Name == "" {
		return errors.New("No Name")
	}
	if data.Command == "" {
		return errors.New("No command")
	}
	if data.ResultFile != "" && data.ReadFunc == nil {
		return errors.New("No command")
	}
	if data.RecInfoLoader == nil {
		return errors.New("No recognizer info loader")
	}
	if data.PreloadManager == nil {
		return errors.New("No Preload manager set")
	}

	go listenQueue(data)
	return nil
}

//work is main method to process of the worker
func work(data *ServiceData, msg *messages.QueueMessage) error {
	data.reapLock.Lock()
	defer data.reapLock.Unlock()

	cmdapp.Log.Infof("Got task %s for ID: %s, rec: %s", data.Name, msg.ID, msg.Recognizer)
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

func listenQueue(data *ServiceData) {
	for d := range data.WorkCh {
		msg, err := processMsg(&d, data)
		if err != nil {
			cmdapp.Log.Error("Message error", err)
			if !data.skipAck {
				d.Nack(false, false)
			}
			continue
		}
		if d.ReplyTo != "" {
			err = data.MessageSender.SendWithCorr(msg, d.ReplyTo, "", d.CorrelationId)
			if err != nil {
				cmdapp.Log.Error("Can't reply result", err)
				if !data.skipAck {
					d.Nack(false, !d.Redelivered) // try redeliver for first time
				}
				continue
			}
			cmdapp.Log.Infof("Sent reply message to %s, corrID: %s", d.ReplyTo, d.CorrelationId)
		}
		if !data.skipAck {
			d.Ack(false)
		}
	}
	cmdapp.Log.Infof("Stopped listening queue")
	data.quitChannel.Close()
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
