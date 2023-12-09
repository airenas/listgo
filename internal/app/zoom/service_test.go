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

	"github.com/airenas/listgo/internal/app/status/api"
	"github.com/airenas/listgo/internal/pkg/messages"
	"github.com/airenas/listgo/internal/pkg/persistence"
	"github.com/airenas/listgo/internal/pkg/status"
	"github.com/airenas/listgo/internal/pkg/test/mocks"
	"github.com/airenas/listgo/internal/pkg/test/mocks/matchers"
	"github.com/airenas/listgo/internal/pkg/utils"
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
var dbMock *mocks.MockWorkPersistence
var statusMock *mocks.MockProvider

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
	dbMock = &mocks.MockWorkPersistence{}
	fileMock = mocks.NewMockFile()
	statusMock = mocks.NewMockProvider()

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

	data = newTestServiceData(t)
	data.DB = nil
	assert.NotNil(t, StartWorkerService(data))

	data = newTestServiceData(t)
	data.StatusProvider = nil
	assert.NotNil(t, StartWorkerService(data))
}

type testdata struct {
	decodeCh      chan amqp.Delivery
	joinAudioCh   chan amqp.Delivery
	joinResultsCh chan amqp.Delivery
	statusCh      chan amqp.Delivery
	completeCh    chan amqp.Delivery
	data          *ServiceData
	fc            <-chan os.Signal
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
	res.DB = dbMock
	res.StatusProvider = statusMock
	return res
}

func initTestData(t *testing.T) *testdata {
	res := testdata{}
	res.data = newTestServiceData(t)

	res.decodeCh = make(chan amqp.Delivery)
	res.joinAudioCh = make(chan amqp.Delivery)
	res.joinResultsCh = make(chan amqp.Delivery)
	res.statusCh = make(chan amqp.Delivery)
	res.completeCh = make(chan amqp.Delivery)

	res.data.DecodeMultiCh = res.decodeCh
	res.data.JoinAudioCh = res.joinAudioCh
	res.data.JoinResultsCh = res.joinResultsCh
	res.data.OneStatusCh = res.statusCh
	res.data.OneCompletedCh = res.completeCh

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
	td.decodeCh <- amqp.Delivery{}
	close(td.decodeCh)
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
	td.decodeCh <- amqp.Delivery{Body: msgdata}
	close(td.decodeCh)
	<-td.fc
	verifySendInform(t, messages.InformTypeStarted, 1)
	verifySendMessage(t, messages.JoinAudio, 1)
	getterMock.VerifyWasCalled(pegomock.Once()).List(pegomock.AnyString())
	loaderMock.VerifyWasCalled(pegomock.Times(4)).Load(pegomock.AnyString())
	verifySendMessage(t, messages.Decode, 2)
}

func TestHandlesMessagesDecodeMsg_SkipJoinAudion(t *testing.T) {
	td := initTestData(t)

	pegomock.When(getterMock.List(pegomock.AnyString())).ThenReturn([]string{"1.mp4", "2.mp4"}, nil)
	pegomock.When(loaderMock.Load(pegomock.AnyString())).ThenReturn(fileMock, nil)
	pegomock.When(lenMock.Get(pegomock.AnyString(), matchers.AnyIoReader())).ThenReturn(time.Second, nil)
	msg := newTestMsg()
	msg.Tags = append(msg.Tags, messages.NewTag(messages.TagSepSpeakersOnChannel, "1"), messages.NewTag("olia", "1"))
	msgdata, _ := json.Marshal(msg)
	td.decodeCh <- amqp.Delivery{Body: msgdata}
	close(td.decodeCh)
	<-td.fc
	verifySendInform(t, messages.InformTypeStarted, 1)
	verifySendMessage(t, messages.JoinAudio, 0)
	getterMock.VerifyWasCalled(pegomock.Once()).List(pegomock.AnyString())
	loaderMock.VerifyWasCalled(pegomock.Times(4)).Load(pegomock.AnyString())
	dMsg := verifySendMessage(t, messages.Decode, 2)
	if len(dMsg) > 0 {
		str, _ := messages.GetTag(dMsg[0].Tags, messages.TagSepSpeakersOnChannel)
		assert.Equal(t, "", str)
		str, _ = messages.GetTag(dMsg[0].Tags, "olia")
		assert.Equal(t, "1", str)
	}
}

