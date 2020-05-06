package kafkaintegration

import (
	"os"
	"testing"
	"time"

	"github.com/cenkalti/backoff"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"

	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	errc "bitbucket.org/airenas/listgo/internal/pkg/err"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"

	"github.com/stretchr/testify/assert"
)

type testdata struct {
	readerMock *mocks.MockKafkaReader
	writerMock *mocks.MockKafkaWriter
	dbMock     *mocks.MockDB
	filerMock  *mocks.MockFiler
	trMock     *mocks.MockTranscriber
	data       *ServiceData
	fc         <-chan os.Signal
}

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
}

func initTestData(t *testing.T) *testdata {
	initTest(t)
	res := testdata{}
	res.readerMock = mocks.NewMockKafkaReader()
	res.writerMock = mocks.NewMockKafkaWriter()
	res.dbMock = mocks.NewMockDB()
	res.filerMock = mocks.NewMockFiler()
	res.data = &ServiceData{}
	res.data.fc = utils.NewSignalChannel()
	res.fc = res.data.fc.C
	res.trMock = mocks.NewMockTranscriber()
	res.data.bp = &noBackOffProvider{}
	res.data.statusSleep = time.Nanosecond
	res.data.kReader = res.readerMock
	res.data.kWriter = res.writerMock
	res.data.db = res.dbMock
	res.data.tr = res.trMock
	res.data.filer = res.filerMock
	pegomock.When(res.data.kReader.Get()).ThenReturn(nil, errors.New("Can not read"))
	return &res
}

func Test_Initializes(t *testing.T) {
	td := initTestData(t)
	err := StartServer(td.data)
	go td.data.fc.Close()
	<-td.fc
	assert.Nil(t, err)
}

func Test_FailsReader_Exit(t *testing.T) {
	td := initTestData(t)

	pegomock.When(td.readerMock.Get()).ThenReturn(nil, errors.New("Can not read"))

	err := StartServer(td.data)
	<-td.fc
	assert.Nil(t, err)
}

func Test_TranscriptionOK(t *testing.T) {
	td := initTestData(t)
	mockReadMsg(td, &kafkaapi.Msg{ID: "1", RealMsg: &kafka.Message{}}, 1)
	pegomock.When(td.data.tr.Upload(matchers.AnyPtrToKafkaapiUploadData())).ThenReturn("u1", nil)
	pegomock.When(td.data.db.GetAudio(pegomock.AnyString())).ThenReturn(&kafkaapi.DBEntry{ID: "1", Data: "data"}, nil)
	pegomock.When(td.data.tr.GetStatus(pegomock.AnyString())).ThenReturn(&kafkaapi.Status{ID: "1", Completed: true, Text: "olia"}, nil)
	pegomock.When(td.data.tr.GetResult(pegomock.AnyString())).ThenReturn(&kafkaapi.Result{ID: "1", LatticeData: "fd", WebVTTData: "web"}, nil)
	StartServer(td.data)

	waitToFinish(t, td)

	msg := td.readerMock.VerifyWasCalled(pegomock.Once()).Commit(matchers.AnyPtrToKafkaapiMsg()).GetCapturedArguments()
	assert.NotNil(t, msg)
	assert.Equal(t, "1", msg.ID)

	dbCall := td.dbMock.VerifyWasCalled(pegomock.Once()).SaveResult(matchers.AnyPtrToKafkaapiDBResultEntry()).GetCapturedArguments()
	assert.NotNil(t, dbCall)
	assert.Equal(t, "1", dbCall.ID)
	assert.Equal(t, "fd", dbCall.Transcription.LatticeData)
	assert.Equal(t, "web", dbCall.Transcription.WebVTT)
	assert.Equal(t, "olia", dbCall.Transcription.Text)

	kafkaMsg := td.writerMock.VerifyWasCalled(pegomock.Once()).Write(matchers.AnyPtrToKafkaapiResponseMsg()).GetCapturedArguments()
	assert.Equal(t, "1", kafkaMsg.ID)

	filerMap := td.filerMock.VerifyWasCalled(pegomock.Once()).SetWorking(matchers.AnyPtrToKafkaapiKafkaTrMap()).GetCapturedArguments()
	assert.Equal(t, "u1", filerMap.TrID)
	assert.Equal(t, "1", filerMap.KafkaID)

	filerDelete := td.filerMock.VerifyWasCalled(pegomock.Once()).Delete(pegomock.AnyString()).GetCapturedArguments()
	assert.Equal(t, "1", filerDelete)

	td.trMock.VerifyWasCalled(pegomock.Once()).Delete(pegomock.AnyString())
}

