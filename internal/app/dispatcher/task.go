package dispatcher

import (
	"encoding/json"
	"sync"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type task struct {
	d   *amqp.Delivery
	msg *messages.QueueMessage

	requiredModelType    string
	expDuration          time.Duration
	expModelLoadDuration time.Duration
	addedAt              time.Time
	rtFactor             float32

	worker    *worker
	started   bool
	failCount int32
	startedAt time.Time
}

type tasks struct {
	tsks map[string]*task
	lock *sync.Mutex

	changedFunc changedFunc
}

func newTask() *task {
	res := &task{}
	return res
}

func newTasks() *tasks {
	res := &tasks{}
	res.lock = &sync.Mutex{}
	res.tsks = make(map[string]*task)
	return res
}

func failRequeueTask(t *task) {
	t.worker = nil
	t.started = false
	t.failCount++
}

func (ts *tasks) addTask(t *task) error {
	ts.lock.Lock()
	defer ts.lock.Unlock()
	ts.tsks[t.msg.ID] = t
	go ts.changedFunc()
	return nil
}

func (ts *tasks) processResponse(d *amqp.Delivery, sender messages.Sender) error {
	ts.lock.Lock()
	defer ts.lock.Unlock()
	id := d.CorrelationId
	cmdapp.Log.Infof("Process response message %s", id)
	t, f := ts.tsks[id]
	if !f {
		return errors.Errorf("Hmm, correlation ID '%s' not found in task list, old task arrived?", id)
	}

	acked := false
	if t.d.ReplyTo != "" {
		var msg messages.QueueMessage
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			cmdapp.Log.Error(errors.Wrap(err, "Can't unmarshal message "+string(d.Body)))
			err := t.d.Nack(false, !t.d.Redelivered) // try redeliver for first time
			if err != nil {
				cmdapp.Log.Error(err, "Can't nack")
			}
			acked = true
		}
		err := sender.Send(msg, t.d.ReplyTo, "")
		if err != nil {
			cmdapp.Log.Error("Can't reply result", err)
			err := t.d.Nack(false, !t.d.Redelivered) // try redeliver for first time
			if err != nil {
				cmdapp.Log.Error(err, "Can't nack")
			}
			acked = true
		}
		cmdapp.Log.Infof("Sent response to %s, corrID: %s", t.d.ReplyTo, id)
	}
	if !acked {
		err := t.d.Ack(false)
		if err != nil {
			cmdapp.Log.Error(err, "Can't ack")
		}
	}

	w := t.worker
	if w != nil {
		w.completeTask()
		delete(ts.tsks, id)
	}
	go ts.changedFunc()
	return nil
}

func (t *task) startOn(w *worker, sender messages.Sender) error {
	cmdapp.Log.Infof("Delivering task(%s) %s to %s", t.requiredModelType, t.msg.ID, w.queue)
	err := sender.Send(t.msg, w.queue, t.msg.ID)
	if err != nil {
		return errors.Wrap(err, "Can't send msg")
	}
	w.startTask(t)
	t.worker = w
	return nil
}
