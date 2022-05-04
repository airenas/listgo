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
var speechIndicatorMock *mocks.MockSpeechIndicator

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	statusSaverMock = mocks.NewMockSaver()
	resultSaverMock = mocks.NewMockResultSaver()
	publisherMock = mocks.NewMockPublisher()
	msgSenderMock = mocks.NewMockSender()
	msgInformSenderMock = mocks.NewMockSender()
	speechIndicatorMock = mocks.NewMockSpeechIndicator()
}

func TestInitManagerNoResultSaver(t *testing.T) {
	data := newTestServiceData(t)
	data.ResultSaver = nil
	err := StartWorkerService(data)
	assert.NotNil(t, err)
}

func TestInitManagerOK(t *testing.T) {
	data := newTestServiceData(t)
	err := StartWorkerService(data)
	assert.Nil(t, err)
}

func TestInitManagerNoPublisher(t *testing.T) {
	data := newTestServiceData(t)
	data.Publisher = nil
	err := StartWorkerService(data)
	assert.NotNil(t, err)
}

func TestInitManagerNoSender(t *testing.T) {
	data := newTestServiceData(t)
	data.MessageSender = nil
	err := StartWorkerService(data)
	assert.NotNil(t, err)
}

func TestInitManagerNoInformSender(t *testing.T) {
	data := newTestServiceData(t)
	data.InformMessageSender = nil
	err := StartWorkerService(data)
	assert.NotNil(t, err)
}

func TestInitManagerNoSpeechIndicator(t *testing.T) {
	data := newTestServiceData(t)
	data.speechIndicator = nil
	err := StartWorkerService(data)
	assert.NotNil(t, err)
}

type testdata struct {
	dc     chan amqp.Delivery
	ac     chan amqp.Delivery
	splitc chan amqp.Delivery
	diac   chan amqp.Delivery
	tc     chan amqp.Delivery
	rescCh chan amqp.Delivery
	rc     chan amqp.Delivery
	data   *ServiceData
	fc     <-chan os.Signal
}

func newTestServiceData(t *testing.T) *ServiceData {
	initTest(t)
	res := &ServiceData{}
	res.StatusSaver = statusSaverMock
	res.MessageSender = msgSenderMock
	res.InformMessageSender = msgInformSenderMock
	res.ResultSaver = resultSaverMock
	res.Publisher = publisherMock
	res.speechIndicator = speechIndicatorMock
	return res
}

func initTestData(t *testing.T) *testdata {
	res := testdata{}
	res.data = newTestServiceData(t)

	res.dc = make(chan amqp.Delivery)
	res.ac = make(chan amqp.Delivery)
	res.splitc = make(chan amqp.Delivery)
	res.diac = make(chan amqp.Delivery)
	res.tc = make(chan amqp.Delivery)
	res.rescCh = make(chan amqp.Delivery)
	res.rc = make(chan amqp.Delivery)

	res.data.DecodeCh = res.dc
	res.data.AudioConvertCh = res.ac
	res.data.SplitChannelsCh = res.splitc
	res.data.DiarizationCh = res.diac
	res.data.TranscriptionCh = res.tc
	res.data.RescoreCh = res.rescCh
	res.data.ResultMakeCh = res.rc
	res.data.fc = utils.NewMultiCloseChannel()

	res.fc = res.data.fc.C
	err := StartWorkerService(res.data)
	assert.Nil(t, err)
	if err != nil {
		return nil
	}
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
	verifySendInformOnce(t, messages.InformTypeStarted)
	verifySendMessageOnce(t, messages.AudioConvert)
}

func TestHandlesMessagesDecodeMsg_SplitChannels(t *testing.T) {
	td := initTestData(t)
	msg := newTestMsg()
	msg.Tags = append(msg.Tags, messages.NewTag(messages.TagSepSpeakersOnChannel, "1"))
	msgdata, _ := json.Marshal(msg)
	td.dc <- amqp.Delivery{Body: msgdata}
	close(td.dc)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.SplitChannels))
	verifySendInformOnce(t, messages.InformTypeStarted)
	verifySendMessageOnce(t, messages.SplitChannels)
}

func verifySendMessageOnce(t *testing.T, mType string) {
	t.Helper()
	dm, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(), pegomock.EqString(mType), pegomock.AnyString()).
		GetCapturedArguments()
	m1 := dm.(*messages.QueueMessage)
	assert.Equal(t, "rec", m1.Recognizer)
}

func verifySendInformOnce(t *testing.T, tp string) {
	t.Helper()
	dm, _, _ := msgInformSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(), pegomock.EqString(messages.Inform), pegomock.AnyString()).
		GetCapturedArguments()
	m1 := dm.(*messages.InformMessage)
	assert.Equal(t, "rec", m1.Recognizer)
	assert.Equal(t, tp, m1.Type)
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

