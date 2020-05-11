package manager

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"testing"

	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/streadway/amqp"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/status"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"
)

var statusSaverMock *mocks.MockSaver
var resultSaverMock *mocks.MockResultSaver
var publisherMock *mocks.MockPublisher
var msgSenderMock *mocks.MockSender
var msgInformSenderMock *mocks.MockSender

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	statusSaverMock = mocks.NewMockSaver()
	resultSaverMock = mocks.NewMockResultSaver()
	publisherMock = mocks.NewMockPublisher()
	msgSenderMock = mocks.NewMockSender()
	msgInformSenderMock = mocks.NewMockSender()
}

func TestInitManagerNoResultSaver(t *testing.T) {
	initTest(t)
	data := ServiceData{}
	data.Publisher = publisherMock
	err := StartWorkerService(&data)
	assert.NotNil(t, err)
}

func TestInitManagerOK(t *testing.T) {
	initTest(t)
	data := ServiceData{}
	data.ResultSaver = resultSaverMock
	data.Publisher = publisherMock
	data.MessageSender = msgSenderMock
	data.InformMessageSender = msgInformSenderMock

	err := StartWorkerService(&data)
	assert.Nil(t, err)
}

func TestInitManagerNoPublisher(t *testing.T) {
	initTest(t)
	data := ServiceData{}
	data.ResultSaver = resultSaverMock
	data.MessageSender = msgSenderMock
	data.InformMessageSender = msgSenderMock
	err := StartWorkerService(&data)
	assert.NotNil(t, err)
}

func TestInitManagerNoSender(t *testing.T) {
	initTest(t)
	data := ServiceData{}
	data.ResultSaver = resultSaverMock
	data.Publisher = publisherMock
	data.InformMessageSender = msgSenderMock
	err := StartWorkerService(&data)
	assert.NotNil(t, err)
}

func TestInitManagerNoInformSender(t *testing.T) {
	initTest(t)
	data := ServiceData{}
	data.ResultSaver = resultSaverMock
	data.Publisher = publisherMock
	data.MessageSender = msgSenderMock
	err := StartWorkerService(&data)
	assert.NotNil(t, err)
}

type testdata struct {
	dc     chan amqp.Delivery
	ac     chan amqp.Delivery
	diac   chan amqp.Delivery
	tc     chan amqp.Delivery
	rescCh chan amqp.Delivery
	rc     chan amqp.Delivery
	data   *ServiceData
	fc     <-chan os.Signal
}

func initTestData(t *testing.T) *testdata {
	initTest(t)
	res := testdata{}
	res.dc = make(chan amqp.Delivery)
	res.ac = make(chan amqp.Delivery)
	res.diac = make(chan amqp.Delivery)
	res.tc = make(chan amqp.Delivery)
	res.rescCh = make(chan amqp.Delivery)
	res.rc = make(chan amqp.Delivery)
	res.data = &ServiceData{}
	res.data.StatusSaver = statusSaverMock
	res.data.MessageSender = msgSenderMock
	res.data.InformMessageSender = msgInformSenderMock
	res.data.DecodeCh = res.dc
	res.data.AudioConvertCh = res.ac
	res.data.DiarizationCh = res.diac
	res.data.TranscriptionCh = res.tc
	res.data.RescoreCh = res.rescCh
	res.data.ResultMakeCh = res.rc
	res.data.ResultSaver = resultSaverMock
	res.data.Publisher = publisherMock
	res.data.fc = utils.NewMultiCloseChannel()
	res.fc = res.data.fc.C
	err := StartWorkerService(res.data)
	assert.Nil(t, err)
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

	msgdata, _ := json.Marshal(newTestMsg())
	td.dc <- amqp.Delivery{Body: msgdata}
	close(td.dc)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.AudioConvert))
	verifySendInformOnce(t, messages.InformType_Started)
	verifySendMessageOnce(t, messages.AudioConvert)
}

func verifySendMessageOnce(t *testing.T, mType string) {
	dm, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(), pegomock.EqString(mType), pegomock.AnyString()).
		GetCapturedArguments()
	m1 := dm.(*messages.QueueMessage)
	assert.Equal(t, "rec", m1.Recognizer)
}

