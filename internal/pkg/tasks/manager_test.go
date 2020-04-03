package tasks

import (
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var runnerMock *mocks.MockProcessRunner

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	runnerMock = mocks.NewMockProcessRunner()
	pegomock.When(runnerMock.Running()).ThenReturn(true)
}

func TestFailsOnPrefix(t *testing.T) {
	initTest(t)
	m, err := newManager("", runnerMock)
	assert.NotNil(t, err)
	assert.Nil(t, m)
}

func TestFailsOnRunner(t *testing.T) {
	initTest(t)
	m, err := newManager("pr", nil)
	assert.NotNil(t, err)
	assert.Nil(t, m)
}

func TestInit_OK(t *testing.T) {
	m := newTestManager(t)
	assert.NotNil(t, m)
}

func TestCloses(t *testing.T) {
	m := newTestManager(t)
	m.Close()
	runnerMock.VerifyWasCalledOnce().Close()
}

func TestEnsure_NoKey(t *testing.T) {
	m := newTestManager(t)
	err := m.EnsureRunning(newMap("<none>", "olia cmd"))
	assert.NotNil(t, err)
}

func TestEnsure_StartsCmd(t *testing.T) {
	m := newTestManager(t)
	m.currentKey = "key"
	err := m.EnsureRunning(newMap("key", "olia cmd"))
	assert.Nil(t, err)
	runnerMock.VerifyWasCalled(pegomock.Never()).Run(pegomock.AnyString(), pegomock.AnyStringSlice())
}

func TestEnsure_DoesNotStartOnRunning(t *testing.T) {
	m := newTestManager(t)
	err := m.EnsureRunning(newMap("key", "olia cmd"))
	assert.Nil(t, err)
	cmd, env := runnerMock.VerifyWasCalledOnce().Run(pegomock.AnyString(), pegomock.AnyStringSlice()).GetCapturedArguments()
	assert.Equal(t, "olia cmd", cmd)
	assert.Contains(t, env, "PR_KEY=key")
	assert.Contains(t, env, "PR_CMD=olia cmd")
	assert.Equal(t, "key", m.currentKey)
}

func TestEnsure_StartOnNotRunning(t *testing.T) {
	m := newTestManager(t)
	m.currentKey = "key"
	pegomock.When(runnerMock.Running()).ThenReturn(false).ThenReturn(true)
	err := m.EnsureRunning(newMap("key", "olia cmd"))
	runnerMock.VerifyWasCalledOnce().Run(pegomock.AnyString(), pegomock.AnyStringSlice())
	assert.Nil(t, err)
}

func TestEnsure_CloseOnRunning(t *testing.T) {
	m := newTestManager(t)
	m.currentKey = "key"
	pegomock.When(runnerMock.Running()).ThenReturn(true).ThenReturn(false)
	err := m.EnsureRunning(newMap("", "olia cmd"))
	runnerMock.VerifyWasCalledOnce().Close()
	assert.Nil(t, err)
}

func TestEnsure_FailsOnClose(t *testing.T) {
	m := newTestManager(t)
	m.currentKey = "key"
	pegomock.When(runnerMock.Close()).ThenReturn(errors.New("error"))
	pegomock.When(runnerMock.Running()).ThenReturn(true)
	err := m.EnsureRunning(newMap("new", "olia cmd"))
	runnerMock.VerifyWasCalledOnce().Close()
	assert.NotNil(t, err)
}

func TestEnsure_FailsOnNoCmd(t *testing.T) {
	m := newTestManager(t)
	err := m.EnsureRunning(newMap("key", ""))
	runnerMock.VerifyWasCalled(pegomock.Never()).Run(pegomock.AnyString(), pegomock.AnyStringSlice())
	assert.NotNil(t, err)
}

func TestEnsure_FailsOnRunning(t *testing.T) {
	m := newTestManager(t)
	pegomock.When(runnerMock.Run(pegomock.AnyString(), pegomock.AnyStringSlice())).ThenReturn(errors.New("error"))
	err := m.EnsureRunning(newMap("key", "cmd"))
	runnerMock.VerifyWasCalled(pegomock.Once()).Run(pegomock.AnyString(), pegomock.AnyStringSlice())
	assert.NotNil(t, err)
}

func newTestManager(t *testing.T) *Manager {
	initTest(t)
	m, err := newManager("pr", runnerMock)
	assert.NotNil(t, m)
	assert.Nil(t, err)
	return m
}

func newMap(key, cmd string) map[string]string {
	res := make(map[string]string)
	if key != "<none>" {
		res["pr_key"] = key
	}
	if cmd != "<none>" {
		res["pr_cmd"] = cmd
	}
	return res
}
