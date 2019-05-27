package manager

import (
	"encoding/json"
	"io"
	"log"
	"testing"

	"github.com/petergtz/pegomock"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/status"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"
	. "github.com/smartystreets/goconvey/convey"
)

var statusSaverMock *mocks.MockSaver
var resultSaverMock *mocks.MockResultSaver
var publisherMock *mocks.MockPublisher
var msgSenderMock *mocks.MockSender

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	statusSaverMock = mocks.NewMockSaver()
	resultSaverMock = mocks.NewMockResultSaver()
	publisherMock = mocks.NewMockPublisher()
	msgSenderMock = mocks.NewMockSender()
}

func TestInitManager(t *testing.T) {
	Convey("Given a manager", t, func() {
		initTest(t)
		data := ServiceData{}
		data.ResultSaver = resultSaverMock
		data.Publisher = publisherMock
		Convey("When no result Saver", func() {
			data.ResultSaver = nil
			_, err := StartWorkerService(&data)
			So(err, ShouldNotBeNil)
		})
		Convey("When ResultSaver Provided", func() {
			_, err := StartWorkerService(&data)
			So(err, ShouldBeNil)
		})
		Convey("When no Publisher", func() {
			data.Publisher = nil
			_, err := StartWorkerService(&data)
			So(err, ShouldNotBeNil)
		})
		Convey("When Publisher Provided", func() {
			_, err := StartWorkerService(&data)
			So(err, ShouldBeNil)
		})
	})
}

