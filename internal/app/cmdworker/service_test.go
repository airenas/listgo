package cmdworker

import (
	"encoding/json"
	"testing"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/test"
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

func TestHandlesMessages(t *testing.T) {
	Convey("Given a worker", t, func() {
		wc := make(chan amqp.Delivery)
		data := ServiceData{}
		data.Command = "ls -la"
		data.WorkingDir = "."
		data.TaskName = "olia"
		ts := &test.Sender{Msgs: make([]test.Msg, 0)}
		data.MessageSender = ts
		data.WorkCh = wc
		go StartWorkerService(&data)
		Convey("When wrong msg is put", func() {
			d := amqp.Delivery{}
			ack := &mocks.Acknowledger{}
			ack.On("Nack").Return(nil)
			d.Acknowledger = ack
			wc <- d
			close(wc)
			Convey("No msg sent", func() {
				So(cap(ts.Msgs), ShouldEqual, 0)
			})
			Convey("Nack is called", func() {
				So(ack.AssertExpectations(t), ShouldBeTrue)
			})
		})
		Convey("When good msg is put with reply", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
			d := amqp.Delivery{Body: msgdata}
			ack := &mocks.Acknowledger{}
			ack.On("Ack").Return(nil)
			d.Acknowledger = ack
			d.ReplyTo = "rt"
			wc <- d
			close(wc)
			time.Sleep(time.Second * 2)
			Convey("msg replied", func() {
				So(cap(ts.Msgs), ShouldEqual, 1)
			})
			Convey("Ack is called", func() {
				So(ack.AssertExpectations(t), ShouldBeTrue)
			})
		})
		Convey("When good msg is put with no reply", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
			d := amqp.Delivery{Body: msgdata}
			ack := &mocks.Acknowledger{}
			ack.On("Ack").Return(nil)
			d.Acknowledger = ack
			d.ReplyTo = ""
			wc <- d
			close(wc)
			time.Sleep(time.Second * 2)
			Convey("No msg replied", func() {
				So(cap(ts.Msgs), ShouldEqual, 0)
			})
			Convey("Ack is called", func() {
				So(ack.AssertExpectations(t), ShouldBeTrue)
			})
		})
	})
}