func Test_TranscriptionOK_NoFailOnDelete(t *testing.T) {
	td := initTestData(t)
	mockReadMsg(td, &kafkaapi.Msg{ID: "1", RealMsg: &kafka.Message{}}, 1)
	pegomock.When(td.data.tr.Upload(matchers.AnyPtrToKafkaapiUploadData())).ThenReturn("u1", nil)
	pegomock.When(td.data.db.GetAudio(pegomock.AnyString())).ThenReturn(&kafkaapi.DBEntry{ID: "1", Data: "data"}, nil)
	pegomock.When(td.data.tr.GetStatus(pegomock.AnyString())).ThenReturn(&kafkaapi.Status{ID: "1", Completed: true, Text: "olia"}, nil)
	pegomock.When(td.data.tr.GetResult(pegomock.AnyString())).ThenReturn(&kafkaapi.Result{ID: "1", LatticeData: "fd"}, nil)
	pegomock.When(td.data.tr.Delete(pegomock.AnyString())).ThenReturn(errors.New("error"))
	StartServer(td.data)

	waitToFinish(t, td)

	td.writerMock.VerifyWasCalled(pegomock.Once()).Write(matchers.AnyPtrToKafkaapiResponseMsg())
	td.filerMock.VerifyWasCalled(pegomock.Once()).Delete(pegomock.AnyString())
	err := td.trMock.VerifyWasCalled(pegomock.Once()).Delete(pegomock.AnyString()).GetCapturedArguments()
	assert.NotNil(t, err)
}

func Test_Upload_Fails(t *testing.T) {
	td := initTestData(t)
	mockReadMsg(td, &kafkaapi.Msg{ID: "1", RealMsg: &kafka.Message{}}, 1)
	pegomock.When(td.data.db.GetAudio(pegomock.AnyString())).ThenReturn(&kafkaapi.DBEntry{ID: "1", Data: "data"}, nil)
	pegomock.When(td.data.tr.Upload(matchers.AnyPtrToKafkaapiUploadData())).ThenReturn("", errors.New("fail upload"))
	StartServer(td.data)

	waitToFinish(t, td)

	td.readerMock.VerifyWasCalled(pegomock.Once()).Commit(matchers.AnyPtrToKafkaapiMsg())
	saveData := td.dbMock.VerifyWasCalled(pegomock.Once()).SaveResult(matchers.AnyPtrToKafkaapiDBResultEntry()).GetCapturedArguments()
	assert.Equal(t, "1", saveData.ID)
	assert.Equal(t, errc.DefaultCode, saveData.Error.Code)
	assert.Contains(t, saveData.Error.Error, "upload")
	td.filerMock.VerifyWasCalled(pegomock.Never()).SetWorking(matchers.AnyPtrToKafkaapiKafkaTrMap())
	td.filerMock.VerifyWasCalled(pegomock.Never()).Delete(pegomock.AnyString())
}

func mockReadMsg(td *testdata, msg *kafkaapi.Msg, count int) {
	i := 0
	pegomock.When(td.readerMock.Get()).Then(func(params []pegomock.Param) pegomock.ReturnValues {
		if i >= count {
			return []pegomock.ReturnValue{nil, errors.New("stop")}
		}
		i = i + 1
		return []pegomock.ReturnValue{msg, nil}
	})
}

