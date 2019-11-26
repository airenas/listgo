package cmdworker

import (
	"encoding/json"
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/recognizer"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestRun_NoParameter_Fail(t *testing.T) {
	cmd := "ls"
	err := RunCommand(cmd, "/", "id", nil)

	assert.NotNil(t, err, "Error expected")
}

func TestRun_WrongParameter_Fail(t *testing.T) {
	cmd := "ls -{olia}"
	err := RunCommand(cmd, "/", "id", nil)

	assert.NotNil(t, err, "Error expected")
}
func TestRun(t *testing.T) {
	cmd := "ls -la"
	err := RunCommand(cmd, "/", "id", nil)
	assert.Nil(t, err)
}

func TestRun_ID_Changed(t *testing.T) {
	cmd := "ls -{ID}"
	err := RunCommand(cmd, "/", "la", nil)
	assert.Nil(t, err)
}

var ackMock *mocks.MockAcknowledger
var message amqp.Delivery
var msgSenderMock *mocks.MockSender
var recInfoLoaderMock *mocks.MockRecInfoLoader

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	ackMock = mocks.NewMockAcknowledger()
	msgdata, _ := json.Marshal(messages.NewQueueMessage("1", "rec", nil))
	message = amqp.Delivery{Body: msgdata}
	message.Acknowledger = ackMock
	msgSenderMock = mocks.NewMockSender()
	recInfoLoaderMock = mocks.NewMockRecInfoLoader()
	pegomock.When(recInfoLoaderMock.Get(pegomock.AnyString())).ThenReturn(&recognizer.Info{}, nil)

}

func initData(t *testing.T, wc chan amqp.Delivery) ServiceData {
	data := ServiceData{}
	data.Command = "ls -la"
	data.WorkingDir = "."
	data.TaskName = "olia"
	data.MessageSender = msgSenderMock
	data.RecInfoLoader = recInfoLoaderMock
	data.WorkCh = wc
	return data
}

func TestHandlesWrongMessages(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)

	fc, _ := StartWorkerService(&data)
	message.Body = make([]byte, 0)
	wc <- message
	close(wc)
	<-fc // wait for complete
	msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
	ackMock.VerifyWasCalledOnce().Nack(pegomock.AnyUint64(), pegomock.AnyBool(), pegomock.AnyBool())
}

func TestHandlesWrongWithReply(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	fc, _ := StartWorkerService(&data)

	message.ReplyTo = "rt"
	wc <- message
	close(wc)
	<-fc // wait for complete
	msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
	ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}

func TestHandlesGoodNoReply(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	fc, _ := StartWorkerService(&data)

	wc <- message
	close(wc)
	<-fc // wait for complete
	msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
	ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}

func TestHandlesWhenTaskFails(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	fc, _ := StartWorkerService(&data)

	data.Command = "lsss"
	message.ReplyTo = "rt"
	wc <- message
	close(wc)
	<-fc // wait for complete
	cMsg, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
		pegomock.AnyString(), pegomock.AnyString()).GetCapturedArguments()
	assert.NotEmpty(t, cMsg.(*messages.QueueMessage).Error)
	ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}

func TestHandlesLoaderFails(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	fc, _ := StartWorkerService(&data)

	pegomock.When(recInfoLoaderMock.Get(pegomock.AnyString())).ThenReturn(nil, errors.New("error"))
	message.ReplyTo = "rt"

	wc <- message
	close(wc)

	<-fc // wait for complete
	cMsg, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
		pegomock.AnyString(), pegomock.AnyString()).GetCapturedArguments()
	assert.NotEmpty(t, cMsg.(*messages.QueueMessage).Error)
	ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}

func TestHandlesLoaderFailsWithNoReply(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	fc, _ := StartWorkerService(&data)

	pegomock.When(recInfoLoaderMock.Get(pegomock.AnyString())).ThenReturn(nil, errors.New("error"))
	wc <- message
	close(wc)
	<-fc // wait for complete
	ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}

func TestHandlesResultRequired(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	fc, _ := StartWorkerService(&data)

	data.ReadFunc = func(file string, id string) (string, error) {
		return "olia", nil
	}
	data.ResultFile = "rFile"
	message.ReplyTo = "rt"

	wc <- message
	close(wc)
	<-fc // wait for complete
	cMsg, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
		pegomock.AnyString(), pegomock.AnyString()).GetCapturedArguments()
	assert.Equal(t, cMsg.(*messages.ResultMessage).Result, "olia")
	ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}

func TestHandlesWithResultFailing(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	fc, _ := StartWorkerService(&data)

	data.ReadFunc = func(file string, id string) (string, error) {
		return "", errors.New("error")
	}
	data.ResultFile = "rFile"
	message.ReplyTo = "rt"

	wc <- message
	close(wc)
	<-fc // wait for completeBuildTestingFailHandler
	cMsg, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
		pegomock.AnyString(), pegomock.AnyString()).GetCapturedArguments()
	assert.NotEmpty(t, cMsg.(*messages.ResultMessage).Error)
	ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}

func TestCheckInputParametersNoFunction(t *testing.T) {
	wc := make(chan amqp.Delivery)
	data := ServiceData{}
	data.Command = "ls -la"
	data.WorkingDir = "."
	data.TaskName = "olia"
	data.WorkCh = wc

	data.ResultFile = "olia"
	_, error := StartWorkerService(&data)
	assert.NotNil(t, error)
}

func TestCheckInputParametersWithFunction(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := ServiceData{}
	data.Command = "ls -la"
	data.WorkingDir = "."
	data.TaskName = "olia"
	data.WorkCh = wc

	data.ResultFile = "olia"
	data.ReadFunc = ReadFile
	data.RecInfoLoader = recInfoLoaderMock
	_, error := StartWorkerService(&data)
	assert.Nil(t, error)
}

func Test_NoRecInfoLoader(t *testing.T) {
	initTest(t)
	wc := make(chan amqp.Delivery)
	data := initData(t, wc)
	data.RecInfoLoader = nil

	_, err := StartWorkerService(&data)
	assert.NotNil(t, err)
	close(wc)
}