func TestHandlesMessagesDecodeMsg_FailLen(t *testing.T) {
	td := initTestData(t)

	pegomock.When(getterMock.List(pegomock.AnyString())).ThenReturn([]string{"1.mp4", "2.mp4"}, nil)
	pegomock.When(loaderMock.Load(pegomock.AnyString())).ThenReturn(fileMock, nil)
	pegomock.When(lenMock.Get(pegomock.EqString("1.mp4"), matchers.AnyIoReader())).ThenReturn(time.Second, nil)
	pegomock.When(lenMock.Get(pegomock.EqString("2.mp4"), matchers.AnyIoReader())).ThenReturn(time.Second*2, nil)
	msgdata, _ := json.Marshal(newTestMsg())
	td.decodeCh <- amqp.Delivery{Body: msgdata}
	close(td.decodeCh)
	<-td.fc
	verifySendInform(t, messages.InformTypeFailed, 1)
	verifySendMessage(t, messages.JoinAudio, 0)
	verifySendMessage(t, messages.Decode, 0)
}

func TestHandlesMessagesDecodeMsg_FailLenError(t *testing.T) {
	td := initTestData(t)
	pegomock.When(getterMock.List(pegomock.AnyString())).ThenReturn([]string{"1.mp4", "2.mp4"}, nil)
	pegomock.When(loaderMock.Load(pegomock.AnyString())).ThenReturn(fileMock, nil)
	pegomock.When(lenMock.Get(pegomock.AnyString(), matchers.AnyIoReader())).ThenReturn(time.Second, errors.New("err"))
	msgdata, _ := json.Marshal(newTestMsg())
	td.decodeCh <- amqp.Delivery{Body: msgdata}
	close(td.decodeCh)
	<-td.fc
	verifySendInform(t, messages.InformTypeFailed, 0)
	verifySendMessage(t, messages.JoinAudio, 0)
	verifySendMessage(t, messages.Decode, 0)
}

func TestHandlesMessagesDecodeMsg_FailGet(t *testing.T) {
	td := initTestData(t)
	pegomock.When(getterMock.List(pegomock.AnyString())).ThenReturn(nil, errors.New("err"))
	pegomock.When(loaderMock.Load(pegomock.AnyString())).ThenReturn(fileMock, nil)
	pegomock.When(lenMock.Get(pegomock.AnyString(), matchers.AnyIoReader())).ThenReturn(time.Second, nil)
	msgdata, _ := json.Marshal(newTestMsg())
	td.decodeCh <- amqp.Delivery{Body: msgdata}
	close(td.decodeCh)
	<-td.fc
	verifySendInform(t, messages.InformTypeFailed, 0)
	verifySendMessage(t, messages.JoinAudio, 0)
	verifySendMessage(t, messages.Decode, 0)
}

func TestHandlesJoinAudio(t *testing.T) {
	td := initTestData(t)
	pegomock.When(statusMock.Get(pegomock.AnyString())).ThenReturn(&api.TranscriptionResult{}, nil)
	msgdata, _ := json.Marshal(newTestMsg())
	td.joinAudioCh <- amqp.Delivery{Body: msgdata}
	close(td.joinAudioCh)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Once()).SaveF(pegomock.AnyString(),
		matchers.AnyMapOfStringToInterface(), matchers.AnyMapOfStringToInterface())
}

func TestHandlesJoinAudio_NoSave(t *testing.T) {
	td := initTestData(t)
	pegomock.When(statusMock.Get(pegomock.AnyString())).ThenReturn(&api.TranscriptionResult{ErrorCode: "EC"}, nil)
	msgdata, _ := json.Marshal(newTestMsg())
	td.joinAudioCh <- amqp.Delivery{Body: msgdata}
	close(td.joinAudioCh)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).SaveF(pegomock.AnyString(),
		matchers.AnyMapOfStringToInterface(), matchers.AnyMapOfStringToInterface())
}

func TestHandlesJoinResults(t *testing.T) {
	td := initTestData(t)
	pegomock.When(statusMock.Get(pegomock.AnyString())).ThenReturn(&api.TranscriptionResult{}, nil)
	msgdata, _ := json.Marshal(newTestResMsg())
	td.joinResultsCh <- amqp.Delivery{Body: msgdata}
	close(td.joinResultsCh)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Once()).SaveF(pegomock.AnyString(),
		matchers.AnyMapOfStringToInterface(), matchers.AnyMapOfStringToInterface())
	resultSaverMock.VerifyWasCalled(pegomock.Once()).Save(pegomock.AnyString(),
		pegomock.AnyString())
	verifySendInform(t, messages.InformTypeFinished, 1)
}

func TestHandlesJoinResults_Failure(t *testing.T) {
	td := initTestData(t)
	pegomock.When(statusMock.Get(pegomock.AnyString())).ThenReturn(&api.TranscriptionResult{}, nil)
	rm := newTestResMsg()
	rm.Error = "error"
	msgdata, _ := json.Marshal(rm)
	td.joinResultsCh <- amqp.Delivery{Body: msgdata}
	close(td.joinResultsCh)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Once()).SaveError(pegomock.AnyString(), pegomock.AnyString())
	resultSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), pegomock.AnyString())
	verifySendInform(t, messages.InformTypeFailed, 1)
}

