package cmdworker

import (
	"encoding/json"
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestRun_NoParameter_Fail(t *testing.T) {
	cmd := "ls"
	err := RunCommand(cmd, "/", "id")

	assert.NotNil(t, err, "Error expected")
}

func TestRun_WrongParameter_Fail(t *testing.T) {
	cmd := "ls -{olia}"
	err := RunCommand(cmd, "/", "id")

	assert.NotNil(t, err, "Error expected")
}
func TestRun(t *testing.T) {
	Convey("Given a command", t, func() {
		cmd := "ls -la"
		Convey("When the command is executed", func() {
			err := RunCommand(cmd, "/", "id")
			Convey("Then the result should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestRun_ID_Changed(t *testing.T) {
	Convey("Given a command with {ID} tag", t, func() {
		cmd := "ls -{ID}"
		Convey("When the command is executed", func() {
			err := RunCommand(cmd, "/", "la")
			Convey("Then the result should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func AttachToConvey(t *testing.T) pegomock.FailHandler {
	return func(message string, callerSkip ...int) {
		So(message, ShouldBeEmpty)
	}
}

var ackMock *mocks.MockAcknowledger
var message amqp.Delivery
var msgSenderMock *mocks.MockSender

func initTest(t *testing.T) {
	mocks.AttachMockToConvey(t)
	ackMock = mocks.NewMockAcknowledger()
	msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
	message = amqp.Delivery{Body: msgdata}
	message.Acknowledger = ackMock
	msgSenderMock = mocks.NewMockSender()
}

func TestHandlesMessages(t *testing.T) {
	Convey("Given a worker", t, func() {
		initTest(t)
		// init worker service
		wc := make(chan amqp.Delivery)
		data := ServiceData{}
		data.Command = "ls -la"
		data.WorkingDir = "."
		data.TaskName = "olia"
		data.MessageSender = msgSenderMock
		data.WorkCh = wc
		fc, _ := StartWorkerService(&data)
		Convey("When wrong msg is put", func() {
			message.Body = make([]byte, 0)
			wc <- message
			close(wc)
			<-fc // wait for complete
			Convey("No msg sent", func() {
				msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
			})
			Convey("Nack is called", func() {
				ackMock.VerifyWasCalledOnce().Nack(pegomock.AnyUint64(), pegomock.AnyBool(), pegomock.AnyBool())
			})
		})
		Convey("When good msg is put with reply", func() {
			message.ReplyTo = "rt"
			wc <- message
			close(wc)
			<-fc // wait for complete
			Convey("msg replied", func() {
				msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
					pegomock.AnyString(), pegomock.AnyString())
			})
			Convey("Ack is called", func() {
				ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
			})
		})
		Convey("When good msg is put with no reply", func() {
			wc <- message
			close(wc)
			<-fc // wait for complete
			Convey("No msg replied", func() {
				msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
			})
			Convey("Ack is called", func() {
				ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
			})
		})
		Convey("When task fails", func() {
			data.Command = "lsss"
			message.ReplyTo = "rt"
			wc <- message
			close(wc)
			<-fc // wait for complete
			Convey("msg replied", func() {
				cMsg, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
					pegomock.AnyString(), pegomock.AnyString()).GetCapturedArguments()
				So(cMsg.(*messages.QueueMessage).Error, ShouldNotBeEmpty)
			})
			Convey("Ack is called", func() {
				ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
			})
		})
		Convey("When good msg is put with result required", func() {
			data.ReadFunc = func(file string, id string) (string, error) {
				return "olia", nil
			}
			data.ResultFile = "rFile"
			message.ReplyTo = "rt"

			wc <- message
			close(wc)
			<-fc // wait for complete
			Convey("msg replied with result", func() {
				cMsg, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
					pegomock.AnyString(), pegomock.AnyString()).GetCapturedArguments()
				So(cMsg.(*messages.ResultMessage).Result, ShouldEqual, "olia")
			})
			Convey("Ack is called", func() {
				ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
			})
		})
		Convey("When good msg is put with result failing", func() {
			data.ReadFunc = func(file string, id string) (string, error) {
				return "", errors.New("error")
			}
			data.ResultFile = "rFile"
			message.ReplyTo = "rt"

			wc <- message
			close(wc)
			<-fc // wait for completeBuildTestingFailHandler
			Convey("msg replied with error", func() {
				cMsg, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
					pegomock.AnyString(), pegomock.AnyString()).GetCapturedArguments()
				So(cMsg.(*messages.ResultMessage).Error, ShouldNotBeEmpty)
			})
			Convey("Ack is called", func() {
				ackMock.VerifyWasCalledOnce().Ack(pegomock.AnyUint64(), pegomock.AnyBool())
			})
		})
	})
}

func TestCheckInputParameters(t *testing.T) {
	wc := make(chan amqp.Delivery)
	data := ServiceData{}
	data.Command = "ls -la"
	data.WorkingDir = "."
	data.TaskName = "olia"
	data.WorkCh = wc

	Convey("Given resultFile", t, func() {
		data.ResultFile = "olia"
		Convey("And no function", func() {
			_, error := StartWorkerService(&data)
			Convey("Error returned", func() {
				So(error, ShouldNotBeNil)
			})
		})
		Convey("Given function", func() {
			data.ReadFunc = ReadFile
			_, error := StartWorkerService(&data)
			Convey("No error returned", func() {
				So(error, ShouldBeNil)
			})
		})
	})
}
