package dispatcher

import (
	"bitbucket.org/airenas/listgo/internal/pkg/strategy/api"
	"errors"
	"testing"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"
	"github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
)

var taskSelectorMock *mocks.MockTaskSelector

func initTestStrategy(t *testing.T) {
	mocks.AttachMockToTest(t)
	taskSelectorMock = mocks.NewMockTaskSelector()
}

func TestInitStrategy(t *testing.T) {
	initTestStrategy(t)
	s, err := newStrategyWrapper(taskSelectorMock)
	assert.Nil(t, err)
	assert.NotNil(t, s)
}

func TestInitStrategy_NoSelector(t *testing.T) {
	initTestStrategy(t)
	_, err := newStrategyWrapper(nil)
	assert.NotNil(t, err)
}

func TestStrategy_MapsWorker(t *testing.T) {
	now := time.Now()
	res := mapWorkers([]*worker{{endAt: now, mType: "olia"}})
	assert.Equal(t, 1, len(res))
	assert.Equal(t, "olia", res[0].TaskType)
	assert.Equal(t, now, res[0].EndAt)
}

func TestStrategy_MapsTask(t *testing.T) {
	now := time.Now()
	tsk := &task{addedAt: now, expDuration: time.Second, requiredModelType: "olia", started: false}
	res := mapTasks(map[string]*task{"1": tsk})
	assert.Equal(t, 1, len(res))
	assert.Equal(t, "olia", res[0].TaskType)
	assert.Equal(t, time.Second, res[0].Duration)
	assert.Equal(t, now, res[0].ArrivedAt)
	assert.Equal(t, tsk, res[0].RealObject)
}

func TestStrategy_MapsTask_SkipsStarted(t *testing.T) {
	now := time.Now()
	tsk := &task{addedAt: now, expDuration: time.Second, requiredModelType: "olia", started: true}
	res := mapTasks(map[string]*task{"1": tsk})
	assert.Equal(t, 0, len(res))
}

func TestStrategy_FindBest(t *testing.T) {
	initTestStrategy(t)
	s, _ := newStrategyWrapper(taskSelectorMock)
	now := time.Now()
	wrks := []*worker{{endAt: now, mType: "olia"}}
	tsk := &task{addedAt: now, expDuration: time.Second, requiredModelType: "olia", started: false}
	rTask := &api.Task{RealObject: tsk}
	pegomock.When(taskSelectorMock.FindBest(matchers.AnySliceOfPtrToApiWorker(), matchers.AnySliceOfPtrToApiTask(),
		pegomock.AnyInt())).ThenReturn(rTask, nil)
	tsks := map[string]*task{"1": tsk}
	res, err := s.findBest(wrks, tsks, 0)
	assert.Nil(t, err)
	assert.Equal(t, tsk, res)
}

func TestStrategy_FindBestWithError(t *testing.T) {
	initTestStrategy(t)
	s, _ := newStrategyWrapper(taskSelectorMock)
	now := time.Now()
	wrks := []*worker{{endAt: now, mType: "olia"}}
	tsk := &task{addedAt: now, expDuration: time.Second, requiredModelType: "olia", started: false}
	pegomock.When(taskSelectorMock.FindBest(matchers.AnySliceOfPtrToApiWorker(), matchers.AnySliceOfPtrToApiTask(),
		pegomock.AnyInt())).ThenReturn(nil, errors.New("error"))
	tsks := map[string]*task{"1": tsk}
	_, err := s.findBest(wrks, tsks, 0)
	assert.NotNil(t, err)
}