func TestHandlesOneStatus(t *testing.T) {
	td := initTestData(t)
	pegomock.When(dbMock.Get(pegomock.AnyString())).ThenReturn(&persistence.WorkData{Related: []string{"1", "11"}}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("2"))).ThenReturn(&api.TranscriptionResult{Status: "AudioConvert"}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("1"))).ThenReturn(&api.TranscriptionResult{Status: "Diarization"}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("11"))).ThenReturn(&api.TranscriptionResult{Status: "Diarization"}, nil)
	msg := newTestMsg()
	msg.Tags = append(msg.Tags, messages.NewTag(messages.TagParentID, "2"))
	msgdata, _ := json.Marshal(msg)
	td.statusCh <- amqp.Delivery{Body: msgdata}
	close(td.statusCh)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Once()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
}

func TestHandlesOneStatus_Skip(t *testing.T) {
	td := initTestData(t)
	pegomock.When(dbMock.Get(pegomock.AnyString())).ThenReturn(&persistence.WorkData{Related: []string{"1", "11"}}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("2"))).ThenReturn(&api.TranscriptionResult{Status: "AudioConvert"}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("1"))).ThenReturn(&api.TranscriptionResult{Status: "Diarization"}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("11"))).ThenReturn(&api.TranscriptionResult{Status: "AudioConvert"}, nil)
	msg := newTestMsg()
	msg.Tags = append(msg.Tags, messages.NewTag(messages.TagParentID, "2"))
	msgdata, _ := json.Marshal(msg)
	td.statusCh <- amqp.Delivery{Body: msgdata}
	close(td.statusCh)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
}

func TestHandlesOneStatus_SkipError(t *testing.T) {
	td := initTestData(t)
	pegomock.When(dbMock.Get(pegomock.AnyString())).ThenReturn(&persistence.WorkData{Related: []string{"1", "11"}}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("2"))).ThenReturn(&api.TranscriptionResult{Error: "error"}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("1"))).ThenReturn(&api.TranscriptionResult{Status: "Diarization"}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("11"))).ThenReturn(&api.TranscriptionResult{Status: "AudioConvert"}, nil)
	msg := newTestMsg()
	msg.Tags = append(msg.Tags, messages.NewTag(messages.TagParentID, "2"))
	msgdata, _ := json.Marshal(msg)
	td.statusCh <- amqp.Delivery{Body: msgdata}
	close(td.statusCh)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
}

func TestHandlesOneStatus_SavesError(t *testing.T) {
	td := initTestData(t)
	pegomock.When(dbMock.Get(pegomock.AnyString())).ThenReturn(&persistence.WorkData{Related: []string{"1", "11"}}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("2"))).ThenReturn(&api.TranscriptionResult{Status: "AudioConvert"}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("1"))).ThenReturn(&api.TranscriptionResult{Status: "Diarization"}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("11"))).ThenReturn(&api.TranscriptionResult{Error: "error"}, nil)
	msg := newTestMsg()
	msg.Tags = append(msg.Tags, messages.NewTag(messages.TagParentID, "2"))
	msgdata, _ := json.Marshal(msg)
	td.statusCh <- amqp.Delivery{Body: msgdata}
	close(td.statusCh)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Once()).SaveError(pegomock.AnyString(), pegomock.AnyString())
	verifySendInform(t, messages.InformTypeFailed, 1)
}

func TestHandlesOneCompleted(t *testing.T) {
	td := initTestData(t)
	pegomock.When(dbMock.Get(pegomock.AnyString())).ThenReturn(&persistence.WorkData{Related: []string{"1", "11"}}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("2"))).ThenReturn(&api.TranscriptionResult{Status: "AudioConvert"}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("1"))).ThenReturn(&api.TranscriptionResult{Status: status.Name(status.Completed)}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("11"))).ThenReturn(&api.TranscriptionResult{Status: status.Name(status.Completed)}, nil)
	msg := newTestMsg()
	msg.Tags = append(msg.Tags, messages.NewTag(messages.TagParentID, "2"))
	msgdata, _ := json.Marshal(msg)
	td.completeCh <- amqp.Delivery{Body: msgdata}
	close(td.completeCh)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Once()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
	verifySendMessage(t, messages.JoinResults, 1)
}

