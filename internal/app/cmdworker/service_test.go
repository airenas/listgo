package cmdworker

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/airenas/listgo/internal/pkg/recognizer"
	"github.com/airenas/listgo/internal/pkg/test/mocks"
	"github.com/airenas/listgo/internal/pkg/test/mocks/matchers"
	"github.com/airenas/listgo/internal/pkg/utils"

	"github.com/airenas/listgo/internal/pkg/messages"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

var ackMock *mocks.MockAcknowledger
var message amqp.Delivery
var msgSenderMock *mocks.MockSenderWithCorr
var recInfoLoaderMock *mocks.MockRecInfoLoader
var preloadTaskManagerMock *mocks.MockPreloadTaskManager

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	ackMock = mocks.NewMockAcknowledger()
	msgdata, _ := json.Marshal(messages.NewQueueMessage("1", "rec", nil))
	message = amqp.Delivery{Body: msgdata}
	message.Acknowledger = ackMock
	msgSenderMock = mocks.NewMockSenderWithCorr()
	recInfoLoaderMock = mocks.NewMockRecInfoLoader()
	preloadTaskManagerMock = mocks.NewMockPreloadTaskManager()
	pegomock.When(recInfoLoaderMock.Get(pegomock.AnyString())).ThenReturn(&recognizer.Info{}, nil)
}

func initData(t *testing.T, wc chan amqp.Delivery) ServiceData {
	data := ServiceData{}
	data.Command = "ls -la"
	data.WorkingDir = "."
	data.Name = "olia"
	data.MessageSender = msgSenderMock
	data.RecInfoLoader = recInfoLoaderMock
	data.PreloadManager = preloadTaskManagerMock
	data.WorkCh = wc
	data.quitChannel = utils.NewMultiCloseChannel()
	data.reapLock = &sync.RWMutex{}
	return data
}

func TestHandlesWrongMessages(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)

	StartWorkerService(&data)
	message.Body = make([]byte, 0)
	wc <- message
	close(wc)
	<-data.quitChannel.C // wait for complete
	msgSenderMock.VerifyWasCalled(pegomock.Never()).SendWithCorr(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyString())
	ackMock.VerifyWasCalledOnce().Nack(pegomock.AnyUint64(), pegomock.AnyBool(), pegomock.AnyBool())
}

func TestHandlesWrongWithReply(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	StartWorkerService(&data)

	message.ReplyTo = "rt"
	wc <- message
	close(wc)
	<-data.quitChannel.C // wait for complete
	msgSenderMock.VerifyWasCalled(pegomock.Once()).SendWithCorr(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyString())
	ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}

func TestHandlesGoodNoReply(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	StartWorkerService(&data)

	wc <- message
	close(wc)
	<-data.quitChannel.C // wait for complete
	msgSenderMock.VerifyWasCalled(pegomock.Never()).SendWithCorr(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyString())
	ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}

func TestHandlesWhenTaskFails(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	StartWorkerService(&data)

	data.Command = "lsss"
	message.ReplyTo = "rt"
	wc <- message
	close(wc)
	<-data.quitChannel.C // wait for complete
	cMsg, _, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).SendWithCorr(matchers.AnyMessagesMessage(),
		pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyString()).GetCapturedArguments()
	assert.NotEmpty(t, cMsg.(*messages.QueueMessage).Error)
	ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}

func TestHandlesWhenPreloadFails(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	StartWorkerService(&data)
	pegomock.When(preloadTaskManagerMock.EnsureRunning(matchers.AnyMapOfStringToString())).ThenReturn(errors.New("error"))
	message.ReplyTo = "rt"
	wc <- message
	close(wc)
	<-data.quitChannel.C // wait for complete
	cMsg, _, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).SendWithCorr(matchers.AnyMessagesMessage(),
		pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyString()).GetCapturedArguments()
	assert.NotEmpty(t, cMsg.(*messages.QueueMessage).Error)
	ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}

func TestHandlesLoaderFails(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	StartWorkerService(&data)

	pegomock.When(recInfoLoaderMock.Get(pegomock.AnyString())).ThenReturn(nil, errors.New("error"))
	message.ReplyTo = "rt"

	wc <- message
	close(wc)

	<-data.quitChannel.C // wait for complete
	cMsg, _, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).SendWithCorr(matchers.AnyMessagesMessage(),
		pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyString()).GetCapturedArguments()
	assert.NotEmpty(t, cMsg.(*messages.QueueMessage).Error)
	ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}

func TestHandlesLoaderFailsWithNoReply(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	StartWorkerService(&data)

	pegomock.When(recInfoLoaderMock.Get(pegomock.AnyString())).ThenReturn(nil, errors.New("error"))
	wc <- message
	close(wc)
	<-data.quitChannel.C // wait for complete
	ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}

func TestHandlesResultRequired(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	StartWorkerService(&data)

	data.ReadFunc = func(file string, id string) (string, error) {
		return "olia", nil
	}
	data.ResultFile = "rFile"
	message.ReplyTo = "rt"

	wc <- message
	close(wc)
	<-data.quitChannel.C // wait for complete
	cMsg, _, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).SendWithCorr(matchers.AnyMessagesMessage(),
		pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyString()).GetCapturedArguments()
	assert.Equal(t, cMsg.(*messages.ResultMessage).Result, "olia")
	ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}

func TestHandlesWithResultFailing(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	StartWorkerService(&data)

	data.ReadFunc = func(file string, id string) (string, error) {
		return "", errors.New("error")
	}
	data.ResultFile = "rFile"
	message.ReplyTo = "rt"

	wc <- message
	close(wc)
	<-data.quitChannel.C // wait for completeBuildTestingFailHandler
	cMsg, _, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).SendWithCorr(matchers.AnyMessagesMessage(),
		pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyString()).GetCapturedArguments()
	assert.NotEmpty(t, cMsg.(*messages.ResultMessage).Error)
	ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}

func TestCheckInputParametersNoFunction(t *testing.T) {
	wc := make(chan amqp.Delivery)
	data := ServiceData{}
	data.Command = "ls -la"
	data.WorkingDir = "."
	data.Name = "olia"
	data.WorkCh = wc

	data.ResultFile = "olia"
	error := StartWorkerService(&data)
	assert.NotNil(t, error)
}

func TestCheckInputParametersWithFunction(t *testing.T) {
	initTest(t)

	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	data.ReadFunc = ReadFile
	error := StartWorkerService(&data)
	assert.Nil(t, error)
}

func Test_NoRecInfoLoader(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	data.RecInfoLoader = nil

	err := StartWorkerService(&data)
	assert.NotNil(t, err)
	close(wc)
}

func Test_NoPreloadManager(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	data.PreloadManager = nil

	err := StartWorkerService(&data)
	assert.NotNil(t, err)
	close(wc)
}

func TestHandlesFailureWithNoAck(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	data.skipAck = true
	StartWorkerService(&data)

	message.ReplyTo = "rt"
	wc <- message
	close(wc)
	<-data.quitChannel.C // wait for complete
	msgSenderMock.VerifyWasCalled(pegomock.Once()).SendWithCorr(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyString())
	ackMock.VerifyWasCalled(pegomock.Never()).Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}

func TestHandlesOKWithNoAck(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	data.skipAck = true
	StartWorkerService(&data)

	wc <- message
	close(wc)
	<-data.quitChannel.C // wait for complete
	msgSenderMock.VerifyWasCalled(pegomock.Never()).SendWithCorr(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyString())
	ackMock.VerifyWasCalled(pegomock.Never()).Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}
