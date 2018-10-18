package inform

import (
	"encoding/json"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/inform"

	"github.com/jordan-wright/email"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

//Sender send emails
type Sender interface {
	Send(email *email.Email) error
}

//EmailMaker prepares the email
type EmailMaker interface {
	Make(data *inform.Data) (*email.Email, error)
}

//EmailRetriever return the email by ID
type EmailRetriever interface {
	Get(ID string) (string, error)
}

//Locker tracks email sending process
//It is used to quarantee not to send the emails twice
type Locker interface {
	Lock(id string, lockKey string) error
	UnLock(id string, lockKey string, value *int) error
}

// ServiceData keeps data required for service work
type ServiceData struct {
	taskName       string
	workCh         <-chan amqp.Delivery
	emailSender    Sender
	emailMaker     EmailMaker
	emailRetriever EmailRetriever
	locker         Locker
	location       *time.Location
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
	if data.taskName == "" {
		return nil, errors.New("No Task Name")
	}
	if data.emailMaker == nil {
		return nil, errors.New("No email maker")
	}
	if data.emailRetriever == nil {
		return nil, errors.New("No email retriever")
	}
	if data.emailSender == nil {
		return nil, errors.New("No sender")
	}
	if data.locker == nil {
		return nil, errors.New("No locker")
	}
	if data.workCh == nil {
		return nil, errors.New("No work channel")
	}

	fc := make(chan bool)

	go listenQueue(data, fc)
	return fc, nil
}

//work is main method to send the message
func work(data *ServiceData, message *messages.InformMessage) error {
	cmdapp.Log.Infof("Got task %s for ID: %s", data.taskName, message.ID)

	mailData := inform.Data{}
	mailData.ID = message.ID
	mailData.MsgTime = toLocalTime(data, message.At)
	mailData.MsgType = message.Type

	var err error
	mailData.Email, err = data.emailRetriever.Get(message.ID)
	if err != nil {
		cmdapp.Log.Error(err)
		return errors.Wrap(err, "Can't retrieve email")
	}

	email, err := data.emailMaker.Make(&mailData)
	if err != nil {
		cmdapp.Log.Error(err)
		return errors.Wrap(err, "Can't prepare email")
	}

	err = data.locker.Lock(mailData.ID, mailData.MsgType)
	if err != nil {
		cmdapp.Log.Error(err)
		return errors.Wrap(err, "Can't lock mail table")
	}
	var unlockValue = 0
	defer data.locker.UnLock(mailData.ID, mailData.MsgType, &unlockValue)

	err = data.emailSender.Send(email)
	if err != nil {
		cmdapp.Log.Error(err)
		return errors.Wrap(err, "Can't send email")
	}
	unlockValue = 2
	return nil
}

func listenQueue(data *ServiceData, fc chan<- bool) {
	for d := range data.workCh {
		redeliver, err := processMsg(&d, data)
		if err != nil {
			cmdapp.Log.Error("Message error", err)
			d.Nack(false, redeliver && !d.Redelivered) // try redeliver for the first time
			continue
		}
		d.Ack(false)
	}
	cmdapp.Log.Infof("Stopped listening queue")
	fc <- true
}

func toLocalTime(data *ServiceData, t time.Time) time.Time {
	if data.location != nil {
		return t.In(data.location)
	}
	return t
}

//processMsg returns true if it needs to retry on error again
func processMsg(d *amqp.Delivery, data *ServiceData) (bool, error) {
	var message messages.InformMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return false, errors.Wrap(err, "Can't unmarshal message "+string(d.Body))
	}
	err := work(data, &message)
	cmdapp.Log.Infof("Msg processed")
	return true, err
}
