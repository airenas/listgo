package dispatcher

import (
	"testing"
	"time"

	"github.com/airenas/listgo/internal/pkg/messages"
	"github.com/stretchr/testify/assert"
)

func TestExpires(t *testing.T) {
	now := time.Now()
	assert.True(t, expired(now.Add(-3*time.Minute), now))
	assert.True(t, expired(now.Add(-3000*time.Minute), now))
	assert.True(t, expired(now.Add(-121*time.Second), now))
	assert.False(t, expired(now.Add(-119*time.Second), now))
	assert.False(t, expired(now.Add(-1*time.Minute), now))
	assert.False(t, expired(now.Add(-0*time.Minute), now))
	assert.False(t, expired(now.Add(10*time.Minute), now))
}

func TestAddExpiredWorker(t *testing.T) {
	wrks := newWorkers()
	err := processWorker(wrks, newMsg("1", messages.RgrTypeRegister, time.Now().Add(-3*time.Minute)))
	assert.Nil(t, err)
	assert.Equal(t, 0, len(wrks.workers))
}

func TestProcessFails(t *testing.T) {
	wrks := newWorkers()
	err := processWorker(wrks, newMsg("1", "olia", time.Now()))
	assert.NotNil(t, err)
}

func TestAddWorker(t *testing.T) {
	wrks := newWorkers()
	err := processWorker(wrks, newMsg("1", messages.RgrTypeRegister, time.Now()))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(wrks.workers))
	processWorker(wrks, newMsg("2", messages.RgrTypeRegister, time.Now()))
	assert.Nil(t, err)
	assert.Equal(t, 2, len(wrks.workers))
}

func TestAddWitBeatWorker(t *testing.T) {
	wrks := newWorkers()
	processWorker(wrks, newMsg("1", messages.RgrTypeBeat, time.Now()))
	assert.Equal(t, 1, len(wrks.workers))
	processWorker(wrks, newMsg("2", messages.RgrTypeBeat, time.Now()))
	assert.Equal(t, 2, len(wrks.workers))
}

func TestAddSameWorker(t *testing.T) {
	wrks := newWorkers()
	processWorker(wrks, newMsg("1", messages.RgrTypeRegister, time.Now()))
	processWorker(wrks, newMsg("1", messages.RgrTypeRegister, time.Now()))
	assert.Equal(t, 1, len(wrks.workers))
}

func TestRemoveWorker(t *testing.T) {
	wrks := newWorkers()
	processWorker(wrks, newMsg("1", messages.RgrTypeRegister, time.Now()))
	assert.Equal(t, 1, len(wrks.workers))
	processWorker(wrks, newMsg("1", messages.RgrTypeExit, time.Now()))
	assert.Equal(t, 0, len(wrks.workers))
}

func TestRemoveOnExpire(t *testing.T) {
	wrks := newWorkers()
	processWorker(wrks, newMsg("1", messages.RgrTypeRegister, time.Now()))
	processWorker(wrks, newMsg("2", messages.RgrTypeRegister, time.Now().Add(45*time.Second)))
	assert.Equal(t, 2, len(wrks.workers))
	checkForExpired(wrks, time.Now().Add(2*time.Minute))
	assert.Equal(t, 1, len(wrks.workers))
}

func TestDoesNotRemoveOnExpire(t *testing.T) {
	wrks := newWorkers()
	processWorker(wrks, newMsg("1", messages.RgrTypeRegister, time.Now()))
	checkForExpired(wrks, time.Now().Add(20*time.Second))
	assert.Equal(t, 1, len(wrks.workers))
	processWorker(wrks, newMsg("1", messages.RgrTypeBeat, time.Now().Add(100*time.Second)))
	checkForExpired(wrks, time.Now().Add(120*time.Second))
	assert.Equal(t, 1, len(wrks.workers))
}

func TestWorkerComplete(t *testing.T) {
	wrk := newWorker()
	wrk.task = newTask()
	wrk.working = true

	wrk.completeTask()

	assert.Equal(t, false, wrk.working)
	assert.Nil(t, wrk.task)
	assert.False(t, wrk.endAt.After(time.Now()))
}

func TestWorkerCompleteFails(t *testing.T) {
	wrk := newWorker()
	wrk.working = false

	err := wrk.completeTask()
	assert.NotNil(t, err)
	assert.Equal(t, false, wrk.working)
}

func TestWorkerStartTask(t *testing.T) {
	wrk := newWorker()
	tsk := newTask()
	tsk.expDuration = time.Second
	tsk.expModelLoadDuration = time.Minute
	tsk.rtFactor = 2
	tsk.requiredModelType = ""
	now := time.Now()
	wrk.startTaskAt(tsk, now)

	assert.Equal(t, true, wrk.working)
	assert.Equal(t, tsk, wrk.task)
	assert.Equal(t, noneWorkerModelType, wrk.mType)
	assert.Equal(t, now.Add(time.Minute+time.Second*2), wrk.endAt)
}

func TestWorkerStartTask_Duration(t *testing.T) {
	wrk := newWorker()
	tsk := newTask()
	tsk.expDuration = time.Second
	tsk.expModelLoadDuration = time.Minute
	tsk.rtFactor = 2.5
	tsk.requiredModelType = ""
	now := time.Now()
	wrk.startTaskAt(tsk, now)

	assert.Equal(t, now.Add(time.Minute+time.Millisecond*2500), wrk.endAt)
}

func TestWorkerStartTask_NoModelLoad(t *testing.T) {
	wrk := newWorker()
	tsk := newTask()
	tsk.expDuration = time.Second
	tsk.expModelLoadDuration = time.Minute
	tsk.rtFactor = 2
	tsk.requiredModelType = "M1"
	now := time.Now()
	wrk.mType = "M1"
	err := wrk.startTaskAt(tsk, now)

	assert.Nil(t, err)
	assert.Equal(t, "M1", wrk.mType)
	assert.Equal(t, now.Add(time.Second*2), wrk.endAt)
}

func TestWorkerStartTask_Fails(t *testing.T) {
	wrk := newWorker()
	tsk := newTask()
	wrk.working = true
	err := wrk.startTaskAt(tsk, time.Now())
	assert.NotNil(t, err)
}

func TestDurTimes(t *testing.T) {
	assert.Equal(t, 500*time.Millisecond, durTimes(time.Second, 0.5))
	assert.Equal(t, 30*time.Minute, durTimes(time.Hour, 0.5))
	assert.Equal(t, 900*time.Millisecond, durTimes(time.Second, 0.9))
	assert.Equal(t, 1100*time.Millisecond, durTimes(time.Second, 1.1))
}

func TestDurTimesBigger(t *testing.T) {
	assert.Equal(t, 20*time.Hour, durTimes(100*time.Hour, 0.2))
}

func newMsg(name string, tp string, t time.Time) *messages.RegistrationMessage {
	return &messages.RegistrationMessage{Queue: name, Type: tp, Working: false, Timestamp: t.Unix()}
}
