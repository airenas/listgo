package manager

import (
	"encoding/json"
	"io"
	"log"
	"testing"

	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/streadway/amqp"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/status"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"
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

func TestInitManagerNoResultSaver(t *testing.T) {
	initTest(t)
	data := ServiceData{}
	data.Publisher = publisherMock
	_, err := StartWorkerService(&data)
	assert.NotNil(t, err)
}

func TestInitManagerOK(t *testing.T) {
	initTest(t)
	data := ServiceData{}
	data.ResultSaver = resultSaverMock
	data.Publisher = publisherMock
	_, err := StartWorkerService(&data)
	assert.Nil(t, err)
}

func TestInitManagerNoPublisher(t *testing.T) {
	initTest(t)
	data := ServiceData{}
	data.ResultSaver = resultSaverMock
	_, err := StartWorkerService(&data)
	assert.NotNil(t, err)
}

type testdata struct {
	dc   chan amqp.Delivery
	ac   chan amqp.Delivery
	diac chan amqp.Delivery
	tc   chan amqp.Delivery
	rc   chan amqp.Delivery
	data *ServiceData
	fc   <-chan struct{}
}

func initTestData(t *testing.T) *testdata {
	initTest(t)
	res := testdata{}
	res.dc = make(chan amqp.Delivery)
	res.ac = make(chan amqp.Delivery)
	res.diac = make(chan amqp.Delivery)
	res.tc = make(chan amqp.Delivery)
	res.rc = make(chan amqp.Delivery)
	res.data = &ServiceData{}
	res.data.StatusSaver = statusSaverMock
	res.data.MessageSender = msgSenderMock
	res.data.DecodeCh = res.dc
	res.data.AudioConvertCh = res.ac
	res.data.DiarizationCh = res.diac
	res.data.TranscriptionCh = res.tc
	res.data.ResultMakeCh = res.rc
	res.data.ResultSaver = resultSaverMock
	res.data.Publisher = publisherMock
	res.fc, _ = StartWorkerService(res.data)
	return &res
}

func TestHandlesMessagesWrongMsg(t *testing.T) {
	td := initTestData(t)
	td.dc <- amqp.Delivery{}
	close(td.dc)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
	msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
}

func TestHandlesMessagesDecodeMsg(t *testing.T) {
	td := initTestData(t)

	msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
	td.dc <- amqp.Delivery{Body: msgdata}
	close(td.dc)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.AudioConvert))
	msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
		pegomock.EqString(messages.AudioConvert), pegomock.AnyString())
	msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
		pegomock.EqString(messages.Inform), pegomock.AnyString())
}

func TestHandlesMessagesWrongAudioConvertMsg(t *testing.T) {
	td := initTestData(t)

	td.ac <- amqp.Delivery{}
	close(td.ac)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
	msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
}

func TestHandlesMessagesAudioConvertMsg(t *testing.T) {
	td := initTestData(t)

	msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
	td.ac <- amqp.Delivery{Body: msgdata}
	close(td.ac)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.Diarization))
	msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
		pegomock.EqString(messages.Diarization), pegomock.AnyString())
}

func TestHandlesMessagesAudioConvertWithError(t *testing.T) {
	td := initTestData(t)

	msgdata, _ := json.Marshal(messages.NewQueueMsgWithError("1", "error"))
	td.ac <- amqp.Delivery{Body: msgdata}
	close(td.ac)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).SaveError(pegomock.AnyString(),
		pegomock.EqString("error"))
	msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
}

func TestHandlesMessagesWrongDiariazationMsg(t *testing.T) {
	td := initTestData(t)

	td.diac <- amqp.Delivery{}
	close(td.diac)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
	msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())

}

func TestHandlesMessagesDiarizationMsg(t *testing.T) {
	td := initTestData(t)

	msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
	td.diac <- amqp.Delivery{Body: msgdata}
	close(td.diac)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.Transcription))
	msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
		pegomock.EqString(messages.Transcription), pegomock.AnyString())
}

func TestHandlesMessagesDiarizationWithError(t *testing.T) {
	td := initTestData(t)

	msgdata, _ := json.Marshal(messages.NewQueueMsgWithError("1", "error"))
	td.diac <- amqp.Delivery{Body: msgdata}
	close(td.diac)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).SaveError(pegomock.AnyString(),
		pegomock.EqString("error"))
	msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
}

func TestHandlesMessagesTranscriptionMsg(t *testing.T) {
	td := initTestData(t)

	msgdata, _ := json.Marshal(messages.NewQueueMessage("1"))
	td.tc <- amqp.Delivery{Body: msgdata}
	close(td.tc)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.ResultMake))
	msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
		pegomock.EqString(messages.ResultMake), pegomock.AnyString())
}

func TestHandlesMessagesTranscriptionWithError(t *testing.T) {
	td := initTestData(t)

	msgdata, _ := json.Marshal(messages.NewQueueMsgWithError("1", "error"))
	td.tc <- amqp.Delivery{Body: msgdata}
	close(td.tc)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).SaveError(pegomock.AnyString(),
		pegomock.EqString("error"))
	msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
}

func TestHandlesMessagesResultMakeMsgSaveFails(t *testing.T) {
	td := initTestData(t)

	pegomock.When(resultSaverMock.Save(pegomock.AnyString(), pegomock.AnyString())).ThenReturn(errors.New("Fail"))
	msgdata, _ := json.Marshal(messages.NewQueueMsgWithError("1", ""))
	td.rc <- amqp.Delivery{Body: msgdata}
	close(td.rc)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
	msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
	resultSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), pegomock.AnyString())
}

func TestHandlesMessagesResultMakeMsg(t *testing.T) {
	td := initTestData(t)

	msg := messages.ResultMessage{QueueMessage: messages.QueueMessage{ID: "1"}, Result: "result"}
	msgdata, _ := json.Marshal(msg)
	td.rc <- amqp.Delivery{Body: msgdata}
	close(td.rc)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.Completed))
	msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(),
		pegomock.EqString(messages.Inform), pegomock.AnyString())
	resultSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), pegomock.AnyString())
}

func TestHandlesMessagesResultMakeWithError(t *testing.T) {
	td := initTestData(t)

	msg := messages.NewQueueMsgWithError("1", "error")
	msgdata, _ := json.Marshal(msg)
	td.rc <- amqp.Delivery{Body: msgdata}
	close(td.rc)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).SaveError(pegomock.AnyString(),
		pegomock.EqString("error"))
	msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
	resultSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), pegomock.AnyString())
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