func TestHandlesMessagesSplitChannelsMsg(t *testing.T) {
	td := initTestData(t)

	msgdata, _ := json.Marshal(newTestMsg())
	td.splitc <- amqp.Delivery{Body: msgdata}
	close(td.splitc)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.AudioConvert))
	verifySendMessageOnce(t, messages.DecodeMultiple)
}

func TestHandlesMessagesSplitChannelsMsg_Fail(t *testing.T) {
	td := initTestData(t)

	td.splitc <- amqp.Delivery{}
	close(td.splitc)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
	msgSenderMock.VerifyWasCalled(pegomock.Never()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(), pegomock.AnyString())
}

func TestHandlesMessagesAudioConvertMsgWithTargetQueue(t *testing.T) {
	td := initTestData(t)
	msg := newTestMsg()
	msg.Tags = append(msg.Tags, messages.NewTag(messages.TagStatusQueue, "Q1"))
	msgdata, _ := json.Marshal(msg)
	td.ac <- amqp.Delivery{Body: msgdata}
	close(td.ac)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.Diarization))
	verifySendMessageOnce(t, messages.Diarization)
	verifySendMessageOnce(t, "Q1")
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
	verifySendInformOnce(t, messages.InformTypeFailed)

}

func TestHandlesMessagesAudioConvertWithErrorAndTargetQueue(t *testing.T) {
	td := initTestData(t)
	msg := newTestMsgError()
	msg.Tags = append(msg.Tags, messages.NewTag(messages.TagResultQueue, "Q1"))
	msgdata, _ := json.Marshal(msg)
	td.ac <- amqp.Delivery{Body: msgdata}
	close(td.ac)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).SaveError(pegomock.AnyString(),
		pegomock.EqString("error"))
	msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(), pegomock.EqString("Q1"), pegomock.AnyString())
	verifySendInformOnce(t, messages.InformTypeFailed)

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

	pegomock.When(speechIndicatorMock.Test(pegomock.AnyString())).ThenReturn(true, nil)
	msgdata, _ := json.Marshal(newTestMsg())
	td.diac <- amqp.Delivery{Body: msgdata}
	close(td.diac)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.Transcription))
	speechIndicatorMock.VerifyWasCalled(pegomock.Times(1)).Test(pegomock.AnyString())
	verifySendMessageOnce(t, messages.Transcription)
}

func TestHandlesMessagesDiarizationMsg_FailsSpeechIndicator(t *testing.T) {
	td := initTestData(t)

	pegomock.When(speechIndicatorMock.Test(pegomock.AnyString())).ThenReturn(true, errors.New("error"))
	msgdata, _ := json.Marshal(newTestMsg())
	td.diac <- amqp.Delivery{Body: msgdata}
	close(td.diac)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.Transcription))
	speechIndicatorMock.VerifyWasCalled(pegomock.Times(1)).Test(pegomock.AnyString())
	verifySendMessageOnce(t, messages.Transcription)
}

func TestHandlesMessagesDiarizationMsgNoSpeech(t *testing.T) {
	td := initTestData(t)

	pegomock.When(speechIndicatorMock.Test(pegomock.AnyString())).ThenReturn(false, nil)
	msgdata, _ := json.Marshal(newTestMsg())
	td.diac <- amqp.Delivery{Body: msgdata}
	close(td.diac)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.ResultMake))
	speechIndicatorMock.VerifyWasCalled(pegomock.Times(1)).Test(pegomock.AnyString())
	verifySendMessageOnce(t, messages.ResultMake)
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
	verifySendInformOnce(t, messages.InformTypeFailed)
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
	verifySendInformOnce(t, messages.InformTypeFailed)
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
	verifySendInformOnce(t, messages.InformTypeFailed)
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
	verifySendInformOnce(t, messages.InformTypeFinished)
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
	verifySendInformOnce(t, messages.InformTypeFailed)
}

func TestHandlesMessagesResultMakeMsgWithTarget(t *testing.T) {
	td := initTestData(t)

	msg := messages.ResultMessage{QueueMessage: *newTestMsg(), Result: "result"}
	msg.Tags = append(msg.Tags, messages.NewTag(messages.TagResultQueue, "Q1"))
	msgdata, _ := json.Marshal(msg)
	td.rc <- amqp.Delivery{Body: msgdata}
	close(td.rc)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), matchers.EqStatusStatus(status.Completed))
	verifySendInformOnce(t, messages.InformTypeFinished)
	resultSaverMock.VerifyWasCalled(pegomock.Times(1)).Save(pegomock.AnyString(), pegomock.AnyString())
	verifySendMessageOnce(t, "Q1")
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
