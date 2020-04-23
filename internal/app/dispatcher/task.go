package dispatcher

import (
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
	res.changedFunc = func() {}
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
	if t.msg == nil {
		return errors.New("No msg set")
	}
	if t.d == nil {
		return errors.New("No delivery set")
	}
	if ts.changedFunc == nil {
		return errors.New("No change func")
	}
	ot, found := ts.tsks[t.msg.ID]
	if found {
		cmdapp.Log.Warnf("The same task arrived %s", t.msg.ID)
		if ot.worker != nil {
			cmdapp.Log.Warnf("Hmm. What to do with old task worker. Marking as free %s", ot.worker.queue)
			ot.worker.completeTask()
		}
	}
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
		err := sender.Send(d.Body, t.d.ReplyTo, "")
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
		err := w.completeTask()
		if err != nil {
			cmdapp.Log.Error("Can'not mark worker as completed", err)
		}
		delete(ts.tsks, id)
	} else {
		cmdapp.Log.Error("Task has no worker")
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
	err = w.startTask(t)
	if err != nil {
		return errors.Wrap(err, "Can't mark worker as started")
	}
	t.worker = w
	return nil
}
