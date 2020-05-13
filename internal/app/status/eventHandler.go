package status

import (
	"time"

	"bitbucket.org/airenas/listgo/internal/app/status/api"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type eventChannelFunc func() (<-chan amqp.Delivery, error)

func listenQueue(channel <-chan amqp.Delivery, data *ServiceData, fc chan<- bool) {
	for d := range channel {
		err := processMsg(&d, data)
		if err != nil {
			cmdapp.Log.Errorf("Can't process message %s\n%s", d.MessageId, string(d.Body))
			cmdapp.Log.Error(err)
		}
	}
	cmdapp.Log.Infof("Stopped listening queue")
	close(fc)
}

func registerQueue(data *ServiceData, quitChan <-chan bool, initialWait time.Duration) {
	wait := initialWait
	for {
		select {
		case <-quitChan:
			cmdapp.Log.Infof("Quit listening queue")
			return
		default:
			fc := make(chan bool)
			cmdapp.Log.Infof("Trying listening queue")
			msgs, err := data.EventChannelFunc()
			if err != nil {
				cmdapp.Log.Error(err)
				wait = wait * 2
				if wait > time.Minute {
					wait = time.Minute
				}
				cmdapp.Log.Infof("Wait before reconnect %d s", wait/time.Second)
				time.Sleep(wait)
				continue
			}
			wait = initialWait
			go listenQueue(msgs, data, fc)
			<-fc
		}
	}
}

func processMsg(d *amqp.Delivery, data *ServiceData) error {
	id := string(d.Body)
	cmdapp.Log.Infof("processMsg event " + id)
	conns, found := getConnections(id)
	if found {
		result, err := data.StatusProvider.Get(id)
		if err != nil {
			return errors.Wrap(err, "Cannot get status for ID: "+id)
		}
		for c := range conns {
			err = sendMsg(c, result)
			cmdapp.LogIf(err)
		}
	} else {
		cmdapp.Log.Infof("No connections found for " + id)
	}
	return nil
}

func sendMsg(c WsConn, result *api.TranscriptionResult) error {
	cmdapp.Log.Debugf("Sending result for %s to websockket", result.ID)
	err := c.WriteJSON(result)
	if err != nil {
		return errors.Wrap(err, "Cannot write to websockket")
	}
	cmdapp.Log.Debug("Sent msg to websockket")
	return nil
}
