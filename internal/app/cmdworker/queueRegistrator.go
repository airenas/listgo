package cmdworker

import (
	"time"

	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/airenas/listgo/internal/pkg/messages"
	"github.com/airenas/listgo/internal/pkg/utils"
	"github.com/pkg/errors"
)

type queueRegistrator struct {
	sender        messages.Sender
	registryQueue string
	ownQueue      string
	heartbeatInt  time.Duration
	count         int
	failureCount  int
	close         bool
	closeChan     *utils.MultiCloseChannel
}

func newQueueRegistrator(sender messages.Sender, qName string, closeChan *utils.MultiCloseChannel) (*queueRegistrator, error) {
	res := &queueRegistrator{}
	if sender == nil {
		return nil, errors.New("No sender")
	}
	res.sender = sender
	if qName == "" {
		return nil, errors.New("No own queue name")
	}
	res.ownQueue = qName
	if closeChan == nil {
		return nil, errors.New("No close channel")
	}
	res.closeChan = closeChan

	res.registryQueue = cmdapp.Config.GetString("registry.queue")
	if res.registryQueue == "" {
		return nil, errors.New("No registry.queue config")
	}
	res.heartbeatInt = cmdapp.Config.GetDuration("registry.heartbeat")
	if res.heartbeatInt < time.Second {
		return nil, errors.New("No or very fast registry.heartbeat in config")
	}
	return res, nil
}

func (qr *queueRegistrator) live() {
	for {
		if qr.close {
			cmdapp.Log.Info("Registrator is going to die. Exit live function")
			return
		}
		err := qr.heartbeat()
		qr.count++
		if err != nil {
			cmdapp.Log.Error(err, "Can't send heartbeat")
			qr.failureCount++
			if qr.failureCount > 5 {
				cmdapp.Log.Info("Failure for the 5th time. No point to live. Indicating to stop the app")
				qr.closeChan.Close()
			}
		} else {
			qr.failureCount = 0
		}
		time.Sleep(qr.heartbeatInt)
	}
}

func (qr *queueRegistrator) heartbeat() error {
	if qr.count == 0 {
		return qr.sendMsg(messages.RgrTypeRegister)
	}
	return qr.sendMsg(messages.RgrTypeBeat)
}

func (qr *queueRegistrator) sendMsg(mt string) error {
	cmdapp.Log.Debugf("Sending msg %s to %s", mt, qr.registryQueue)
	msg := messages.RegistrationMessage{}
	msg.Queue = qr.ownQueue
	msg.Type = mt
	msg.Timestamp = time.Now().Unix()
	return qr.sender.Send(msg, qr.registryQueue, "")
}

func (qr *queueRegistrator) Close() error {
	qr.close = true
	return qr.sendMsg(messages.RgrTypeExit)
}