func TestHandlesOneCompleted_Skip(t *testing.T) {
	td := initTestData(t)
	pegomock.When(dbMock.Get(pegomock.AnyString())).ThenReturn(&persistence.WorkData{Related: []string{"1", "11"}}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("2"))).ThenReturn(&api.TranscriptionResult{Status: "AudioConvert"}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("1"))).ThenReturn(&api.TranscriptionResult{Status: status.Name(status.Completed)}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("11"))).ThenReturn(&api.TranscriptionResult{Status: status.Name(status.Rescore)}, nil)
	msg := newTestMsg()
	msg.Tags = append(msg.Tags, messages.NewTag(messages.TagParentID, "2"))
	msgdata, _ := json.Marshal(msg)
	td.completeCh <- amqp.Delivery{Body: msgdata}
	close(td.completeCh)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).Save(pegomock.AnyString(), matchers.AnyStatusStatus())
	verifySendMessage(t, messages.JoinResults, 0)
}

func TestHandlesOneCompleted_SkipError(t *testing.T) {
	td := initTestData(t)
	pegomock.When(dbMock.Get(pegomock.AnyString())).ThenReturn(&persistence.WorkData{Related: []string{"1", "11"}}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("2"))).ThenReturn(&api.TranscriptionResult{Error: "error"}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("1"))).ThenReturn(&api.TranscriptionResult{Status: status.Name(status.Completed)}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("11"))).ThenReturn(&api.TranscriptionResult{Status: status.Name(status.Completed)}, nil)
	msg := newTestMsg()
	msg.Tags = append(msg.Tags, messages.NewTag(messages.TagParentID, "2"))
	msgdata, _ := json.Marshal(msg)
	td.completeCh <- amqp.Delivery{Body: msgdata}
	close(td.completeCh)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Never()).SaveError(pegomock.AnyString(), pegomock.AnyString())
	verifySendMessage(t, messages.JoinResults, 0)
}

func TestHandlesOneCompleted_SaveError(t *testing.T) {
	td := initTestData(t)
	pegomock.When(dbMock.Get(pegomock.AnyString())).ThenReturn(&persistence.WorkData{Related: []string{"1", "11"}}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("2"))).ThenReturn(&api.TranscriptionResult{Status: status.Name(status.Completed)}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("1"))).ThenReturn(&api.TranscriptionResult{Status: status.Name(status.Completed)}, nil)
	pegomock.When(statusMock.Get(pegomock.EqString("11"))).ThenReturn(&api.TranscriptionResult{Error: "error"}, nil)
	msg := newTestMsg()
	msg.Tags = append(msg.Tags, messages.NewTag(messages.TagParentID, "2"))
	msgdata, _ := json.Marshal(msg)
	td.completeCh <- amqp.Delivery{Body: msgdata}
	close(td.completeCh)
	<-td.fc
	statusSaverMock.VerifyWasCalled(pegomock.Once()).SaveError(pegomock.AnyString(), pegomock.AnyString())
	verifySendInform(t, messages.InformTypeFailed, 1)
}

func TestMakeIdsFNMap(t *testing.T) {
	tests := []struct {
		i1, i2 []string
		e      string
		f      bool
	}{
		{i1: []string{"1"}, i2: []string{"a"}, e: "1=a"},
		{i1: []string{"1", "2"}, i2: []string{"a", "b"}, e: "1=a;2=b"},
		{i1: []string{"1", "2"}, i2: []string{"a"}, e: "", f: true},
	}

	for i, tc := range tests {
		v, err := makeIDsFnMap(tc.i1, tc.i2)
		assert.Equal(t, tc.f, err != nil, "Fail %d", i)
		assert.Equal(t, tc.e, v, "Fail %d", i)
	}

}

func TestCmpDur(t *testing.T) {
	assert.True(t, cmpDur(time.Hour, time.Hour))
	assert.True(t, cmpDur(time.Hour, time.Hour+time.Millisecond*900))
	assert.True(t, cmpDur(time.Hour+time.Millisecond*900, time.Hour))

	assert.False(t, cmpDur(time.Hour+time.Second*2, time.Hour))
	assert.False(t, cmpDur(time.Hour, time.Hour+time.Second*2))
}

func verifySendMessage(t *testing.T, mType string, count int) []*messages.QueueMessage {
	t.Helper()
	msgSenderMock.VerifyWasCalled(pegomock.Times(count)).Send(matchers.AnyMessagesMessage(), pegomock.EqString(mType), pegomock.AnyString())
	var res []*messages.QueueMessage
	if count > 0 {
		dm, _, _ := msgSenderMock.VerifyWasCalled(pegomock.Times(count)).Send(matchers.AnyMessagesMessage(), pegomock.EqString(mType), pegomock.AnyString()).
			GetCapturedArguments()

		m1 := dm.(*messages.QueueMessage)
		assert.Equal(t, "rec", m1.Recognizer)
		res = append(res, m1)
	}
	return res
}

func verifySendInform(t *testing.T, tp string, count int) {
	t.Helper()
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

func newTestResMsg() *messages.ResultMessage {
	return &messages.ResultMessage{QueueMessage: messages.QueueMessage{ID: "1", Recognizer: "rec"}, Result: "olia"}
}
