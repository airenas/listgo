package cmdworker

import (
	"encoding/json"
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/test"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/streadway/amqp"
)

func TestRun_NoParameter_Fail(t *testing.T) {
	Convey("Given a command", t, func() {
		cmd := "ls"
		Convey("When the command is executed", func() {
			err := RunCommand(cmd, "/", "id")
			Convey("Then the result should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestRun_WrongParameter_Fail(t *testing.T) {
	Convey("Given a command", t, func() {
		cmd := "ls -{olia}"
		Convey("When the command is executed", func() {
			err := RunCommand(cmd, "/", "id")
			Convey("Then the result should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
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

func initTest(t *testing.T) {
	mocks.AttachMockToConvey(t)
	ackMock = mocks.NewMockAcknowledger()
	msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
	message = amqp.Delivery{Body: msgdata}
	message.Acknowledger = ackMock
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
		ts := &test.Sender{Msgs: make([]test.Msg, 0)}
		data.MessageSender = ts
		data.WorkCh = wc
		fc, _ := StartWorkerService(&data)
		Convey("When wrong msg is put", func() {
			message.Body = make([]byte, 0)
			wc <- message
			close(wc)
			<-fc // wait for complete
			Convey("No msg sent", func() {
				So(cap(ts.Msgs), ShouldEqual, 0)
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
				So(cap(ts.Msgs), ShouldEqual, 1)
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
				So(cap(ts.Msgs), ShouldEqual, 0)
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
				So(cap(ts.Msgs), ShouldEqual, 1)
				So(ts.Msgs[0].M.Error, ShouldNotBeEmpty)
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
				So(cap(ts.Msgs), ShouldEqual, 1)
				So(ts.Msgs[0].M.Result, ShouldEqual, "olia")
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
				So(cap(ts.Msgs), ShouldEqual, 1)
				So(ts.Msgs[0].M.Error, ShouldNotBeEmpty)
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
