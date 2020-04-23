package dispatcher

import (
	"sync"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"github.com/pkg/errors"
)

const (
	timeFormat          = "15:04:05"
	noneWorkerModelType = "<none>"
)

type worker struct {
	queue    string
	beatTime time.Time
	working  bool

	task    *task
	started time.Time
	mType   string
	endAt   time.Time
}

type changedFunc func()

type workers struct {
	workers     map[string]*worker
	changedFunc changedFunc

	lock *sync.Mutex
}

func newWorkers() *workers {
	res := &workers{}
	res.lock = &sync.Mutex{}
	res.workers = make(map[string]*worker)
	res.changedFunc = func() {}
	return res
}

func newWorker() *worker {
	res := &worker{}
	res.mType = noneWorkerModelType
	return res
}

func processWorker(wrks *workers, msg *messages.RegistrationMessage) error {
	if expired(time.Unix(msg.Timestamp, 0), time.Now()) {
		return nil
	}
	if msg.Type == messages.RgrTypeExit {
		return exitWorker(wrks, msg)
	}
	if msg.Type == messages.RgrTypeRegister {
		return registerWorker(wrks, msg)
	}
	if msg.Type == messages.RgrTypeBeat {
		return beatWorker(wrks, msg)
	}
	return errors.Errorf("Unknown msg type: '%s'", msg.Type)
}

func (wrks *workers) log() {
	for _, k := range wrks.workers {
		cmdapp.Log.Debugf("Worker: %s, mt: %s, working: %v, started: %s, endsAt: %s",
			k.queue, k.mType, k.working, k.started.Format(timeFormat), k.endAt.Format(timeFormat))
	}
}

func expired(t time.Time, now time.Time) bool {
	return t.Add(120 * time.Second).Before(now)
}

func exitWorker(wrks *workers, msg *messages.RegistrationMessage) error {
	wrks.lock.Lock()
	defer wrks.lock.Unlock()

	w, f := wrks.workers[msg.Queue]
	if f {
		cmdapp.Log.Infof("Exit worker %s", w.queue)
		delete(wrks.workers, msg.Queue)
		if w.task != nil {
			failRequeueTask(w.task)
		}
		go wrks.changedFunc()
	}
	return nil
}

func registerWorker(wrks *workers, msg *messages.RegistrationMessage) error {
	wrks.lock.Lock()
	defer wrks.lock.Unlock()

	w, f := wrks.workers[msg.Queue]
	if !f {
		w = newWorker()
		w.queue = msg.Queue
		//w.working = msg.Working
		wrks.workers[w.queue] = w
		cmdapp.Log.Infof("Registered worker %s", w.queue)
		go wrks.changedFunc()
	}
	w.beatTime = time.Unix(msg.Timestamp, 0)
	cmdapp.Log.Debugf("Worker count: %d", len(wrks.workers))
	return nil
}

func dropWorker(wrks *workers, w *worker) {
	cmdapp.Log.Infof("Drop worker %s", w.queue)
	delete(wrks.workers, w.queue)
	cmdapp.Log.Debugf("Worker count: %d", len(wrks.workers))
	if w.task != nil {
		failRequeueTask(w.task)
	}
	go wrks.changedFunc()
}

func beatWorker(wrks *workers, msg *messages.RegistrationMessage) error {
	cmdapp.Log.Debugf("Got heartbeat from %s", msg.Queue)
	return registerWorker(wrks, msg)
}

func checkForExpiredWorkers(wrks *workers) {
	for {
		time.Sleep(30 * time.Second)
		cmdapp.Log.Debug("Check for expired workers")
		err := checkForExpired(wrks, time.Now())
		if err != nil {
			cmdapp.Log.Error(err)
		}
	}
}

func checkForExpired(wrks *workers, t time.Time) error {
	tp := t.Add(-100 * time.Second)
	wrks.lock.Lock()
	defer wrks.lock.Unlock()

	cmdapp.Log.Debugf("Worker count: %d", len(wrks.workers))
	for _, w := range wrks.workers {
		if w.beatTime.Before(tp) {
			cmdapp.Log.Infof("Worker is dead? %s. Last seen %v", w.queue, w.beatTime)
			dropWorker(wrks, w)
		}
	}
	return nil
}

func (w *worker) completeTask() error {
	cmdapp.Log.Infof("Task response received from %s at %s. Was estimated %s.", w.queue,
		time.Now().Format(timeFormat), w.endAt.Format(timeFormat))
	if w.working == false {
		return errors.New("Task already not working")
	}
	w.task = nil
	w.working = false
	w.endAt = time.Now()
	return nil
}

func (w *worker) startTask(t *task) error {
	return w.startTaskAt(t, time.Now())
}

func (w *worker) startTaskAt(t *task, now time.Time) error {
	if w.working {
		return errors.Errorf("Tryning to start worker %s, but it is already marked as working", w.queue)
	}
	w.working = true
	w.task = t
	w.started = now
	w.endAt = w.started.Add(t.expDuration * time.Duration(t.rtFactor))
	if w.mType != t.requiredModelType {
		if t.requiredModelType != "" {
			w.mType = t.requiredModelType
		} else {
			w.mType = noneWorkerModelType
		}
		w.endAt = w.endAt.Add(t.expModelLoadDuration)
	}
	cmdapp.Log.Infof("Estimated complete time at %s", w.endAt.Format(timeFormat))
	return nil
}
