package zoom

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/streadway/amqp"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"
)

var statusSaverMock *mocks.MockSaver
var resultSaverMock *mocks.MockResultSaver
var publisherMock *mocks.MockPublisher
var msgSenderMock *mocks.MockSender
var msgInformSenderMock *mocks.MockSender
var getterMock *mocks.MockFilesGetter
var lenMock *mocks.MockAudioDuration
var loaderMock *mocks.MockFileLoader
var saverMock *mocks.MockFileSaver
var requestSaverMock *mocks.MockRequestSaver
var fileMock *mocks.MockFile

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	statusSaverMock = mocks.NewMockSaver()
	resultSaverMock = mocks.NewMockResultSaver()
	publisherMock = mocks.NewMockPublisher()
	msgSenderMock = mocks.NewMockSender()
	msgInformSenderMock = mocks.NewMockSender()
	getterMock = mocks.NewMockFilesGetter()
	lenMock = mocks.NewMockAudioDuration()
	loaderMock = mocks.NewMockFileLoader()
	saverMock = mocks.NewMockFileSaver()
	requestSaverMock = mocks.NewMockRequestSaver()
	fileMock = mocks.NewMockFile()
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

func TestInitManager_Fails(t *testing.T) {
	data := newTestServiceData(t)
	data.AudioLen = nil
	err := StartWorkerService(data)
	assert.NotNil(t, err)

	data = newTestServiceData(t)
	data.FileSaver = nil
	assert.NotNil(t, StartWorkerService(data))

	data = newTestServiceData(t)
	data.FilesGetter = nil
	assert.NotNil(t, StartWorkerService(data))

	data = newTestServiceData(t)
	data.RequestSaver = nil
	assert.NotNil(t, StartWorkerService(data))

	data = newTestServiceData(t)
	data.StatusSaver = nil
	assert.NotNil(t, StartWorkerService(data))
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

func newTestServiceData(t *testing.T) *ServiceData {
	initTest(t)
	res := &ServiceData{}
	res.StatusSaver = statusSaverMock
	res.MessageSender = msgSenderMock
	res.InformMessageSender = msgInformSenderMock
	res.ResultSaver = resultSaverMock
	res.Publisher = publisherMock
	res.FilesGetter = getterMock
	res.AudioLen = lenMock
	res.Loader = loaderMock
	res.FileSaver = saverMock
	res.RequestSaver = requestSaverMock
	return res
}

func initTestData(t *testing.T) *testdata {
	res := testdata{}
	res.data = newTestServiceData(t)

	res.dc = make(chan amqp.Delivery)
	res.ac = make(chan amqp.Delivery)
	res.diac = make(chan amqp.Delivery)
	res.tc = make(chan amqp.Delivery)
	res.rescCh = make(chan amqp.Delivery)
	res.rc = make(chan amqp.Delivery)

	res.data.DecodeMultiCh = res.dc
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

	pegomock.When(getterMock.List(pegomock.AnyString())).ThenReturn([]string{"1.mp4", "2.mp4"}, nil)
	pegomock.When(loaderMock.Load(pegomock.AnyString())).ThenReturn(fileMock, nil)
	pegomock.When(lenMock.Get(pegomock.AnyString(), matchers.AnyIoReader())).ThenReturn(time.Second, nil)
	msgdata, _ := json.Marshal(newTestMsg())
	td.dc <- amqp.Delivery{Body: msgdata}
	close(td.dc)
	<-td.fc
	verifySendInform(t, messages.InformType_Started, 1)
	verifySendMessage(t, messages.JoinAudio, 1)
	getterMock.VerifyWasCalled(pegomock.Once()).List(pegomock.AnyString())
	loaderMock.VerifyWasCalled(pegomock.Times(4)).Load(pegomock.AnyString())
	verifySendMessage(t, messages.Decode, 2)
}

func TestHandlesMessagesDecodeMsg_FailLen(t *testing.T) {
	td := initTestData(t)

	pegomock.When(getterMock.List(pegomock.AnyString())).ThenReturn([]string{"1.mp4", "2.mp4"}, nil)
	pegomock.When(loaderMock.Load(pegomock.AnyString())).ThenReturn(fileMock, nil)
	pegomock.When(lenMock.Get(pegomock.EqString("1.mp4"), matchers.AnyIoReader())).ThenReturn(time.Second, nil)
	pegomock.When(lenMock.Get(pegomock.EqString("2.mp4"), matchers.AnyIoReader())).ThenReturn(time.Second*2, nil)
	msgdata, _ := json.Marshal(newTestMsg())
	td.dc <- amqp.Delivery{Body: msgdata}
	close(td.dc)
	<-td.fc
	verifySendInform(t, messages.InformType_Failed, 1)
	verifySendMessage(t, messages.JoinAudio, 0)
	verifySendMessage(t, messages.Decode, 0)
}

func TestHandlesMessagesDecodeMsg_FailLenError(t *testing.T) {
	td := initTestData(t)
	pegomock.When(getterMock.List(pegomock.AnyString())).ThenReturn([]string{"1.mp4", "2.mp4"}, nil)
	pegomock.When(loaderMock.Load(pegomock.AnyString())).ThenReturn(fileMock, nil)
	pegomock.When(lenMock.Get(pegomock.AnyString(), matchers.AnyIoReader())).ThenReturn(time.Second, errors.New("err"))
	msgdata, _ := json.Marshal(newTestMsg())
	td.dc <- amqp.Delivery{Body: msgdata}
	close(td.dc)
	<-td.fc
	verifySendInform(t, messages.InformType_Failed, 0)
	verifySendMessage(t, messages.JoinAudio, 0)
	verifySendMessage(t, messages.Decode, 0)
}

func TestHandlesMessagesDecodeMsg_FailGet(t *testing.T) {
	td := initTestData(t)
	pegomock.When(getterMock.List(pegomock.AnyString())).ThenReturn(nil, errors.New("err"))
	pegomock.When(loaderMock.Load(pegomock.AnyString())).ThenReturn(fileMock, nil)
	pegomock.When(lenMock.Get(pegomock.AnyString(), matchers.AnyIoReader())).ThenReturn(time.Second, nil)
	msgdata, _ := json.Marshal(newTestMsg())
	td.dc <- amqp.Delivery{Body: msgdata}
	close(td.dc)
	<-td.fc
	verifySendInform(t, messages.InformType_Failed, 0)
	verifySendMessage(t, messages.JoinAudio, 0)
	verifySendMessage(t, messages.Decode, 0)
}

func verifySendMessage(t *testing.T, mType string, count int) {
	msgSenderMock.VerifyWasCalled(pegomock.Times(count)).Send(matchers.AnyMessagesMessage(), pegomock.EqString(mType), pegomock.AnyString())
	if count > 0 {
		dm, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Times(count)).Send(matchers.AnyMessagesMessage(), pegomock.EqString(mType), pegomock.AnyString()).
			GetCapturedArguments()

		m1 := dm.(*messages.QueueMessage)
		assert.Equal(t, "rec", m1.Recognizer)
	}
}

func verifySendInform(t *testing.T, tp string, count int) {
	msgInformSenderMock.VerifyWasCalled(pegomock.Times(count)).Send(matchers.AnyMessagesMessage(), pegomock.EqString(messages.Inform), pegomock.AnyString())
	if count > 0 {
		dm, _, _ := msgInformSenderMock.VerifyWasCalled(pegomock.Times(count)).Send(matchers.AnyMessagesMessage(), pegomock.EqString(messages.Inform), pegomock.AnyString()).
			GetCapturedArguments()
		m1 := dm.(*messages.InformMessage)
		assert.Equal(t, "rec", m1.Recognizer)
		assert.Equal(t, tp, m1.Type)
	}
}

func newTestMsg() *messages.QueueMessage {
	return &messages.QueueMessage{ID: "1", Recognizer: "rec"}
}
