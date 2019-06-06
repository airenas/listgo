package kafkaintegration

import (
	"os"
	"testing"

	"github.com/cenkalti/backoff"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"

	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"

	"github.com/stretchr/testify/assert"
)

type testdata struct {
	readerMock *mocks.MockKafkaReader
	dbMock     *mocks.MockDB
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
	res.dbMock = mocks.NewMockDB()
	res.data = &ServiceData{}
	res.data.fc = utils.NewSignalChannel()
	res.fc = res.data.fc.C
	res.data.bp = &noBackOffProvider{}
	res.data.kReader = res.readerMock
	res.data.kWriter = mocks.NewMockKafkaWriter()
	res.data.db = res.dbMock
	res.data.tr = mocks.NewMockTranscriber()
	res.data.filer = mocks.NewMockFiler()
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
	called := false
	pegomock.When(td.readerMock.Get()).Then(func(params []pegomock.Param) pegomock.ReturnValues {
		if called {
			td.data.fc.Close()
		}
		called = true
		return []pegomock.ReturnValue{&kafkaapi.Msg{ID: "1", RealMsg: &kafka.Message{}}, nil}
	})
	pegomock.When(td.data.db.GetAudio(pegomock.AnyString())).ThenReturn(&kafkaapi.DBEntry{ID: "1", Data: "data"}, nil)
	pegomock.When(td.data.tr.GetStatus(pegomock.AnyString())).ThenReturn(&kafkaapi.Status{ID: "1", Completed: true, Text: "olia"}, nil)
	pegomock.When(td.data.tr.GetResult(pegomock.AnyString())).ThenReturn(&kafkaapi.Result{ID: "1", FileData: "fd"}, nil)
	StartServer(td.data)

	<-td.fc

	msg := td.readerMock.VerifyWasCalled(pegomock.AtLeast(1)).Commit(matchers.AnyPtrToKafkaapiMsg()).GetCapturedArguments()
	assert.NotNil(t, msg)
	assert.Equal(t, "1", msg.ID)

	dbCall := td.dbMock.VerifyWasCalled(pegomock.AtLeast(1)).SaveResult(matchers.AnyPtrToKafkaapiDBResultEntry()).GetCapturedArguments()
	assert.NotNil(t, dbCall)
	assert.Equal(t, "1", dbCall.ID)
	assert.Equal(t, kafkaapi.DBStatusDone, dbCall.Status)
	assert.Equal(t, "fd", dbCall.Transcription.ResultFileData)
	assert.Equal(t, "olia", dbCall.Transcription.Text)
}

func Test_Transcriptin_Fails(t *testing.T) {
	td := initTestData(t)
	called := false
	pegomock.When(td.readerMock.Get()).Then(func(params []pegomock.Param) pegomock.ReturnValues {
		if called {
			td.data.fc.Close()
		}
		called = true
		return []pegomock.ReturnValue{&kafkaapi.Msg{ID: "1", RealMsg: &kafka.Message{}}, nil}
	})
	pegomock.When(td.data.db.GetAudio(pegomock.AnyString())).ThenReturn(&kafkaapi.DBEntry{ID: "1", Data: "data"}, nil)
	pegomock.When(td.data.tr.GetStatus(pegomock.AnyString())).ThenReturn(&kafkaapi.Status{ID: "1", Completed: false,
		ErrorCode: "ec", Error: "er"}, nil)
	StartServer(td.data)

	<-td.fc

	msg := td.readerMock.VerifyWasCalled(pegomock.AtLeast(1)).Commit(matchers.AnyPtrToKafkaapiMsg()).GetCapturedArguments()
	assert.NotNil(t, msg)
	assert.Equal(t, "1", msg.ID)

	dbCall := td.dbMock.VerifyWasCalled(pegomock.AtLeast(1)).SaveResult(matchers.AnyPtrToKafkaapiDBResultEntry()).GetCapturedArguments()
	assert.NotNil(t, dbCall)
	assert.Equal(t, "1", dbCall.ID)
	assert.Equal(t, kafkaapi.DBStatusFailed, dbCall.Status)
	assert.Equal(t, "ec", dbCall.Err.Code)
	assert.Equal(t, "er", dbCall.Err.Error)
}

func Test_GetAudioFails_Exit(t *testing.T) {
	td := initTestData(t)
	pegomock.When(td.readerMock.Get()).Then(func(params []pegomock.Param) pegomock.ReturnValues {
		return []pegomock.ReturnValue{&kafkaapi.Msg{ID: "1", RealMsg: &kafka.Message{}}, nil}
	})
	pegomock.When(td.data.db.GetAudio(pegomock.AnyString())).ThenReturn(nil, errors.New("audio fails"))
	StartServer(td.data)

	<-td.fc

	td.readerMock.VerifyWasCalled(pegomock.Never()).Commit(matchers.AnyPtrToKafkaapiMsg())
}

func Test_NoReader_Fails(t *testing.T) {
	td := initTestData(t)
	td.data.kReader = nil
	testFailsStart(t, td)
}

func testFailsStart(t *testing.T, td *testdata) {
	err := StartServer(td.data)
	go td.data.fc.Close()
	<-td.fc
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
