package tasks

import (
	"bytes"
	"strings"
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
	assert.False(t, r.Running())
}

func TestRuns_GetMsg(t *testing.T) {
	r := newTestRunner(t)
	defer r.Close()

	var b bytes.Buffer
	r.outWriter = &b
	err := r.Run("echo olia", nil)
	assert.Nil(t, err)
	time.Sleep(10 * time.Millisecond)
	r.Close()
	s := strings.TrimSpace(string(b.Bytes()))
	assert.Equal(t, "olia", s)
}

func TestRuns_GetErrorMsg(t *testing.T) {
	r := newTestRunner(t)
	defer r.Close()

	var b bytes.Buffer
	r.errWriter = &b
	err := r.Run("cat xxxx", nil)
	assert.Nil(t, err)
	time.Sleep(10 * time.Millisecond)
	r.Close()
	s := strings.TrimSpace(string(b.Bytes()))
	assert.Contains(t, s, "xxxx")
}

func TestRuns_TakesEnv(t *testing.T) {
	r := newTestRunner(t)
	defer r.Close()

	var b bytes.Buffer
	r.outWriter = &b
	r.errWriter = &b
	err := r.Run("sh -c \"echo $DATATTT\"", []string{"DATATTT=olia"})
	assert.Nil(t, err)
	time.Sleep(10 * time.Millisecond)
	r.Close()
	s := strings.TrimSpace(string(b.Bytes()))
	assert.Equal(t, "olia", s)
}

func TestSplit(t *testing.T) {
	assert.Equal(t, []string{"o"}, stringToArgs("o"))
	assert.Equal(t, []string{"olia"}, stringToArgs("olia"))
	assert.Equal(t, []string{"olia", "1", "2"}, stringToArgs("olia 1   2"))
	assert.Equal(t, []string{}, stringToArgs(""))
	assert.Equal(t, []string{"olia", "-c", "opa opa"}, stringToArgs("olia -c 'opa opa'"))
	assert.Equal(t, []string{"olia", "-c", "opa opa"}, stringToArgs("olia -c \"opa opa\""))
}

func TestSplitComplex(t *testing.T) {
	assert.Equal(t, []string{"olia", "-c", "opa \\\"opa"}, stringToArgs("olia -c \"opa \\\"opa\""))
}

func newTestRunner(t *testing.T) *Runner {
	r, err := NewRunner("")
	assert.Nil(t, err)
	assert.NotNil(t, r)
	return r
}
