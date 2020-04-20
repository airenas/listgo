package dispatcher

import (
	"encoding/json"
	"sync"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

//SelectionStrategy type returns task work worker
type SelectionStrategy interface {
	FindBest(wrks []*worker, tsks map[string]*task, wi int) (*task, error)
}

//DurationGetter type provider duration for transcription ID
type DurationGetter interface {
	Get(v string) (time.Duration, error)
}

//ModelTypeGetter type provides requires preload model type for transcription ID
type ModelTypeGetter interface {
	Get(v string) (string, error)
}

//StartTimeGetter type provides start time for the transcription
type StartTimeGetter interface {
	Get(tags []messages.Tag) (time.Time, error)
}

// ServiceData keeps data required for service work
type ServiceData struct {
	fc    *utils.MultiCloseChannel
	wrkrs *workers
	tsks  *tasks

	selectionStrategy SelectionStrategy
	modelLoadDuration time.Duration
	rtFactor          float32

	startTimeGetter StartTimeGetter
	modelTypeGetter ModelTypeGetter
	durationGetter  DurationGetter

	replySender messages.Sender
	workSender  messages.Sender

	RegistrationCh <-chan amqp.Delivery
	WorkCh         <-chan amqp.Delivery
	ResponseCh     <-chan amqp.Delivery
}

//StartWorkerService starts the event queue listener service to listen for manager and work events
func StartWorkerService(data *ServiceData) error {
	cmdapp.Log.Infof("Starting listen for messages")
	err := validate(data)
	if err != nil {
		return err
	}

	data.tsks.changedFunc = func() { changed(data) }
	data.wrkrs.changedFunc = func() { changed(data) }

	go listenRegistrationQueue(data)
	go checkForExpiredWorkers(data.wrkrs)

	go listenWorkQueue(data)
	go listenResponseQueue(data)

	return nil
}

func validate(data *ServiceData) error {
	if data.RegistrationCh == nil {
		return errors.New("No Registration channel")
	}
	if data.WorkCh == nil {
		return errors.New("No Work channel")
	}
	if data.ResponseCh == nil {
		return errors.New("No Response channel")
	}
	if data.replySender == nil {
		return errors.New("No reply sender")
	}
	if data.workSender == nil {
		return errors.New("No work sender")
	}
	if data.durationGetter == nil {
		return errors.New("No duration getter")
	}
	if data.modelTypeGetter == nil {
		return errors.New("No model type getter")
	}
	if data.selectionStrategy == nil {
		return errors.New("No selection strategy")
	}
	if data.startTimeGetter == nil {
		return errors.New("No start time getter")
	}
	return nil
}

func listenRegistrationQueue(data *ServiceData) {
	for d := range data.RegistrationCh {
		err := processRegistrationMsg(data, &d)
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

func listenWorkQueue(data *ServiceData) {
	for d := range data.WorkCh {
		err := processWorkMsg(data, d)
		if err != nil {
			cmdapp.Log.Error("Message error", err)
			d.Nack(false, false)
		}
	}
	cmdapp.Log.Infof("Stopped listening work queue")
	data.fc.Close()
}

func listenResponseQueue(data *ServiceData) {
	for d := range data.ResponseCh {
		err := data.tsks.processResponse(&d, data.replySender)
		if err != nil {
			cmdapp.Log.Error("Message error", err)
		}
	}
	cmdapp.Log.Infof("Stopped listening response queue")
	data.fc.Close()
}

func processRegistrationMsg(data *ServiceData, d *amqp.Delivery) error {
	var message messages.RegistrationMessage
	if err := json.Unmarshal(d.Body, &message); err != nil {
		return errors.Wrap(err, "Can't unmarshal message "+string(d.Body))
	}
	return processWorker(data.wrkrs, &message)
}

func processWorkMsg(data *ServiceData, d amqp.Delivery) error {
	var msg messages.QueueMessage
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		return errors.Wrap(err, "Can't unmarshal message "+string(d.Body))
	}
	return addTask(data, &d, &msg)
}

func addTask(data *ServiceData, d *amqp.Delivery, msg *messages.QueueMessage) error {
	cmdapp.Log.Infof("Got task %s", msg.ID)
	t := newTask()
	t.d = d
	t.msg = msg
	var err error
	t.addedAt, err = data.startTimeGetter.Get(msg.Tags)
	if err != nil {
		cmdapp.Log.Error("Can't get startTime. ", err)
	}
	t.expModelLoadDuration = data.modelLoadDuration
	t.rtFactor = data.rtFactor
	t.expDuration, err = data.durationGetter.Get(msg.ID)
	if err != nil {
		cmdapp.Log.Error("Can't get duration. ", err)
	}
	t.requiredModelType, err = data.modelTypeGetter.Get(msg.Recognizer)
	if err != nil {
		cmdapp.Log.Error("Can't get model type. ", err)
	}
	return data.tsks.addTask(t)
}

var changedStartup sync.Once

// the main task deliver procedure
func changed(data *ServiceData) {
	//allow to read tasks and workers on service startup
	changedStartup.Do(func() {
		cmdapp.Log.Info("Do wait for the first time")
		time.Sleep(3 * time.Second)
	})
	data.wrkrs.lock.Lock()
	defer data.wrkrs.lock.Unlock()

	wrks := make([]*worker, 0)
	for _, k := range data.wrkrs.workers {
		wrks = append(wrks, k)
	}
	cmdapp.Log.Infof("Workers: %d, tasks: %d", len(wrks), len(data.tsks.tsks))
	for i, w := range wrks {
		if w.working == false {
			t, err := data.selectionStrategy.FindBest(wrks, data.tsks.tsks, i)
			if err != nil {
				cmdapp.Log.Error("Can't get task", err)
			}
			if t != nil {
				err = t.startOn(w, data.workSender)
				if err != nil {
					cmdapp.Log.Error("Can't start task", err)
				}
			}
		}
	}
}
