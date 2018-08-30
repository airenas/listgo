package status

import (
	"errors"
	"log"
	"net/http/httptest"
	"testing"
	"time"

	"bitbucket.org/airenas/listgo/internal/app/status/api"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/streadway/amqp"
)

func TestWrongPath(t *testing.T) {

	Convey("Given a HTTP request for /invalid", t, func() {
		req := httptest.NewRequest("GET", "/invalid", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter(&ServiceData{}).ServeHTTP(resp, req)

			Convey("Then the response should be a 404", func() {
				So(resp.Code, ShouldEqual, 404)
			})
		})
	})
}

func TestNoID(t *testing.T) {
	Convey("Given a HTTP request for /result", t, func() {
		req := httptest.NewRequest("GET", "/result/", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter(&ServiceData{}).ServeHTTP(resp, req)

			Convey("Then the response should be a 404", func() {
				So(resp.Code, ShouldEqual, 404)
			})
		})
	})
}

func Test_ReturnsResult(t *testing.T) {
	Convey("Given a HTTP request for ID x", t, func() {

		req := httptest.NewRequest("GET", "/result/x", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter(&ServiceData{StatusProvider: testStatusProvider{}}).ServeHTTP(resp, req)

			Convey("Then the response should be a 200", func() {
				So(resp.Code, ShouldEqual, 200)
			})
			Convey("Then the response body should start with id", func() {
				So(resp.Body.String(), ShouldStartWith, `{"id":"`)
			})
		})
	})
}

func Test_ProviderFails(t *testing.T) {
	Convey("Given a HTTP request", t, func() {
		req := httptest.NewRequest("GET", "/result/x", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter(&ServiceData{StatusProvider: testStatusFunc(
				func(ID string) (*api.TranscriptionResult, error) {
					return nil, errors.New("Can not get")
				})}).ServeHTTP(resp, req)

			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
	})
}

func Test_Registering_Queue(t *testing.T) {
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

type testStatusFunc func(ID string) (*api.TranscriptionResult, error)

func (f testStatusFunc) Get(ID string) (*api.TranscriptionResult, error) {
	return f(ID)
}

type testStatusProvider struct{}

func (p testStatusProvider) Get(ID string) (*api.TranscriptionResult, error) {
	log.Printf("Get status %s \n", ID)
	return &api.TranscriptionResult{}, nil
}