func Test_Transcription_Fails(t *testing.T) {
	td := initTestData(t)
	mockReadMsg(td, &kafkaapi.Msg{ID: "1", RealMsg: &kafka.Message{}}, 1)
	pegomock.When(td.data.db.GetAudio(pegomock.AnyString())).ThenReturn(&kafkaapi.DBEntry{ID: "1", Data: "data"}, nil)
	pegomock.When(td.data.tr.GetStatus(pegomock.AnyString())).ThenReturn(&kafkaapi.Status{ID: "1", Completed: false,
		ErrorCode: "ec", Error: "er"}, nil)

	StartServer(td.data)

	waitToFinish(t, td)

	msg := td.readerMock.VerifyWasCalled(pegomock.Once()).Commit(matchers.AnyPtrToKafkaapiMsg()).GetCapturedArguments()
	assert.NotNil(t, msg)
	assert.Equal(t, "1", msg.ID)

	dbCall := td.dbMock.VerifyWasCalled(pegomock.Once()).SaveResult(matchers.AnyPtrToKafkaapiDBResultEntry()).GetCapturedArguments()
	assert.NotNil(t, dbCall)
	assert.Equal(t, "1", dbCall.ID)
	assert.Equal(t, "ec", dbCall.Error.Code)
	assert.Equal(t, "er", dbCall.Error.Error)
	td.filerMock.VerifyWasCalled(pegomock.Once()).SetWorking(matchers.AnyPtrToKafkaapiKafkaTrMap())
	saveData := td.writerMock.VerifyWasCalled(pegomock.Once()).Write(matchers.AnyPtrToKafkaapiResponseMsg()).GetCapturedArguments()
	assert.Equal(t, "1", saveData.ID)
	assert.Equal(t, "", saveData.Error.Code)
	assert.Equal(t, "", saveData.Error.DebugMessage)

	td.filerMock.VerifyWasCalled(pegomock.Once()).Delete(pegomock.AnyString())
}

func Test_WriterFails_Exit(t *testing.T) {
	td := initTestData(t)
	mockReadMsg(td, &kafkaapi.Msg{ID: "1", RealMsg: &kafka.Message{}}, 100)
	pegomock.When(td.data.db.GetAudio(pegomock.AnyString())).ThenReturn(&kafkaapi.DBEntry{ID: "1", Data: "data"}, nil)
	pegomock.When(td.data.tr.GetStatus(pegomock.AnyString())).ThenReturn(&kafkaapi.Status{ID: "1", Completed: true, Text: "olia"}, nil)
	pegomock.When(td.data.tr.GetResult(pegomock.AnyString())).ThenReturn(&kafkaapi.Result{ID: "1", LatticeData: "fd"}, nil)
	pegomock.When(td.writerMock.Write(matchers.AnyPtrToKafkaapiResponseMsg())).ThenReturn(errors.New("can't write"))

	StartServer(td.data)

	waitToFinish(t, td)

	td.readerMock.VerifyWasCalled(pegomock.Never()).Commit(matchers.AnyPtrToKafkaapiMsg())
	td.filerMock.VerifyWasCalled(pegomock.Once()).SetWorking(matchers.AnyPtrToKafkaapiKafkaTrMap())
	td.filerMock.VerifyWasCalled(pegomock.Never()).Delete(pegomock.AnyString())
}

