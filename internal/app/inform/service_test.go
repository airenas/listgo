package inform

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/pkg/errors"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks1"

	"github.com/petergtz/pegomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/streadway/amqp"
)

var senderMock *mocks1.MockSender
var message amqp.Delivery
var emailMakerMock *mocks.MockEmailMaker
var emailRetrieverMock *mocks.MockEmailRetriever
var lockerMock *mocks.MockLocker
var ackMock *mocks.MockAcknowledger

func initTest(t *testing.T) {
	mocks.AttachMockToConvey(t)
	ackMock = mocks.NewMockAcknowledger()
	msgdata, _ := json.Marshal(messages.InformMessage{QueueMessage: messages.QueueMessage{ID: "id"}, Type: "it", At: time.Now().UTC()})
	message = amqp.Delivery{Body: msgdata}
	message.Acknowledger = ackMock

	senderMock = mocks1.NewMockSender()
	emailMakerMock = mocks.NewMockEmailMaker()
	emailRetrieverMock = mocks.NewMockEmailRetriever()
	lockerMock = mocks.NewMockLocker()
}

func TestHandlesMessages(t *testing.T) {
	Convey("Given a service", t, func() {
		initTest(t)
		// init worker service
		wc := make(chan amqp.Delivery)
		data := ServiceData{}
		data.taskName = "x"
		data.workCh = wc
		data.emailSender = senderMock
		data.emailMaker = emailMakerMock
		data.emailRetriever = emailRetrieverMock
		data.locker = lockerMock
		fc, _ := StartWorkerService(&data)
		Convey("When wrong msg is put", func() {
			message.Body = make([]byte, 0)
			wc <- message
			close(wc)
			<-fc // wait for complete
			Convey("Nack is called", func() {
				ackMock.VerifyWasCalledOnce().Nack(pegomock.AnyUint64(), pegomock.AnyBool(), pegomock.AnyBool())
			})
		})
		Convey("When good msg is put", func() {
			wc <- message
			close(wc)
			<-fc // wait for complete
			Convey("email send", func() {
				senderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyPtrToEmailEmail())
			})
			Convey("Ack is called", func() {
				ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
			})
			Convey("Lock is called ", func() {
				lockerMock.VerifyWasCalledOnce().Lock(pegomock.EqString("id"), pegomock.EqString("it"))
			})
			Convey("UnLock is called ", func() {
				_, _, ut := lockerMock.VerifyWasCalledOnce().UnLock(pegomock.EqString("id"),
					pegomock.EqString("it"), matchers.AnyPtrToInt()).GetCapturedArguments()
				So(*ut, ShouldEqual, 2)
			})
		})
		Convey("When Maker fails", func() {
			pegomock.When(emailMakerMock.Make(matchers.AnyPtrToInformData())).ThenReturn(nil, errors.New("error"))

			wc <- message
			close(wc)
			<-fc // wait for complete
			Convey("Nack is called", func() {
				ackMock.VerifyWasCalledOnce().Nack(pegomock.AnyUint64(), pegomock.AnyBool(), pegomock.AnyBool())
			})
		})
		Convey("When EmailRetriever fails", func() {
			pegomock.When(emailRetrieverMock.Get(pegomock.AnyString())).ThenReturn("", errors.New("error"))

			wc <- message
			close(wc)
			<-fc // wait for complete
			Convey("Nack is called", func() {
				ackMock.VerifyWasCalledOnce().Nack(pegomock.AnyUint64(), pegomock.AnyBool(), pegomock.AnyBool())
			})
		})
		Convey("When Sender fails", func() {
			pegomock.When(senderMock.Send(matchers.AnyPtrToEmailEmail())).ThenReturn(errors.New("error"))

			wc <- message
			close(wc)
			<-fc // wait for complete
			Convey("Nack is called", func() {
				ackMock.VerifyWasCalledOnce().Nack(pegomock.AnyUint64(), pegomock.AnyBool(), pegomock.AnyBool())
			})
			Convey("Lock is called ", func() {
				lockerMock.VerifyWasCalledOnce().Lock(pegomock.EqString("id"), pegomock.EqString("it"))
			})
			Convey("UnLock is called ", func() {
				_, _, ut := lockerMock.VerifyWasCalledOnce().UnLock(pegomock.EqString("id"),
					pegomock.EqString("it"), matchers.AnyPtrToInt()).GetCapturedArguments()
				So(*ut, ShouldEqual, 0)
			})
		})
		Convey("When Locker fails", func() {
			pegomock.When(lockerMock.Lock(pegomock.AnyString(), pegomock.AnyString())).ThenReturn(errors.New("error"))

			wc <- message
			close(wc)
			<-fc // wait for complete
			Convey("Nack is called", func() {
				ackMock.VerifyWasCalledOnce().Nack(pegomock.AnyUint64(), pegomock.AnyBool(), pegomock.AnyBool())
			})
		})
	})
}

func TestCheckInputParameters(t *testing.T) {
	Convey("Given service", t, func() {
		initTest(t)
		wc := make(chan amqp.Delivery)
		data := ServiceData{}
		data.taskName = "x"
		data.workCh = wc
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
			data.workCh = nil
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
		Convey("Given no TaskName", func() {
			data.taskName = ""
			_, error := StartWorkerService(&data)
			Convey("Should return error", func() {
				So(error, ShouldNotBeNil)
			})
		})
		close(wc)
	})
}
