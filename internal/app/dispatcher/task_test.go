package dispatcher

import (
	"encoding/json"
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

var msgSenderMock *mocks.MockSender
var ackMock *mocks.MockAcknowledger

func initTestTask(t *testing.T) {
	mocks.AttachMockToTest(t)
	msgSenderMock = mocks.NewMockSender()
	ackMock = mocks.NewMockAcknowledger()
}

func TestAddTask(t *testing.T) {
	initTestTask(t)
	tsks := newTasks()
	tsk := newTask()
	tsk.msg = messages.NewQueueMessage("cID", "res", nil)
	tsk.d = newTestDelivery(tsk.msg)

	err := tsks.addTask(tsk)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(tsks.tsks))
}

func TestAddTask_Fail(t *testing.T) {
	initTestTask(t)
	tsks := newTasks()
	tsk := newTask()
	tsk.msg = messages.NewQueueMessage("cID", "res", nil)

	err := tsks.addTask(tsk)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(tsks.tsks))

	tsk.d = newTestDelivery(tsk.msg)
	tsk.msg = nil

	err = tsks.addTask(tsk)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(tsks.tsks))
}

func TestProcessResponse(t *testing.T) {
	initTestTask(t)
	tsks := newTasks()
	tsk := newTask()
	tsk.worker = newWorker()
	tsk.msg = messages.NewQueueMessage("cID", "res", nil)
	tsk.d = newTestDelivery(tsk.msg)
	tsk.d.ReplyTo = "olia"
	err := tsks.addTask(tsk)
	assert.Nil(t, err)
	message := newTestDelivery(messages.NewQueueMessage("cID", "res", nil))

	err = tsks.processResponse(message, msgSenderMock)
	assert.Nil(t, err)
	ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
	assert.Equal(t, 0, len(tsks.tsks))
}

func TestProcessResponse_NackOnFailure(t *testing.T) {
	initTestTask(t)
	tsks := newTasks()
	tsk := newTask()
	tsk.worker = newWorker()
	tsk.msg = messages.NewQueueMessage("cID", "res", nil)
	tsk.d = newTestDelivery(tsk.msg)
	tsk.d.ReplyTo = "olia"
	err := tsks.addTask(tsk)
	assert.Nil(t, err)
	message := newTestDelivery(messages.NewQueueMessage("cID", "res", nil))
	pegomock.When(msgSenderMock.Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())).
		ThenReturn(errors.New("err"))

	err = tsks.processResponse(message, msgSenderMock)
	assert.Nil(t, err)
	ackMock.VerifyWasCalledOnce().Nack(pegomock.AnyUint64(), pegomock.AnyBool(), pegomock.AnyBool())
	assert.Equal(t, 0, len(tsks.tsks))
}

func TestStartOn(t *testing.T) {
	initTestTask(t)
	tsk := newTask()
	tsk.msg = messages.NewQueueMessage("cID", "res", nil)
	tsk.d = newTestDelivery(tsk.msg)
	w := newWorker()
	err := tsk.startOn(w, msgSenderMock)
	assert.Nil(t, err)
	msgSenderMock.VerifyWasCalledOnce().Send(matchers.EqMessagesMessage(tsk.msg), pegomock.AnyString(), pegomock.AnyString())
	assert.Equal(t, tsk, w.task)
}

func TestStartOn_SenderFails_Error(t *testing.T) {
	initTestTask(t)
	tsk := newTask()
	tsk.msg = messages.NewQueueMessage("cID", "res", nil)
	tsk.d = newTestDelivery(tsk.msg)
	w := newWorker()
	pegomock.When(msgSenderMock.Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())).
		ThenReturn(errors.New("err"))
	err := tsk.startOn(w, msgSenderMock)
	assert.NotNil(t, err)
	assert.Nil(t, w.task)
}

func newTestDelivery(msg *messages.QueueMessage) *amqp.Delivery {
	msgdata, _ := json.Marshal(msg)
	res := amqp.Delivery{Body: msgdata, CorrelationId: msg.ID}
	res.Acknowledger = ackMock
	return &res
}
