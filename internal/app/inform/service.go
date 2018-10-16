package inform

import (
	"encoding/json"
	"time"

	"github.com/jordan-wright/email"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type readFunc func(file string, id string) (string, error)

//Sender send emails
type Sender interface {
	Send(email *email.Email) error
}

//EmailMaker prepares the email
type EmailMaker interface {
	Make(data *Data) (*email.Email, error)
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

//Data keeps data for email generation
type Data struct {
	id      string
	msgType string
	email   string
	msgTime time.Time
}

// ServiceData keeps data required for service work
type ServiceData struct {
	TaskName       string
	WorkCh         <-chan amqp.Delivery
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
	if data.TaskName == "" {
		return nil, errors.New("No Task Name")
	}

	fc := make(chan bool)

	go listenQueue(data, fc)
	return fc, nil
}

//work is main method to send the message
func work(data *ServiceData, message *messages.InformMessage) error {
	cmdapp.Log.Infof("Got task %s for ID: %s", data.TaskName, message.ID)

	mailData := Data{}
	mailData.id = message.ID
	mailData.msgTime = toLocalTime(data, message.At)
	mailData.msgType = message.Type

	var err error
	mailData.email, err = data.emailRetriever.Get(message.ID)
	if err != nil {
		cmdapp.Log.Error(err)
		return errors.Wrap(err, "Can't retrieve email")
	}

	email, err := data.emailMaker.Make(&mailData)
	if err != nil {
		cmdapp.Log.Error(err)
		return errors.Wrap(err, "Can't prepare email")
	}

	err = data.locker.Lock(mailData.id, mailData.msgType)
	if err != nil {
		cmdapp.Log.Error(err)
		return errors.Wrap(err, "Can't lock mail table")
	}
	var unlockValue = 0
	defer data.locker.UnLock(mailData.id, mailData.msgType, &unlockValue)

	err = data.emailSender.Send(email)
	if err != nil {
		cmdapp.Log.Error(err)
		return errors.Wrap(err, "Can't send email")
	}
	unlockValue = 2
	return nil
}

func listenQueue(data *ServiceData, fc chan<- bool) {
	for d := range data.WorkCh {
		err := processMsg(&d, data)
		if err != nil {
			cmdapp.Log.Error("Message error", err)
			d.Nack(false, false)
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

func processMsg(d *amqp.Delivery, data *ServiceData) error {
	var message messages.InformMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return errors.Wrap(err, "Can't unmarshal message "+string(d.Body))
	}
	err := work(data, &message)
	cmdapp.Log.Infof("Msg processed")
	return err
}
