package upload

import (
	"bytes"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"github.com/gorilla/mux"
	"github.com/petergtz/pegomock"
	. "github.com/smartystreets/goconvey/convey"
)

var statusSaverMock *mocks.MockSaver

var requestSaverMock *mocks.MockRequestSaver

var msgSenderMock *mocks.MockSender

func initTest() {
	statusSaverMock = mocks.NewMockSaver()
	requestSaverMock = mocks.NewMockRequestSaver()
	msgSenderMock = mocks.NewMockSender()
}

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

func TestNoFile(t *testing.T) {
	Convey("Given a HTTP request for /upload", t, func() {
		req := httptest.NewRequest("POST", "/upload", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter(&ServiceData{}).ServeHTTP(resp, req)

			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
	})
}

func TestNoFilePOST(t *testing.T) {
	Convey("Given a HTTP request for /upload", t, func() {
		req := httptest.NewRequest("POST", "/upload", nil)
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter(&ServiceData{}).ServeHTTP(resp, req)

			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
	})
}

func TestPOST(t *testing.T) {
	initTest()
	Convey("Given a HTTP request for /upload", t, func() {
		req := newReq("filename", "a@a.a")
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			newRouter().ServeHTTP(resp, req)

			Convey("Then the response should be a 200", func() {
				So(resp.Code, ShouldEqual, 200)
			})
			Convey("Then the response body should start with id", func() {
				So(resp.Body.String(), ShouldStartWith, `{"id":"`)
			})
		})
	})
}

func TestPOSTNoFile(t *testing.T) {
	initTest()
	Convey("Given a HTTP request for /upload", t, func() {
		req := newReq("", "a@a.a")
		resp := httptest.NewRecorder()
		Convey("When the request is handled by the Router", func() {
			newRouter().ServeHTTP(resp, req)

			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
	})
}

func newReq(file string, email string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if file != "" {
		part, _ := writer.CreateFormFile("file", "fileName")
		_, _ = io.Copy(part, strings.NewReader("body"))

	}
	if email != "" {
		writer.WriteField("email", email)

	}
	writer.Close()
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func newRouter() *mux.Router {
	return NewRouter(&ServiceData{StatusSaver: statusSaverMock,
		MessageSender: msgSenderMock,
		RequestSaver:  requestSaverMock,
		FileSaver:     testSaver{}})
}

func TestPOST_WrongEmail(t *testing.T) {
	initTest()
	Convey("Given a test", t, func() {
		resp := httptest.NewRecorder()
		Convey("When no email is given", func() {
			newRouter().ServeHTTP(resp, newReq("file", ""))
			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
		Convey("When wrong email is given", func() {
			newRouter().ServeHTTP(resp, newReq("file", "a@"))
			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
		Convey("When wrong email(1) is given", func() {
			newRouter().ServeHTTP(resp, newReq("file", "@a"))
			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
		Convey("When wrong email(2) is given", func() {
			newRouter().ServeHTTP(resp, newReq("file", "a_a"))
			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
	})
}

func TestPOST_Sender(t *testing.T) {
	initTest()
	Convey("Given a HTTP request", t, func() {
		req := newReq("filename", "a@a.a")
		resp := httptest.NewRecorder()
		Convey("When the request is handled by the Router", func() {
			NewRouter(&ServiceData{MessageSender: msgSenderMock, StatusSaver: statusSaverMock,
				RequestSaver: requestSaverMock,
				FileSaver:    testSaver{}}).ServeHTTP(resp, req)

			Convey("Then the response should be a 200", func() {
				So(resp.Code, ShouldEqual, 200)
			})
		})
	})
}

func TestPOST_SenderFails(t *testing.T) {
	initTest()
	Convey("Given a HTTP request", t, func() {
		req := newReq("filename", "a@a.a")
		resp := httptest.NewRecorder()
		pegomock.When(msgSenderMock.Send(matchers.AnyMessagesMessage(), pegomock.AnyString(),
			pegomock.AnyString())).ThenReturn(errors.New("Can not send"))

		Convey("When the request is handled by the Router", func() {
			NewRouter(&ServiceData{MessageSender: msgSenderMock,
				StatusSaver:  statusSaverMock,
				RequestSaver: requestSaverMock,
				FileSaver:    testSaver{}}).ServeHTTP(resp, req)

			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
	})
}

func TestPOST_SaverFails(t *testing.T) {
	initTest()
	Convey("Given a HTTP request", t, func() {
		req := newReq("filename", "a@a.a")
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter(&ServiceData{MessageSender: msgSenderMock, StatusSaver: statusSaverMock,
				RequestSaver: requestSaverMock,
				FileSaver: testSaverFunc(
					func(id string, reader io.Reader) error {
						return errors.New("Can not send")
					})}).ServeHTTP(resp, req)

			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
	})
}

func TestPOST_StatusSaverFails(t *testing.T) {
	initTest()
	Convey("Given a HTTP request", t, func() {
		req := newReq("filename", "a@a.a")
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			pegomock.When(statusSaverMock.Save(pegomock.AnyString(),
				matchers.AnyStatusStatus())).ThenReturn(errors.New("error"))

			newRouter().ServeHTTP(resp, req)

			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
	})
}

func TestPOST_RequestSaverFails(t *testing.T) {
	initTest()
	Convey("Given a HTTP request", t, func() {
		req := newReq("filename", "a@a.a")
		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			pegomock.When(requestSaverMock.Save(pegomock.AnyString(),
				pegomock.AnyString())).ThenReturn(errors.New("error"))

			newRouter().ServeHTTP(resp, req)

			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
	})
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
