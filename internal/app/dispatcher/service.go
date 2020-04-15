package dispatcher

import (
	"encoding/json"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// ServiceData keeps data required for service work
type ServiceData struct {
	fc    *utils.MultiCloseChannel
	wrkrs *workers

	MessageSender  messages.Sender
	RegistrationCh <-chan amqp.Delivery
}

//StartWorkerService starts the event queue listener service to listen for manager and work events
func StartWorkerService(data *ServiceData) error {
	cmdapp.Log.Infof("Starting listen for messages")
	if data.RegistrationCh == nil {
		return errors.New("No Registration channel")
	}
	go listenRegistrationQueue(data)
	go checkForExpiredWorkers(data.wrkrs)
	return nil
}

func listenRegistrationQueue(data *ServiceData) {
	for d := range data.RegistrationCh {
		err := processRegistrationMsg(&d, data)
		if err != nil {
			cmdapp.Log.Error("Message error", err)
		}
		err = d.Ack(false)
		if err != nil {
			cmdapp.Log.Error("Ack error", err)
		}
	}
	cmdapp.Log.Infof("Stopped listening registration queue")
	data.fc.Close()
}

func processRegistrationMsg(d *amqp.Delivery, data *ServiceData) error {
	var message messages.RegistrationMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return errors.Wrap(err, "Can't unmarshal message "+string(d.Body))
	}
	return processWorker(data.wrkrs, &message)
}

func changed() {

}
