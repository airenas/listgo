package result

import (
	"errors"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"github.com/gorilla/mux"
	"github.com/petergtz/pegomock"
	. "github.com/smartystreets/goconvey/convey"
)

var fileLoaderMock *mocks.MockFileLoader
var fileMock *mocks.MockFile
var fileNameProviderMock *mocks.MockFileNameProvider

func initTest() {
	fileLoaderMock = mocks.NewMockFileLoader()
	fileMock = mocks.NewMockFile()
	fileNameProviderMock = mocks.NewMockFileNameProvider()
}

func TestWrongPath(t *testing.T) {

	Convey("Given a HTTP request for /invalid", t, func() {
		req := httptest.NewRequest("GET", "/invalid", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			newRouter().ServeHTTP(resp, req)

			Convey("Then the response should be a 404", func() {
				So(resp.Code, ShouldEqual, 404)
			})
		})
	})
}

func TestNoID(t *testing.T) {
	Convey("Given a HTTP request for /audio/", t, func() {
		req := httptest.NewRequest("GET", "/audio/", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			newRouter().ServeHTTP(resp, req)

			Convey("Then the response should be a 404", func() {
				So(resp.Code, ShouldEqual, 404)
			})
		})
	})
}

func TestGET(t *testing.T) {
	initTest()
	Convey("Given a HTTP request for /audio", t, func() {
		req := httptest.NewRequest("GET", "/audio/id", nil)
		resp := httptest.NewRecorder()
		pegomock.When(fileNameProviderMock.Get(pegomock.AnyString())).ThenReturn("olia", nil)
		pegomock.When(fileLoaderMock.Load(pegomock.AnyString())).ThenReturn(fileMock, nil)
		pegomock.When(fileMock.Stat()).ThenReturn(mockedFileInfo{}, nil)
		pegomock.When(fileMock.Seek(pegomock.AnyInt64(), pegomock.AnyInt())).ThenReturn(int64(2), nil)
		pegomock.When(fileMock.Read(anyByteArray())).Then(
			func(params []pegomock.Param) pegomock.ReturnValues {
				return []pegomock.ReturnValue{2, nil}
			})
		Convey("When the request is handled by the Router", func() {
			newRouter().ServeHTTP(resp, req)

			Convey("Then the response should be a 200", func() {
				So(resp.Code, ShouldEqual, 200)
			})
			Convey("Then the response body should not be empty", func() {
				So(resp.Body.String(), ShouldNotBeEmpty)
			})
		})
	})
}

func newRouter() *mux.Router {
	return NewRouter(&ServiceData{fileLoader: fileLoaderMock,
		fileNameProvider: fileNameProviderMock})
}

func Test_FileNameProviderFails(t *testing.T) {
	initTest()
	Convey("Given a HTTP request", t, func() {
		req := httptest.NewRequest("GET", "/audio/id", nil)
		resp := httptest.NewRecorder()
		pegomock.When(fileNameProviderMock.Get(pegomock.AnyString())).ThenReturn("", errors.New("Can not get"))

		Convey("When the request is handled by the Router", func() {
			newRouter().ServeHTTP(resp, req)
			Convey("Then the response should be a 404", func() {
				So(resp.Code, ShouldEqual, 404)
			})
		})
	})
}

func Test_FileLoaderFails(t *testing.T) {
	initTest()
	Convey("Given a HTTP request", t, func() {
		req := httptest.NewRequest("GET", "/audio/id", nil)
		resp := httptest.NewRecorder()
		pegomock.When(fileNameProviderMock.Get(pegomock.AnyString())).ThenReturn("olia", nil)
		pegomock.When(fileLoaderMock.Load(pegomock.AnyString())).ThenReturn(nil, errors.New("Can not get"))

		Convey("When the request is handled by the Router", func() {
			newRouter().ServeHTTP(resp, req)
			Convey("Then the response should be a 404", func() {
				So(resp.Code, ShouldEqual, 404)
			})
		})
	})
}

func Test_FileStatFails(t *testing.T) {
	initTest()
	Convey("Given a HTTP request", t, func() {
		req := httptest.NewRequest("GET", "/audio/id", nil)
		resp := httptest.NewRecorder()
		pegomock.When(fileNameProviderMock.Get(pegomock.AnyString())).ThenReturn("olia", nil)
		pegomock.When(fileLoaderMock.Load(pegomock.AnyString())).ThenReturn(fileMock, nil)
		pegomock.When(fileMock.Stat()).ThenReturn(mockedFileInfo{}, errors.New("Can not get"))

		Convey("When the request is handled by the Router", func() {
			newRouter().ServeHTTP(resp, req)
			Convey("Then the response should be a 404", func() {
				So(resp.Code, ShouldEqual, 404)
			})
		})
	})
}

type mockedFileInfo struct {
	os.FileInfo
}

func (mockedFileInfo) Name() string {
	return "olia.wav"
}

func (mockedFileInfo) ModTime() time.Time {
	return time.Now()
}

func anyByteArray() []byte {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf([]byte{})))
	return []byte{}
}
