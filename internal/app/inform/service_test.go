package inform

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pkg/errors"

	"github.com/airenas/listgo/internal/pkg/messages"
	"github.com/airenas/listgo/internal/pkg/test/mocks"
	"github.com/airenas/listgo/internal/pkg/test/mocks/matchers"
	"github.com/airenas/listgo/internal/pkg/test/mocks1"
	"github.com/airenas/listgo/internal/pkg/utils"

	"github.com/petergtz/pegomock"
	"github.com/streadway/amqp"
)

var senderMock *mocks1.MockSender
var message amqp.Delivery
var emailMakerMock *mocks.MockEmailMaker
var emailRetrieverMock *mocks.MockEmailRetriever
var lockerMock *mocks.MockLocker
var ackMock *mocks.MockAcknowledger
var wc chan amqp.Delivery
var data *ServiceData

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	ackMock = mocks.NewMockAcknowledger()
	msgdata, _ := json.Marshal(messages.InformMessage{QueueMessage: messages.QueueMessage{ID: "id"}, Type: "it", At: time.Now().UTC()})
	message = amqp.Delivery{Body: msgdata}
	message.Acknowledger = ackMock

	senderMock = mocks1.NewMockSender()
	emailMakerMock = mocks.NewMockEmailMaker()
	emailRetrieverMock = mocks.NewMockEmailRetriever()
	lockerMock = mocks.NewMockLocker()
	wc = make(chan amqp.Delivery)
	data = initData(t, wc)
}

func initData(t *testing.T, wc chan amqp.Delivery) *ServiceData {
	data := ServiceData{}
	data.taskName = "x"
	data.workCh = wc
	data.emailSender = senderMock
	data.emailMaker = emailMakerMock
	data.emailRetriever = emailRetrieverMock
	data.locker = lockerMock
	data.fc = utils.NewMultiCloseChannel()
	return &data
}

func TestHandlesMessagesWhenWrongMsg(t *testing.T) {
	initTest(t)
	StartWorkerService(data)

	message.Body = make([]byte, 0)
	wc <- message
	close(wc)
	<-data.fc.C // wait for complete
	ackMock.VerifyWasCalledOnce().Nack(pegomock.AnyUint64(), pegomock.AnyBool(), pegomock.AnyBool())
}

func TestHandlesMessagesWhenGoodMsg(t *testing.T) {
	initTest(t)
	StartWorkerService(data)

	wc <- message
	close(wc)
	<-data.fc.C // wait for complete
	senderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyPtrToEmailEmail())
	ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
	lockerMock.VerifyWasCalledOnce().Lock(pegomock.EqString("id"), pegomock.EqString("it"))
	_, _, ut := lockerMock.VerifyWasCalledOnce().UnLock(pegomock.EqString("id"),
		pegomock.EqString("it"), matchers.AnyPtrToInt()).GetCapturedArguments()
	assert.Equal(t, *ut, 2)
}

func TestHandlesMessagesWhenMakerFails(t *testing.T) {
	initTest(t)
	StartWorkerService(data)
	pegomock.When(emailMakerMock.Make(matchers.AnyPtrToInformData())).ThenReturn(nil, errors.New("error"))

	wc <- message
	close(wc)
	<-data.fc.C // wait for complete
	ackMock.VerifyWasCalledOnce().Nack(pegomock.AnyUint64(), pegomock.AnyBool(), pegomock.AnyBool())
}

func TestHandlesMessagesWhenEmailRetrieverFails(t *testing.T) {
	initTest(t)
	StartWorkerService(data)
	pegomock.When(emailRetrieverMock.Get(pegomock.AnyString())).ThenReturn("", errors.New("error"))

	wc <- message
	close(wc)
	<-data.fc.C // wait for complete
	ackMock.VerifyWasCalledOnce().Nack(pegomock.AnyUint64(), pegomock.AnyBool(), pegomock.AnyBool())
}

func TestHandlesMessagesWhenSenderFails(t *testing.T) {
	initTest(t)
	StartWorkerService(data)
	pegomock.When(senderMock.Send(matchers.AnyPtrToEmailEmail())).ThenReturn(errors.New("error"))

	wc <- message
	close(wc)
	<-data.fc.C // wait for complete
	ackMock.VerifyWasCalledOnce().Nack(pegomock.AnyUint64(), pegomock.AnyBool(), pegomock.AnyBool())
	lockerMock.VerifyWasCalledOnce().Lock(pegomock.EqString("id"), pegomock.EqString("it"))
	_, _, ut := lockerMock.VerifyWasCalledOnce().UnLock(pegomock.EqString("id"),
		pegomock.EqString("it"), matchers.AnyPtrToInt()).GetCapturedArguments()
	assert.Equal(t, *ut, 0)
}

func TestHandlesMessagesWhenLockerFails(t *testing.T) {
	initTest(t)
	StartWorkerService(data)
	pegomock.When(lockerMock.Lock(pegomock.AnyString(), pegomock.AnyString())).ThenReturn(errors.New("error"))

	wc <- message
	close(wc)
	<-data.fc.C // wait for complete
	ackMock.VerifyWasCalledOnce().Nack(pegomock.AnyUint64(), pegomock.AnyBool(), pegomock.AnyBool())
}

func TestCheckInputParameters(t *testing.T) {
	initTest(t)
	error := StartWorkerService(data)
	assert.Nil(t, error)
}

func TestCheckInputParametersNoChannel(t *testing.T) {
	initTest(t)
	data.workCh = nil
	error := StartWorkerService(data)
	assert.NotNil(t, error)
}

func TestCheckInputParametersNoEmailMaker(t *testing.T) {
	initTest(t)
	data.emailMaker = nil
	error := StartWorkerService(data)
	assert.NotNil(t, error)
}

func TestCheckInputParametersNoEmailRetriever(t *testing.T) {
	initTest(t)
	data.emailRetriever = nil
	error := StartWorkerService(data)
	assert.NotNil(t, error)
}

func TestCheckInputParametersNoLocker(t *testing.T) {
	initTest(t)
	data.locker = nil
	error := StartWorkerService(data)
	assert.NotNil(t, error)
}

func TestCheckInputParametersNoTaskName(t *testing.T) {
	initTest(t)
	data.taskName = ""
	error := StartWorkerService(data)
	assert.NotNil(t, error)
}

func TestCheckInputParametersNoCloseChannel(t *testing.T) {
	initTest(t)
	data.fc = nil
	error := StartWorkerService(data)
	assert.NotNil(t, error)
}