func TestHandlesMessages(t *testing.T) {
	Convey("Given a manager", t, func() {
		initTest(t)
		dc := make(chan amqp.Delivery)
		ac := make(chan amqp.Delivery)
		diac := make(chan amqp.Delivery)
		tc := make(chan amqp.Delivery)
		rc := make(chan amqp.Delivery)
		data := ServiceData{}
		data.StatusSaver = statusSaverMock
		data.MessageSender = msgSenderMock
		data.DecodeCh = dc
		data.AudioConvertCh = ac
		data.DiarizationCh = diac
		data.TranscriptionCh = tc
		data.ResultMakeCh = rc
		data.ResultSaver = resultSaverMock
		data.Publisher = publisherMock
		fc, _ := StartWorkerService(&data)
		Convey("When wrong Decode msg is put", func() {
			dc <- amqp.Delivery{}
			close(dc)
			<-fc
			Convey("Status must not be changed", func() {
				statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
			})
			Convey("No msg sent", func() {
				msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
			})
		})
		Convey("When good Decode msg is put", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
			dc <- amqp.Delivery{Body: msgdata}
			close(dc)
			<-fc
			Convey("Status must be changed", func() {
				statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.AudioConvert))
			})
			Convey("AudioConvert msg sent", func() {
				msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
					pegomock.EqString(messages.AudioConvert), pegomock.AnyString())
			})
			Convey("Inform msg sent", func() {
				msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
					pegomock.EqString(messages.Inform), pegomock.AnyString())
			})
		})
		Convey("When wrong AudioConvertResult msg is put", func() {
			ac <- amqp.Delivery{}
			close(ac)
			<-fc
			Convey("Status must not be changed", func() {
				statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
			})
			Convey("No msg sent", func() {
				msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
			})
		})
		Convey("When good AudioConvertResult msg is put", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
			ac <- amqp.Delivery{Body: msgdata}
			close(ac)
			<-fc
			Convey("Status must be changed", func() {
				statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.Diarization))
			})
			Convey("Diarization msg sent", func() {
				msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
					pegomock.EqString(messages.Diarization), pegomock.AnyString())
			})
		})
		Convey("When good AudioConvertResult msg with error is put", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMsgWithError("1", "error"))
			ac <- amqp.Delivery{Body: msgdata}
			close(ac)
			<-fc
			Convey("Status must be changed", func() {
				statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
				statusSaverMock.VerifyWasCalled(pegomock.Times(1)).SaveError(pegomock.AnyString(),
					pegomock.EqString("error"))
			})
			Convey("No msg sent", func() {
				msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
			})
		})
		Convey("When wrong DiarizationResult msg is put", func() {
			diac <- amqp.Delivery{}
			close(diac)
			Convey("Status must not be changed", func() {
				statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
			})
			Convey("No msg sent", func() {
				msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
			})
		})
		Convey("When good DiarizationResult msg is put", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
			diac <- amqp.Delivery{Body: msgdata}
			close(diac)
			<-fc
			Convey("Status must be changed", func() {
				statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.Transcription))
			})
			Convey("Transcription msg sent", func() {
				msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
					pegomock.EqString(messages.Transcription), pegomock.AnyString())
			})
		})
		Convey("When good DiarizationResult msg with error is put", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMsgWithError("1", "error"))
			diac <- amqp.Delivery{Body: msgdata}
			close(diac)
			<-fc
			Convey("Status must be changed", func() {
				statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
				statusSaverMock.VerifyWasCalled(pegomock.Times(1)).SaveError(pegomock.AnyString(),
					pegomock.EqString("error"))
			})
			Convey("No msg sent", func() {
				msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
			})
		})
		Convey("When good TranscriptionResult msg is put", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
			tc <- amqp.Delivery{Body: msgdata}
			close(tc)
			<-fc
			Convey("Status must be changed", func() {
				statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.ResultMake))
			})
			Convey("Transcription msg sent", func() {
				msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
					pegomock.EqString(messages.ResultMake), pegomock.AnyString())
			})
		})
		Convey("When good TranscriptionResult msg with error is put", func() {
			msgdata, _ := json.Marshal(messages.NewQueueMsgWithError("1", "error"))
			tc <- amqp.Delivery{Body: msgdata}
			close(tc)
			<-fc
			Convey("Status must be changed", func() {
				statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
				statusSaverMock.VerifyWasCalled(pegomock.Times(1)).SaveError(pegomock.AnyString(),
					pegomock.EqString("error"))
			})
			Convey("No msg sent", func() {
				msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
			})
		})
		Convey("When good ResultMakeResult msg is put", func() {
			msg := messages.ResultMessage{QueueMessage: messages.QueueMessage{ID: "1"}, Result: "result"}
			msgdata, _ := json.Marshal(msg)
			rc <- amqp.Delivery{Body: msgdata}
			close(rc)
			<-fc
			Convey("Status must be changed", func() {
				statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.Completed))
			})
			Convey("Inform msg sent", func() {
				msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
					pegomock.EqString(messages.Inform), pegomock.AnyString())
			})
			Convey("result save is called", func() {
				resultSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), pegomock.AnyString())
			})
		})
		Convey("When good ResultMakeResult msg with error is put", func() {
			msg := messages.NewQueueMsgWithError("1", "error")
			msgdata, _ := json.Marshal(msg)
			rc <- amqp.Delivery{Body: msgdata}
			close(rc)
			<-fc
			Convey("Status must be changed", func() {
				statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
				statusSaverMock.VerifyWasCalled(pegomock.Times(1)).SaveError(pegomock.AnyString(),
					pegomock.EqString("error"))
			})
			Convey("No msg sent", func() {
				msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
			})
			Convey("result save is not called", func() {
				resultSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), pegomock.AnyString())
			})
		})
		Convey("When good ResultMakeResult msg is put and", func() {
			Convey("Result save fails", func() {
				pegomock.When(resultSaverMock.Save(pegomock.AnyString(), pegomock.AnyString())).ThenReturn(errors.New("Fail"))
				msgdata, _ := json.Marshal(messages.NewQueueMsgWithError("1", ""))
				rc <- amqp.Delivery{Body: msgdata}
				close(rc)
				<-fc
				Convey("Status must not be changed", func() {
					statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
				})
				Convey("No msg sent", func() {
					msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
				})
				Convey("result save is called", func() {
					resultSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), pegomock.AnyString())
				})
			})

		})
	})
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
