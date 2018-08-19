package cmdworker

import (
	"encoding/json"
	"testing"

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
		// init message
		msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
		d := amqp.Delivery{Body: msgdata}
		ack := &mocks.Acknowledger{}
		d.Acknowledger = ack

		Convey("When wrong msg is put", func() {
			d.Body = make([]byte, 0)
			ack.On("Nack").Return(nil)
			wc <- d
			close(wc)
			<-fc // wait for complete
			Convey("No msg sent", func() {
				So(cap(ts.Msgs), ShouldEqual, 0)
			})
			Convey("Nack is called", func() {
				So(ack.AssertExpectations(t), ShouldBeTrue)
			})
		})
		Convey("When good msg is put with reply", func() {
			ack.On("Ack").Return(nil)
			d.ReplyTo = "rt"
			wc <- d
			close(wc)
			<-fc // wait for complete
			Convey("msg replied", func() {
				So(cap(ts.Msgs), ShouldEqual, 1)
			})
			Convey("Ack is called", func() {
				So(ack.AssertExpectations(t), ShouldBeTrue)
			})
		})
		Convey("When good msg is put with no reply", func() {
			ack.On("Ack").Return(nil)
			wc <- d
			close(wc)
			<-fc // wait for complete
			Convey("No msg replied", func() {
				So(cap(ts.Msgs), ShouldEqual, 0)
			})
			Convey("Ack is called", func() {
				So(ack.AssertExpectations(t), ShouldBeTrue)
			})
		})
		Convey("When task fails", func() {
			data.Command = "lsss"
			ack.On("Ack").Return(nil)
			d.ReplyTo = "rt"
			wc <- d
			close(wc)
			<-fc // wait for complete
			Convey("msg replied", func() {
				So(cap(ts.Msgs), ShouldEqual, 1)
				So(ts.Msgs[0].M.Error, ShouldNotBeEmpty)
			})
			Convey("Ack is called", func() {
				So(ack.AssertExpectations(t), ShouldBeTrue)
			})
		})
	})
}
