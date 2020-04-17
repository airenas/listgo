package dispatcher

import (
	"testing"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
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

func newMsg(name string, tp string, t time.Time) *messages.RegistrationMessage {
	return &messages.RegistrationMessage{Queue: name, Type: tp, Working: false, Timestamp: t.Unix()}
}