func Test_GetAudioFails_WriterInvoked_Exit(t *testing.T) {
	td := initTestData(t)
	mockReadMsg(td, &kafkaapi.Msg{ID: "1", RealMsg: &kafka.Message{}}, 100)
	pegomock.When(td.data.db.GetAudio(pegomock.AnyString())).ThenReturn(nil, errors.New("can't read"))
	pegomock.When(td.writerMock.Write(matchers.AnyPtrToKafkaapiResponseMsg())).ThenReturn(errors.New("can't write"))

	StartServer(td.data)

	waitToFinish(t, td)

	td.dbMock.VerifyWasCalled(pegomock.Never()).SaveResult(matchers.AnyPtrToKafkaapiDBResultEntry())
	td.readerMock.VerifyWasCalled(pegomock.Never()).Commit(matchers.AnyPtrToKafkaapiMsg())
	td.writerMock.VerifyWasCalled(pegomock.Once()).Write(matchers.AnyPtrToKafkaapiResponseMsg())
	td.filerMock.VerifyWasCalled(pegomock.Never()).Delete(pegomock.AnyString())
}

func Test_GetAudioFails(t *testing.T) {
	td := initTestData(t)
	mockReadMsg(td, &kafkaapi.Msg{ID: "1", RealMsg: &kafka.Message{}}, 1)
	pegomock.When(td.data.db.GetAudio(pegomock.AnyString())).ThenReturn(nil, errors.New("audio fails"))
	StartServer(td.data)

	waitToFinish(t, td)

	td.readerMock.VerifyWasCalled(pegomock.Once()).Commit(matchers.AnyPtrToKafkaapiMsg())
	td.writerMock.VerifyWasCalled(pegomock.Once()).Write(matchers.AnyPtrToKafkaapiResponseMsg())
	td.filerMock.VerifyWasCalled(pegomock.Never()).SetWorking(matchers.AnyPtrToKafkaapiKafkaTrMap())
	td.filerMock.VerifyWasCalled(pegomock.Never()).Delete(pegomock.AnyString())
}

func Test_StatusFails(t *testing.T) {
	td := initTestData(t)
	mockReadMsg(td, &kafkaapi.Msg{ID: "1", RealMsg: &kafka.Message{}}, 1)
	pegomock.When(td.data.db.GetAudio(pegomock.AnyString())).ThenReturn(&kafkaapi.DBEntry{ID: "1", Data: "data"}, nil)
	pegomock.When(td.data.tr.GetStatus(pegomock.AnyString())).ThenReturn(nil, errors.New("no status"))
	StartServer(td.data)

	waitToFinish(t, td)

	td.readerMock.VerifyWasCalled(pegomock.Once()).Commit(matchers.AnyPtrToKafkaapiMsg())
	td.writerMock.VerifyWasCalled(pegomock.Once()).Write(matchers.AnyPtrToKafkaapiResponseMsg())
	td.filerMock.VerifyWasCalled(pegomock.Once()).SetWorking(matchers.AnyPtrToKafkaapiKafkaTrMap())
	td.filerMock.VerifyWasCalled(pegomock.Never()).Delete(pegomock.AnyString())
}

func Test_GetResultFails_ReturnError(t *testing.T) {
	td := initTestData(t)
	mockReadMsg(td, &kafkaapi.Msg{ID: "1", RealMsg: &kafka.Message{}}, 1)
	pegomock.When(td.data.db.GetAudio(pegomock.AnyString())).ThenReturn(&kafkaapi.DBEntry{ID: "1", Data: "data"}, nil)
	pegomock.When(td.data.tr.GetStatus(pegomock.AnyString())).ThenReturn(&kafkaapi.Status{ID: "1", Completed: true, Text: "olia"}, nil)
	pegomock.When(td.data.tr.GetResult(pegomock.AnyString())).ThenReturn(nil, errors.New("no result"))
	StartServer(td.data)

	waitToFinish(t, td)

	msg := td.readerMock.VerifyWasCalled(pegomock.AtLeast(1)).Commit(matchers.AnyPtrToKafkaapiMsg()).GetCapturedArguments()
	assert.NotNil(t, msg)
	assert.Equal(t, "1", msg.ID)

	dbCall := td.dbMock.VerifyWasCalled(pegomock.AtLeast(1)).SaveResult(matchers.AnyPtrToKafkaapiDBResultEntry()).GetCapturedArguments()
	assert.NotNil(t, dbCall)
	assert.Equal(t, "1", dbCall.ID)
	assert.Equal(t, errc.DefaultCode, dbCall.Error.Code)
	assert.NotEqual(t, "", dbCall.Error.Error)

	td.writerMock.VerifyWasCalled(pegomock.Once()).Write(matchers.AnyPtrToKafkaapiResponseMsg())
	td.filerMock.VerifyWasCalled(pegomock.Once()).SetWorking(matchers.AnyPtrToKafkaapiKafkaTrMap())
	td.filerMock.VerifyWasCalled(pegomock.Once()).Delete(pegomock.AnyString())
}

