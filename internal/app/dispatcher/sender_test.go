package dispatcher

import (
	"testing"

	"github.com/airenas/listgo/internal/pkg/messages"
	"github.com/airenas/listgo/internal/pkg/test/mocks"
	"github.com/airenas/listgo/internal/pkg/test/mocks/matchers"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var senderMock *mocks.MockSenderWithCorr

func initTestSender(t *testing.T) {
	mocks.AttachMockToTest(t)
	senderMock = mocks.NewMockSenderWithCorr()
}

func TestInitSender(t *testing.T) {
	initTestSender(t)
	s, err := newMsgWithCorrSender(senderMock, "queue")
	assert.Nil(t, err)
	assert.NotNil(t, s)
}

func TestInitSender_NoSender(t *testing.T) {
	initTestSender(t)
	_, err := newMsgWithCorrSender(nil, "queue")
	assert.NotNil(t, err)
}

func TestInitSender_NoQueue(t *testing.T) {
	initTestSender(t)
	_, err := newMsgWithCorrSender(nil, "")
	assert.NotNil(t, err)
}

func TestSender_Send(t *testing.T) {
	initTestSender(t)
	s, err := newMsgWithCorrSender(senderMock, "replyQueue")
	msg := messages.NewQueueMessage("id", "", nil)
	err = s.Send(msg, "oliaq", "corrID")
	assert.Nil(t, err)
	capMsg, capP2, capP3, capP4 := senderMock.VerifyWasCalled(pegomock.Once()).
		SendWithCorr(matchers.AnyMessagesMessage(),
			pegomock.AnyString(), pegomock.AnyString(), pegomock.AnyString()).GetCapturedArguments()
	assert.Equal(t, msg, capMsg)
	assert.Equal(t, "oliaq", capP2)
	assert.Equal(t, "replyQueue", capP3)
	assert.Equal(t, "corrID", capP4)
}

func TestSender_Fail(t *testing.T) {
	initTestSender(t)
	s, err := newMsgWithCorrSender(senderMock, "queue")
	msg := messages.NewQueueMessage("id", "", nil)
	pegomock.When(senderMock.SendWithCorr(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString(),
		pegomock.AnyString())).ThenReturn(errors.New("err"))
	err = s.Send(msg, "oliaq", "corrID")
	assert.NotNil(t, err)
}
