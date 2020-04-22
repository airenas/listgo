package strategy

import (
	"testing"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/strategy/api"
	"github.com/stretchr/testify/assert"
)

var now time.Time

func testInit(t *testing.T) {
	now = time.Now()
}

func TestInit(t *testing.T) {
	c, err := newCost(time.Second, 2, 3)
	assert.Nil(t, err)
	assert.NotNil(t, c)
}

func TestInit_Fails(t *testing.T) {
	_, err := newCost(time.Second*0, 2, 3)
	assert.NotNil(t, err)
	_, err = newCost(time.Second, 0, 3)
	assert.NotNil(t, err)
	_, err = newCost(time.Second, 500, 3)
	assert.NotNil(t, err)
	_, err = newCost(time.Second, 2, 0)
	assert.NotNil(t, err)
	_, err = newCost(time.Second, 2, 11)
	assert.NotNil(t, err)
}

func TestFind_Fails(t *testing.T) {
	testInit(t)
	s, err := newCost(time.Second*100, 2, 3)
	assert.NotNil(t, s)
	_, err = s.FindBest(nil, nil, 0)
	assert.NotNil(t, err)
	_, err = s.FindBest(testWrks(testW("", 0)), nil, -1)
	assert.NotNil(t, err)
	_, err = s.FindBest(testWrks(testW("", 0)), nil, 0)
	assert.Nil(t, err)
	_, err = s.FindBest(testWrks(testW("", 0)), nil, 2)
	assert.NotNil(t, err)
}

func TestFind(t *testing.T) {
	testInit(t)
	s, err := newCost(time.Second*100, 2, 3)
	assert.NotNil(t, s)
	bt, err := s.FindBest(testWrks(testW("1", 0), testW("2", 0)),
		testTsks(testT("2", 0, 20), testT("2", 0, 20), testT("1", 0, 20)), 0)
	assert.Nil(t, err)
	assert.NotNil(t, bt)
	assert.Equal(t, "1", bt.TaskType)
	bt, err = s.FindBest(testWrks(testW("1", 0), testW("2", 0)),
		testTsks(testT("2", 0, 20), testT("2", 0, 20), testT("1", 0, 20)), 1)
	assert.Nil(t, err)
	assert.NotNil(t, bt)
	assert.Equal(t, "2", bt.TaskType)
}

func TestFind_SelectLatest(t *testing.T) {
	testInit(t)
	s, _ := newCost(time.Second*100, 2, 3)
	assert.NotNil(t, s)
	t1 := testT("1", 0, 20)
	t2 := testT("1", 10, 20)
	t3 := testT("1", 20, 20)

	bt, err := s.FindBest(testWrks(testW("1", 0), testW("2", 0)), testTsks(t1, t2, t3), 0)
	assert.Nil(t, err)
	assert.Equal(t, t3, bt)
}

func TestFind_DoesReturn(t *testing.T) {
	testInit(t)
	s, _ := newCost(time.Second*100, 1, 0.01)
	assert.NotNil(t, s)
	t1 := testT("1", 0, 20)
	t2 := testT("1", 10, 20)
	t3 := testT("1", 20, 80)

	bt, _ := s.FindBest(testWrks(testW("1", 0), testW("2", 0)), testTsks(t1, t2, t3), 1)
	assert.Nil(t, bt)
}

func TestFind_StartsNewSame(t *testing.T) {
	testInit(t)
	s, _ := newCost(time.Second*100, 1, 0.01)
	assert.NotNil(t, s)
	t1 := testT("1", 0, 20)
	t2 := testT("1", 10, 20)
	t3 := testT("1", 20, 90)

	bt, _ := s.FindBest(testWrks(testW("1", 0), testW("2", 0)), testTsks(t1, t2, t3), 1)
	assert.Equal(t, t1, bt)
}

func testWrks(wrks ...*api.Worker) []*api.Worker {
	return wrks
}

func testW(mt string, addSec int) *api.Worker {
	res := &api.Worker{}
	res.TaskType = mt
	res.EndAt = now.Add(time.Second * time.Duration(addSec))
	return res
}

func testTsks(tsks ...*api.Task) []*api.Task {
	return tsks
}

func testT(mt string, arrivedBefore int, durSec int) *api.Task {
	res := &api.Task{}
	res.TaskType = mt
	res.ArrivedAt = now.Add(-time.Second * time.Duration(arrivedBefore))
	res.Duration = time.Second * time.Duration(durSec)
	return res
}