func verifySendInformOnce(t *testing.T, tp string) {
	dm, _, _ := msgInformSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(), pegomock.EqString(messages.Inform), pegomock.AnyString()).
		GetCapturedArguments()
	m1 := dm.(*messages.InformMessage)
	assert.Equal(t, "rec", m1.Recognizer)
	assert.Equal(t, tp, m1.Type)
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

	msgdata, _ := json.Marshal(newTestMsg())
	td.ac <- amqp.Delivery{Body: msgdata}
	close(td.ac)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.Diarization))
	verifySendMessageOnce(t, messages.Diarization)
}

func TestHandlesMessagesAudioConvertWithError(t *testing.T) {
	td := initTestData(t)

	msgdata, _ := json.Marshal(newTestMsgError())
	td.ac <- amqp.Delivery{Body: msgdata}
	close(td.ac)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).SaveError(pegomock.AnyString(),
		pegomock.EqString("error"))
	msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
	verifySendInformOnce(t, messages.InformType_Failed)
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

	msgdata, _ := json.Marshal(newTestMsg())
	td.diac <- amqp.Delivery{Body: msgdata}
	close(td.diac)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.Transcription))
	verifySendMessageOnce(t, messages.Transcription)
}

func TestHandlesMessagesDiarizationWithError(t *testing.T) {
	td := initTestData(t)

	msgdata, _ := json.Marshal(newTestMsgError())
	td.diac <- amqp.Delivery{Body: msgdata}
	close(td.diac)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).SaveError(pegomock.AnyString(),
		pegomock.EqString("error"))
	msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
	verifySendInformOnce(t, messages.InformType_Failed)
}

func TestHandlesMessagesTranscriptionMsg(t *testing.T) {
	td := initTestData(t)

	msgdata, _ := json.Marshal(newTestMsg())
	td.tc <- amqp.Delivery{Body: msgdata}
	close(td.tc)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.Rescore))
	verifySendMessageOnce(t, messages.Rescore)
}

func TestHandlesMessagesTranscriptionWithError(t *testing.T) {
	td := initTestData(t)

	msgdata, _ := json.Marshal(newTestMsgError())
	td.tc <- amqp.Delivery{Body: msgdata}
	close(td.tc)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).SaveError(pegomock.AnyString(),
		pegomock.EqString("error"))
	msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
	verifySendInformOnce(t, messages.InformType_Failed)
}

func TestHandlesMessagesRescoreMsg(t *testing.T) {
	td := initTestData(t)

	msgdata, _ := json.Marshal(newTestMsg())
	td.rescCh <- amqp.Delivery{Body: msgdata}
	close(td.rescCh)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.ResultMake))
	verifySendMessageOnce(t, messages.ResultMake)
}

func TestHandlesMessagesRescoreWithError(t *testing.T) {
	td := initTestData(t)

	msgdata, _ := json.Marshal(newTestMsgError())
	td.rescCh <- amqp.Delivery{Body: msgdata}
	close(td.rescCh)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).SaveError(pegomock.AnyString(),
		pegomock.EqString("error"))
	msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
	verifySendInformOnce(t, messages.InformType_Failed)
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

	msg := messages.ResultMessage{QueueMessage: *newTestMsg(), Result: "result"}
	msgdata, _ := json.Marshal(msg)
	td.rc <- amqp.Delivery{Body: msgdata}
	close(td.rc)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.Completed))
	verifySendInformOnce(t, messages.InformType_Finished)
	resultSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), pegomock.AnyString())
}

func TestHandlesMessagesResultMakeWithError(t *testing.T) {
	td := initTestData(t)

	msg := newTestMsgError()
	msgdata, _ := json.Marshal(msg)
	td.rc <- amqp.Delivery{Body: msgdata}
	close(td.rc)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).SaveError(pegomock.AnyString(),
		pegomock.EqString("error"))
	msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
	resultSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), pegomock.AnyString())
	verifySendInformOnce(t, messages.InformType_Failed)
}

func newTestMsg() *messages.QueueMessage {
	return &messages.QueueMessage{ID: "1", Recognizer: "rec"}
}

func newTestMsgError() *messages.QueueMessage {
	res := messages.NewQueueMsgWithError("1", "error")
	res.Recognizer = "rec"
	return res
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
