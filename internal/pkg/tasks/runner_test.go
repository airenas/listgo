package tasks

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRuns(t *testing.T) {
	r := newTestRunner(t)
	defer r.Close()

	err := r.Run("ls", nil)
	assert.Nil(t, err)
}

func TestFails(t *testing.T) {
	r := newTestRunner(t)
	defer r.Close()

	err := r.Run("xxxxx", nil)
	assert.NotNil(t, err)
}

func TestRuns_NotRunning(t *testing.T) {
	r := newTestRunner(t)
	defer r.Close()

	err := r.Run("ls", nil)
	assert.Nil(t, err)
	time.Sleep(10 * time.Millisecond)
	assert.False(t, r.Running())
}

func TestRuns_Running(t *testing.T) {
	r := newTestRunner(t)
	defer r.Close()

	err := r.Run("sleep 30", nil)
	assert.Nil(t, err)
	assert.True(t, r.Running())
}

func TestRuns_Closed(t *testing.T) {
	r := newTestRunner(t)
	defer r.Close()

	err := r.Run("sleep 30", nil)
	assert.Nil(t, err)
	assert.True(t, r.Running())
	err = r.Close()
	assert.Nil(t, err)
	time.Sleep(10 * time.Millisecond)
	assert.False(t, r.Running())
}

func newTestRunner(t *testing.T) *Runner {
	r, err := NewRunner("")
	assert.Nil(t, err)
	assert.NotNil(t, r)
	return r
}
