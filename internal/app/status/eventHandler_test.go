package status

import (
	"errors"
	"testing"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"

	"bitbucket.org/airenas/listgo/internal/app/status/api"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"

	"github.com/petergtz/pegomock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/streadway/amqp"
)

var statusProviderMock *mocks.MockProvider
var connMock *mocks.MockWsConn

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	statusProviderMock = mocks.NewMockProvider()
	connMock = mocks.NewMockWsConn()
}

func Test_ListenQueue(t *testing.T) {
	Convey("Invoking the listen function", t, func() {
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
		Convey("When msg is send", func() {
			go f()
			d := amqp.Delivery{Body: []byte("id")}
			c <- d
			close(c)
			<-waitc
			Convey("msg is processed", func() {
				So(true, ShouldBeTrue) // no error
			})
		})
		Convey("When msg is send with existing id", func() {
			saveConnection(connMock, "id")
			pegomock.When(statusProviderMock.Get(pegomock.AnyString())).ThenReturn(&api.TranscriptionResult{}, nil)
			go f()
			d := amqp.Delivery{Body: []byte("id")}
			c <- d
			close(c)
			<-waitc
			Convey("msg is processed", func() {
				So(true, ShouldBeTrue) // no error
			})
			deleteConnection(connMock)
		})
		Convey("When msg is send with failing provider", func() {
			saveConnection(connMock, "id")
			pegomock.When(statusProviderMock.Get(pegomock.AnyString())).ThenReturn(nil, errors.New("error"))
			go f()
			d := amqp.Delivery{Body: []byte("id")}
			c <- d
			close(c)
			<-waitc
			Convey("msg is processed", func() {
				So(true, ShouldBeTrue) // no error
			})
			deleteConnection(connMock)
		})
		Convey("When msg is send with failing connection", func() {
			saveConnection(connMock, "id")
			pegomock.When(statusProviderMock.Get(pegomock.AnyString())).ThenReturn(&api.TranscriptionResult{}, nil)
			pegomock.When(connMock.WriteJSON(matchers.AnyPtrToApiTranscriptionResult())).ThenReturn(errors.New("error"))

			go f()
			d := amqp.Delivery{Body: []byte("id")}
			c <- d
			close(c)
			<-waitc
			Convey("msg is processed", func() {
				So(true, ShouldBeTrue) // no error
			})
			deleteConnection(connMock)
		})
		Convey("When msg is send to multiple connections", func() {
			saveConnection(connMock, "id1")
			pegomock.When(statusProviderMock.Get(pegomock.AnyString())).ThenReturn(&api.TranscriptionResult{}, nil)
			pegomock.When(connMock.WriteJSON(matchers.AnyPtrToApiTranscriptionResult())).ThenReturn(nil)
			connMock1 := mocks.NewMockWsConn()
			pegomock.When(connMock1.WriteJSON(matchers.AnyPtrToApiTranscriptionResult())).ThenReturn(nil)
			saveConnection(connMock1, "id1")
			go f()
			d := amqp.Delivery{Body: []byte("id1")}
			c <- d
			close(c)
			<-waitc
			Convey("msg is processed", func() {
				connMock.VerifyWasCalled(pegomock.Times(1)).WriteJSON(matchers.AnyPtrToApiTranscriptionResult())
				connMock1.VerifyWasCalled(pegomock.Times(1)).WriteJSON(matchers.AnyPtrToApiTranscriptionResult())
			})
			deleteConnection(connMock)
			deleteConnection(connMock1)
		})
	})
}

func Test_RegisteringQueue(t *testing.T) {
	Convey("Invoking the register function", t, func() {
		data := &ServiceData{}
		i := 0
		fail := true
		c := make(chan amqp.Delivery)
		data.EventChannelFunc = func() (<-chan amqp.Delivery, error) {
			i++
			if fail {
				return nil, errors.New("error")
			}
			return c, nil
		}
		fc := make(chan bool)
		waitc := make(chan bool)
		f := func() {
			registerQueue(data, fc, time.Millisecond)
			waitc <- true
		}
		Convey("When queue func fails", func() {
			go f()
			time.Sleep(time.Millisecond * 100)
			close(fc)
			<-waitc
			Convey("Tries reconnect", func() {
				So(i, ShouldBeGreaterThan, 1)
			})
		})
		Convey("Restores after failure", func() {
			go f()
			time.Sleep(time.Millisecond * 100)
			fail = false
			i = 0
			time.Sleep(time.Millisecond * 100)
			close(fc)
			close(c)
			<-waitc
			Convey("No retry", func() {
				So(i, ShouldEqual, 1)
			})
		})
		Convey("No failure", func() {
			fail = false
			go f()
			time.Sleep(time.Millisecond * 100)
			close(fc)
			close(c)
			<-waitc
			Convey("No retry", func() {
				So(i, ShouldEqual, 1)
			})
		})
	})
}
