package status

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"

	"bitbucket.org/airenas/listgo/internal/app/status/api"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"

	"github.com/petergtz/pegomock"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

var statusProviderMock *mocks.MockProvider
var connMock *mocks.MockWsConn

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	statusProviderMock = mocks.NewMockProvider()
	connMock = mocks.NewMockWsConn()
}

func Test_ListenQueue_MsgSent(t *testing.T) {
	initTest(t)
	data := &ServiceData{}
	data.StatusProvider = statusProviderMock
	c := make(chan amqp.Delivery)
	fc := make(chan bool)
	waitc := make(chan bool)
	f := func() {
		listenQueue(c, data, fc)
		waitc <- true
	}
	go f()
	d := amqp.Delivery{Body: []byte("id")}
	c <- d
	close(c)
	<-waitc
}

type testdata struct {
	c     chan amqp.Delivery
	data  *ServiceData
	fc    chan bool
	waitc chan bool
	f     func()
	fail  int64
	i     int64
}

func initTestData(t *testing.T) *testdata {
	initTest(t)
	res := testdata{}
	res.c = make(chan amqp.Delivery)
	res.data = &ServiceData{}
	res.data.StatusProvider = statusProviderMock
	res.fc = make(chan bool)
	res.waitc = make(chan bool)
	res.f = func() {
		listenQueue(res.c, res.data, res.fc)
		res.waitc <- true
	}
	return &res
}

func Test_ListenQueue_MsgSentWithID(t *testing.T) {
	td := initTestData(t)

	saveConnection(connMock, "id")
	pegomock.When(statusProviderMock.Get(pegomock.AnyString())).ThenReturn(&api.TranscriptionResult{}, nil)
	go td.f()
	d := amqp.Delivery{Body: []byte("id")}
	td.c <- d
	close(td.c)
	<-td.waitc
	deleteConnection(connMock)
}

func Test_ListenQueue_MsgSentWithExistingID(t *testing.T) {
	td := initTestData(t)

	saveConnection(connMock, "id")
	pegomock.When(statusProviderMock.Get(pegomock.AnyString())).ThenReturn(&api.TranscriptionResult{}, nil)
	go td.f()
	d := amqp.Delivery{Body: []byte("id")}
	td.c <- d
	close(td.c)
	<-td.waitc
	deleteConnection(connMock)
}

func Test_ListenQueue_WithFailingProvider(t *testing.T) {
	td := initTestData(t)

	saveConnection(connMock, "id")
	pegomock.When(statusProviderMock.Get(pegomock.AnyString())).ThenReturn(nil, errors.New("error"))
	go td.f()
	d := amqp.Delivery{Body: []byte("id")}
	td.c <- d
	close(td.c)
	<-td.waitc
	deleteConnection(connMock)
}

func Test_ListenQueue_WithFailingConnection(t *testing.T) {
	td := initTestData(t)

	saveConnection(connMock, "id")
	pegomock.When(statusProviderMock.Get(pegomock.AnyString())).ThenReturn(&api.TranscriptionResult{}, nil)
	pegomock.When(connMock.WriteJSON(matchers.AnyPtrToApiTranscriptionResult())).ThenReturn(errors.New("error"))

	go td.f()
	d := amqp.Delivery{Body: []byte("id")}
	td.c <- d
	close(td.c)
	<-td.waitc
	deleteConnection(connMock)
}

func Test_ListenQueue_NultipleConnections(t *testing.T) {
	td := initTestData(t)

	saveConnection(connMock, "id1")
	defer deleteConnection(connMock)
	pegomock.When(statusProviderMock.Get(pegomock.AnyString())).ThenReturn(&api.TranscriptionResult{}, nil)
	pegomock.When(connMock.WriteJSON(matchers.AnyPtrToApiTranscriptionResult())).ThenReturn(nil)
	connMock1 := mocks.NewMockWsConn()
	defer deleteConnection(connMock1)
	pegomock.When(connMock1.WriteJSON(matchers.AnyPtrToApiTranscriptionResult())).ThenReturn(nil)
	saveConnection(connMock1, "id1")
	go td.f()
	d := amqp.Delivery{Body: []byte("id1")}
	td.c <- d
	close(td.c)
	<-td.waitc
	connMock.VerifyWasCalled(pegomock.Times(1)).WriteJSON(matchers.AnyPtrToApiTranscriptionResult())
	connMock1.VerifyWasCalled(pegomock.Times(1)).WriteJSON(matchers.AnyPtrToApiTranscriptionResult())
}

func initTestDataRegisterQueue(t *testing.T) *testdata {
	t.Helper()
	res := initTestData(t)
	res.c = make(chan amqp.Delivery)
	res.data = &ServiceData{}
	res.data.StatusProvider = statusProviderMock
	res.fc = make(chan bool)
	res.waitc = make(chan bool)
	res.fail = 1
	res.i = 0

	res.data.EventChannelFunc = func() (<-chan amqp.Delivery, error) {
		atomic.AddInt64(&res.i, 1)
		if atomic.LoadInt64(&res.fail) > 0 {
			return nil, errors.New("error")
		}
		return res.c, nil
	}
	res.f = func() {
		registerQueue(res.data, res.fc, time.Millisecond)
		res.waitc <- true
	}
	return res
}

func Test_RegisteringQueue_FunctionFails(t *testing.T) {
	td := initTestDataRegisterQueue(t)

	go td.f()
	time.Sleep(time.Millisecond * 100)
	close(td.fc)
	<-td.waitc
	assert.True(t, td.i > 1)
}

func Test_RegisteringQueue_Restores(t *testing.T) {
	td := initTestDataRegisterQueue(t)

	go td.f()
	time.Sleep(time.Millisecond * 100)
	atomic.StoreInt64(&td.fail, 0)
	atomic.StoreInt64(&td.i, 0)
	time.Sleep(time.Millisecond * 100)
	close(td.fc)
	close(td.c)
	<-td.waitc
	assert.Equal(t, int64(1), td.i)
}

func Test_RegisteringQueue_NoFailure(t *testing.T) {
	td := initTestDataRegisterQueue(t)
	atomic.StoreInt64(&td.fail, 0)
	go td.f()
	time.Sleep(time.Millisecond * 100)
	close(td.fc)
	close(td.c)
	<-td.waitc
	assert.Equal(t, int64(1), td.i)
}
