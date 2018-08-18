package manager

import (
	"encoding/json"
	"io"
	"log"
	"testing"
	"time"

	"github.com/streadway/amqp"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/test"
	. "github.com/smartystreets/goconvey/convey"
)

func TestHandlesMessages(t *testing.T) {
	Convey("Given a manager", t, func() {
		dc := make(chan amqp.Delivery)
		ac := make(chan amqp.Delivery)
		diac := make(chan amqp.Delivery)
		tc := make(chan amqp.Delivery)
		rc := make(chan amqp.Delivery)
		data := ServiceData{}
		ts := testStatusSaver{statuses: make([]string, 0)}
		data.StatusSaver = &ts
		tsn := test.Sender{Msgs: make([]test.Msg, 0)}
		data.MessageSender = &tsn
		data.DecodeCh = dc
		data.AudioConvertCh = ac
		data.DiarizationCh = diac
		data.TranscriptionCh = tc
		data.ResultMakeCh = rc
		go StartWorkerService(&data)
		Convey("When wrong Decode msg is put", func() {
			dc <- amqp.Delivery{}
			close(dc)
			Convey("Status must not be changed", func() {
				So(cap(ts.statuses), ShouldEqual, 0)
			})
			Convey("No msg sent", func() {
				So(cap(tsn.Msgs), ShouldEqual, 0)
			})
		})
		Convey("When good Decode msg is put", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
			dc <- amqp.Delivery{Body: msgdata}
			close(dc)
			Convey("Status must be changed", func() {
				So(test.Contains(ts.statuses, messages.AudioConvert), ShouldBeTrue)
			})
			Convey("AudioConvert msg sent", func() {
				So(test.ContainsMsg(tsn.Msgs, test.NewMsg("1", messages.AudioConvert, true)), ShouldBeTrue)
			})
			Convey("StartedDecode msg sent", func() {
				So(test.ContainsMsg(tsn.Msgs, test.NewMsg("1", messages.StartedDecode, false)), ShouldBeTrue)
			})
		})
		Convey("When wrong AudioConvertResult msg is put", func() {
			ac <- amqp.Delivery{}
			close(ac)
			Convey("Status must not be changed", func() {
				So(cap(ts.statuses), ShouldEqual, 0)
			})
			Convey("No msg sent", func() {
				So(cap(tsn.Msgs), ShouldEqual, 0)
			})
		})
		Convey("When good AudioConvertResult msg is put", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
			ac <- amqp.Delivery{Body: msgdata}
			close(ac)
			Convey("Status must be changed", func() {
				So(test.Contains(ts.statuses, messages.Diarization), ShouldBeTrue)
			})
			Convey("Diarization msg sent", func() {
				So(test.ContainsMsg(tsn.Msgs, test.NewMsg("1", messages.Diarization, true)), ShouldBeTrue)
			})
		})
		Convey("When good AudioConvertResult msg with error is put", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMsgWithError("1", "error"))
			ac <- amqp.Delivery{Body: msgdata}
			close(ac)
			Convey("Status must be changed", func() {
				So(test.Contains(ts.statuses, messages.AudioConvert+"error"), ShouldBeTrue)
			})
			Convey("No msg sent", func() {
				So(cap(tsn.Msgs), ShouldEqual, 0)
			})
		})
		Convey("When wrong DiarizationResult msg is put", func() {
			diac <- amqp.Delivery{}
			close(diac)
			Convey("Status must not be changed", func() {
				So(cap(ts.statuses), ShouldEqual, 0)
			})
			Convey("No msg sent", func() {
				So(cap(tsn.Msgs), ShouldEqual, 0)
			})
		})
		Convey("When good DiarizationResult msg is put", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
			diac <- amqp.Delivery{Body: msgdata}
			close(diac)
			Convey("Status must be changed", func() {
				So(test.Contains(ts.statuses, messages.Transcription), ShouldBeTrue)
			})
			Convey("Transcription msg sent", func() {
				So(test.ContainsMsg(tsn.Msgs, test.NewMsg("1", messages.Transcription, true)), ShouldBeTrue)
			})
		})
		Convey("When good DiarizationResult msg with error is put", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMsgWithError("1", "error"))
			diac <- amqp.Delivery{Body: msgdata}
			close(diac)
			Convey("Status must be changed", func() {
				So(test.Contains(ts.statuses, messages.Diarization+"error"), ShouldBeTrue)
			})
			Convey("No msg sent", func() {
				So(cap(tsn.Msgs), ShouldEqual, 0)
			})
		})
		Convey("When good TranscriptionResult msg is put", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
			tc <- amqp.Delivery{Body: msgdata}
			close(tc)
			time.Sleep(time.Second * 2)
			Convey("Status must be changed", func() {
				So(test.Contains(ts.statuses, messages.ResultMake), ShouldBeTrue)
			})
			Convey("Transcription msg sent", func() {
				So(test.ContainsMsg(tsn.Msgs, test.NewMsg("1", messages.ResultMake, true)), ShouldBeTrue)
			})
		})
		Convey("When good TranscriptionResult msg with error is put", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMsgWithError("1", "error"))
			tc <- amqp.Delivery{Body: msgdata}
			close(tc)
			time.Sleep(time.Second * 2)
			Convey("Status must be changed", func() {
				So(test.Contains(ts.statuses, messages.Transcription+"error"), ShouldBeTrue)
			})
			Convey("No msg sent", func() {
				So(cap(tsn.Msgs), ShouldEqual, 0)
			})
		})
		Convey("When good ResultMakeResult msg is put", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
			rc <- amqp.Delivery{Body: msgdata}
			close(rc)
			time.Sleep(time.Second * 2)
			Convey("Status must be changed", func() {
				So(test.Contains(ts.statuses, "COMPLETED"), ShouldBeTrue)
			})
			Convey("FinishDecode msg sent", func() {
				So(test.ContainsMsg(tsn.Msgs, test.NewMsg("1", messages.FinishDecode, false)), ShouldBeTrue)
			})
		})
		Convey("When good ResultMakeResult msg with error is put", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMsgWithError("1", "error"))
			rc <- amqp.Delivery{Body: msgdata}
			close(rc)
			time.Sleep(time.Second * 2)
			Convey("Status must be changed", func() {
				So(test.Contains(ts.statuses, messages.ResultMake+"error"), ShouldBeTrue)
			})
			Convey("No msg sent", func() {
				So(cap(tsn.Msgs), ShouldEqual, 0)
			})
		})
	})
}

type testSenderFunc func(m *messages.QueueMessage, q string, rq string) error

func (f testSenderFunc) Send(m *messages.QueueMessage, q string, rq string) error {
	return f(m, q, rq)
}

type testSaverFunc func(name string, reader io.Reader) error

func (f testSaverFunc) Save(name string, reader io.Reader) error {
	return f(name, reader)
}

type testSaver struct{}

func (saver testSaver) Save(name string, reader io.Reader) error {
	log.Printf("Saving file %s\n", name)
	return nil
}

type testStatusSaverFunc func(ID string, status string, errorStr string) error

func (f testStatusSaverFunc) Save(ID string, status string, errorStr string) error {
	return f(ID, status, errorStr)
}

type testStatusSaver struct {
	statuses []string
}

func (saver *testStatusSaver) Save(ID string, status string, errorStr string) error {
	log.Printf("Saving status %s %s\n", ID, status)
	saver.statuses = append(saver.statuses, status+errorStr)
	return nil
}
