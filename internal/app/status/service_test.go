package status

import (
	"errors"
	"log"
	"net/http/httptest"
	"testing"

	"bitbucket.org/airenas/listgo/internal/app/status/api"
	. "github.com/smartystreets/goconvey/convey"
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

type testStatusFunc func(ID string) (*api.TranscriptionResult, error)

func (f testStatusFunc) Get(ID string) (*api.TranscriptionResult, error) {
	return f(ID)
}

type testStatusProvider struct{}

func (p testStatusProvider) Get(ID string) (*api.TranscriptionResult, error) {
	log.Printf("Get status %s \n", ID)
	return &api.TranscriptionResult{}, nil
}
