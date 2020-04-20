package dispatcher

import (
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

// var ackMock *mocks.MockAcknowledger
// var message amqp.Delivery
var wrkSenderMock *mocks.MockSender
var durGetterMock *mocks.MockDurationGetter
var modelTypeGetterMock *mocks.MockModelTypeGetter
var startTimeGetterMock *mocks.MockStartTimeGetter

// var recInfoLoaderMock *mocks.MockRecInfoLoader
// var preloadTaskManagerMock *mocks.MockPreloadTaskManager

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	// ackMock = mocks.NewMockAcknowledger()
	// msgdata, _ := json.Marshal(messages.NewQueueMessage("1", "rec", nil))
	// message = amqp.Delivery{Body: msgdata}
	// message.Acknowledger = ackMock
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
