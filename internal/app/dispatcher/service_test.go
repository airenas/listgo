package dispatcher

import (
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"

	"github.com/streadway/amqp"
)

// var ackMock *mocks.MockAcknowledger
// var message amqp.Delivery
// var msgSenderMock *mocks.MockSender
// var recInfoLoaderMock *mocks.MockRecInfoLoader
// var preloadTaskManagerMock *mocks.MockPreloadTaskManager

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	// ackMock = mocks.NewMockAcknowledger()
	// msgdata, _ := json.Marshal(messages.NewQueueMessage("1", "rec", nil))
	// message = amqp.Delivery{Body: msgdata}
	// message.Acknowledger = ackMock
	// msgSenderMock = mocks.NewMockSender()
	// recInfoLoaderMock = mocks.NewMockRecInfoLoader()
	// preloadTaskManagerMock = mocks.NewMockPreloadTaskManager()
	// pegomock.When(recInfoLoaderMock.Get(pegomock.AnyString())).ThenReturn(&recognizer.Info{}, nil)
}

func initData(t *testing.T, wc chan amqp.Delivery) ServiceData {
	data := ServiceData{}
	// data.Command = "ls -la"
	// data.WorkingDir = "."
	// data.TaskName = "olia"
	// data.MessageSender = msgSenderMock
	// data.RecInfoLoader = recInfoLoaderMock
	// data.PreloadManager = preloadTaskManagerMock
	// data.WorkCh = wc
	return data
}

func TestHandlesWrongWithReply(t *testing.T) {
	initTest(t)
	// wc := make(chan amqp.Delivery)
	// data := initData(t, wc)
	// fc, _ := StartWorkerService(&data)

	// message.ReplyTo = "rt"
	// wc <- message
	// close(wc)
	// <-fc // wait for complete
	// msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
	// ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
}
