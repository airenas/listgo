package inform

import (
	"encoding/json"
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks1"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/streadway/amqp"
)

var senderMock *mocks1.MockSender
var message amqp.Delivery
var msgSenderMock *mocks1.MockSender
var emailMakerMock *mocks.MockEmailMaker
var emailRetrieverMock *mocks.MockEmailRetriever
var lockerMock *mocks.MockLocker

func initTest(t *testing.T) {
	mocks.AttachMockToConvey(t)
	msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
	message = amqp.Delivery{Body: msgdata}
	senderMock = mocks1.NewMockSender()
	emailMakerMock = mocks.NewMockEmailMaker()
	emailRetrieverMock = mocks.NewMockEmailRetriever()
	lockerMock = mocks.NewMockLocker()
}

// func TestHandlesMessages(t *testing.T) {
// 	Convey("Given a worker", t, func() {
// 		initTest(t)
// 		// init worker service
// 		wc := make(chan amqp.Delivery)
// 		data := ServiceData{}
// 		data.Command = "ls -la"
// 		data.WorkingDir = "."
// 		data.TaskName = "olia"
// 		data.MessageSender = msgSenderMock
// 		data.WorkCh = wc
// 		fc, _ := StartWorkerService(&data)
// 		Convey("When wrong msg is put", func() {
// 			message.Body = make([]byte, 0)
// 			wc <- message
// 			close(wc)
// 			<-fc // wait for complete
// 			Convey("No msg sent", func() {
// 				msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
// 			})
// 			Convey("Nack is called", func() {
// 				ackMock.VerifyWasCalledOnce().Nack(pegomock.AnyUint64(), pegomock.AnyBool(), pegomock.AnyBool())
// 			})
// 		})
// 		Convey("When good msg is put with reply", func() {
// 			message.ReplyTo = "rt"
// 			wc <- message
// 			close(wc)
// 			<-fc // wait for complete
// 			Convey("msg replied", func() {
// 				msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
// 					pegomock.AnyString(), pegomock.AnyString())
// 			})
// 			Convey("Ack is called", func() {
// 				ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
// 			})
// 		})
// 		Convey("When good msg is put with no reply", func() {
// 			wc <- message
// 			close(wc)
// 			<-fc // wait for complete
// 			Convey("No msg replied", func() {
// 				msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
// 			})
// 			Convey("Ack is called", func() {
// 				ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
// 			})
// 		})
// 		Convey("When task fails", func() {
// 			data.Command = "lsss"
// 			message.ReplyTo = "rt"
// 			wc <- message
// 			close(wc)
// 			<-fc // wait for complete
// 			Convey("msg replied", func() {
// 				cMsg, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
// 					pegomock.AnyString(), pegomock.AnyString()).GetCapturedArguments()
// 				So(cMsg.(*messages.QueueMessage).Error, ShouldNotBeEmpty)
// 			})
// 			Convey("Ack is called", func() {
// 				ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
// 			})
// 		})
// 		Convey("When good msg is put with result required", func() {
// 			data.ReadFunc = func(file string, id string) (string, error) {
// 				return "olia", nil
// 			}
// 			data.ResultFile = "rFile"
// 			message.ReplyTo = "rt"

// 			wc <- message
// 			close(wc)
// 			<-fc // wait for complete
// 			Convey("msg replied with result", func() {
// 				cMsg, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
// 					pegomock.AnyString(), pegomock.AnyString()).GetCapturedArguments()
// 				So(cMsg.(*messages.ResultMessage).Result, ShouldEqual, "olia")
// 			})
// 			Convey("Ack is called", func() {
// 				ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
// 			})
// 		})
// 		Convey("When good msg is put with result failing", func() {
// 			data.ReadFunc = func(file string, id string) (string, error) {
// 				return "", errors.New("error")
// 			}
// 			data.ResultFile = "rFile"
// 			message.ReplyTo = "rt"

// 			wc <- message
// 			close(wc)
// 			<-fc // wait for completeBuildTestingFailHandler
// 			Convey("msg replied with error", func() {
// 				cMsg, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
// 					pegomock.AnyString(), pegomock.AnyString()).GetCapturedArguments()
// 				So(cMsg.(*messages.ResultMessage).Error, ShouldNotBeEmpty)
// 			})
// 			Convey("Ack is called", func() {
// 				ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
// 			})
// 		})
// 	})
// }

func TestCheckInputParameters(t *testing.T) {
	Convey("Given service", t, func() {
		initTest(t)
		wc := make(chan amqp.Delivery)
		data := ServiceData{}
		data.TaskName = "x"
		data.WorkCh = wc
		data.emailSender = senderMock
		data.emailMaker = emailMakerMock
		data.emailRetriever = emailRetrieverMock
		data.locker = lockerMock

		Convey("Given correct data", func() {
			_, error := StartWorkerService(&data)
			Convey("Should not return error", func() {
				So(error, ShouldBeNil)
			})
		})
		Convey("Given no channel", func() {
			data.WorkCh = nil
			_, error := StartWorkerService(&data)
			Convey("Should return error", func() {
				So(error, ShouldNotBeNil)
			})
		})
		Convey("Given no emailMaker", func() {
			data.emailMaker = nil
			_, error := StartWorkerService(&data)
			Convey("Should return error", func() {
				So(error, ShouldNotBeNil)
			})
		})
		Convey("Given no emailRetriever", func() {
			data.emailRetriever = nil
			_, error := StartWorkerService(&data)
			Convey("Should return error", func() {
				So(error, ShouldNotBeNil)
			})
		})
		Convey("Given no locker", func() {
			data.locker = nil
			_, error := StartWorkerService(&data)
			Convey("Should return error", func() {
				So(error, ShouldNotBeNil)
			})
		})
		close(wc)
	})
}
