package dispatcher

import (
	"errors"
	"testing"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"
	"github.com/petergtz/pegomock"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

var wrkSenderMock *mocks.MockSender
var durGetterMock *mocks.MockDurationGetter
var modelTypeGetterMock *mocks.MockModelTypeGetter
var startTimeGetterMock *mocks.MockStartTimeGetter

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	msgSenderMock = mocks.NewMockSender()
	wrkSenderMock = mocks.NewMockSender()
	durGetterMock = mocks.NewMockDurationGetter()
	modelTypeGetterMock = mocks.NewMockModelTypeGetter()
	taskSelectorMock = mocks.NewMockTaskSelector()
	startTimeGetterMock = mocks.NewMockStartTimeGetter()
}

func initTestData(t *testing.T) *ServiceData {
	data := &ServiceData{}
	data.fc = utils.NewMultiCloseChannel()
	data.wrkrs = newWorkers()
	data.tsks = newTasks()
	// make same lock
	data.tsks.lock = data.wrkrs.lock

	data.startTimeGetter = startTimeGetterMock
	data.durationGetter = durGetterMock
	data.modelTypeGetter = modelTypeGetterMock
	data.selectionStrategy, _ = newStrategyWrapper(taskSelectorMock)

	data.replySender = msgSenderMock
	data.workSender = wrkSenderMock
	data.RegistrationCh = make(chan amqp.Delivery)
	data.WorkCh = make(chan amqp.Delivery)
	data.ResponseCh = make(chan amqp.Delivery)
	return data
}

func TestServiceInit(t *testing.T) {
	initTest(t)
	data := initTestData(t)
	err := StartWorkerService(data)
	assert.Nil(t, err)
}

func TestServiceInit_Fails(t *testing.T) {
	initTest(t)
	data := initTestData(t)
	data.durationGetter = nil
	assert.NotNil(t, StartWorkerService(data))

	data = initTestData(t)
	data.fc = nil
	assert.NotNil(t, StartWorkerService(data))

	data = initTestData(t)
	data.modelTypeGetter = nil
	assert.NotNil(t, StartWorkerService(data))

	data = initTestData(t)
	data.replySender = nil
	assert.NotNil(t, StartWorkerService(data))

	data = initTestData(t)
	data.selectionStrategy = nil
	assert.NotNil(t, StartWorkerService(data))

	data = initTestData(t)
	data.startTimeGetter = nil
	assert.NotNil(t, StartWorkerService(data))

	data = initTestData(t)
	data.workSender = nil
	assert.NotNil(t, StartWorkerService(data))

	data = initTestData(t)
	data.wrkrs = nil
	assert.NotNil(t, StartWorkerService(data))

	data = initTestData(t)
	data.tsks = nil
	assert.NotNil(t, StartWorkerService(data))
}

func TestServiceAddTask(t *testing.T) {
	initTest(t)
	data := initTestData(t)
	err := StartWorkerService(data)
	assert.Nil(t, err)
	msg := messages.NewQueueMessage("ID", "model", nil)
	d := newTestDelivery(msg)
	now := time.Now()
	pegomock.When(startTimeGetterMock.Get(matchers.AnySliceOfMessagesTag())).ThenReturn(now, nil)
	pegomock.When(modelTypeGetterMock.Get(pegomock.AnyString())).ThenReturn("mmm", nil)
	pegomock.When(durGetterMock.Get(pegomock.AnyString())).ThenReturn(time.Second, nil)

	err = addTask(data, d, msg)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(data.tsks.tsks))
	tsk := data.tsks.tsks["ID"]
	assert.NotNil(t, tsk)
	assert.Equal(t, msg, tsk.msg)
	assert.Equal(t, d, tsk.d)
	assert.Equal(t, now, tsk.addedAt)
	assert.Equal(t, "mmm", tsk.requiredModelType)
	assert.Equal(t, time.Second, tsk.expDuration)
}

func TestServiceAddOnFailure(t *testing.T) {
	initTest(t)
	data := initTestData(t)
	err := StartWorkerService(data)
	assert.Nil(t, err)
	msg := messages.NewQueueMessage("ID", "model", nil)
	d := newTestDelivery(msg)
	now := time.Now()
	pegomock.When(startTimeGetterMock.Get(matchers.AnySliceOfMessagesTag())).ThenReturn(now, errors.New("olia"))
	pegomock.When(modelTypeGetterMock.Get(pegomock.AnyString())).ThenReturn("", errors.New("olia"))
	pegomock.When(durGetterMock.Get(pegomock.AnyString())).ThenReturn(time.Second, errors.New("olia"))

	err = addTask(data, d, msg)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(data.tsks.tsks))
	tsk := data.tsks.tsks["ID"]
	assert.NotNil(t, tsk)
	assert.Equal(t, msg, tsk.msg)
	assert.Equal(t, d, tsk.d)
	assert.Equal(t, now, tsk.addedAt)
	assert.Equal(t, "", tsk.requiredModelType)
	assert.Equal(t, time.Second, tsk.expDuration)
}