func Test_SaveResultFails(t *testing.T) {
	td := initTestData(t)
	mockReadMsg(td, &kafkaapi.Msg{ID: "1", RealMsg: &kafka.Message{}}, 1)
	pegomock.When(td.data.db.GetAudio(pegomock.AnyString())).ThenReturn(&kafkaapi.DBEntry{ID: "1", Data: "data"}, nil)
	pegomock.When(td.data.tr.GetStatus(pegomock.AnyString())).ThenReturn(&kafkaapi.Status{ID: "1", Completed: true, Text: "olia"}, nil)
	pegomock.When(td.data.tr.GetResult(pegomock.AnyString())).ThenReturn(&kafkaapi.Result{ID: "1", LatticeData: "fd"}, nil)
	pegomock.When(td.data.db.SaveResult(matchers.AnyPtrToKafkaapiDBResultEntry())).ThenReturn(errors.New("can't save"))
	StartServer(td.data)

	waitToFinish(t, td)

	td.readerMock.VerifyWasCalled(pegomock.Once()).Commit(matchers.AnyPtrToKafkaapiMsg())

	saveData := td.writerMock.VerifyWasCalled(pegomock.Once()).Write(matchers.AnyPtrToKafkaapiResponseMsg()).GetCapturedArguments()
	assert.Equal(t, "1", saveData.ID)
	assert.NotNil(t, saveData.Error)
	assert.Equal(t, errc.DefaultCode, saveData.Error.Code)
	assert.Contains(t, saveData.Error.DebugMessage, "can't save")

	td.filerMock.VerifyWasCalled(pegomock.Once()).SetWorking(matchers.AnyPtrToKafkaapiKafkaTrMap())
	td.filerMock.VerifyWasCalled(pegomock.Once()).Delete(pegomock.AnyString())
}

func waitToFinish(t *testing.T, td *testdata) {
	select {
	case <-td.fc:
	case <-time.After(5 * time.Second):
		assert.Fail(t, "method timeout")
	}
}

func Test_NoReader_Fails(t *testing.T) {
	td := initTestData(t)
	td.data.kReader = nil
	testFailsStart(t, td)
}

func testFailsStart(t *testing.T, td *testdata) {
	err := StartServer(td.data)
	go td.data.fc.Close()
	waitToFinish(t, td)

	assert.NotNil(t, err)
}

func Test_NoWriter_Fails(t *testing.T) {
	td := initTestData(t)
	td.data.kWriter = nil
	testFailsStart(t, td)
}

func Test_NoDB_Fails(t *testing.T) {
	td := initTestData(t)
	td.data.db = nil
	testFailsStart(t, td)
}

func Test_NoTranscriber_Fails(t *testing.T) {
	td := initTestData(t)
	td.data.tr = nil
	testFailsStart(t, td)
}

func Test_NoFiler_Fails(t *testing.T) {
	td := initTestData(t)
	td.data.filer = nil
	testFailsStart(t, td)
}

func Test_NoBackoff_Fails(t *testing.T) {
	td := initTestData(t)
	td.data.bp = nil
	testFailsStart(t, td)
}

type noBackOffProvider struct {
}

func (bp *noBackOffProvider) Get() backoff.BackOff {
	return &backoff.StopBackOff{}
}
