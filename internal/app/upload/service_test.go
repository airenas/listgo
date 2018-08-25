package upload

import (
	"bytes"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"github.com/petergtz/pegomock"
	. "github.com/smartystreets/goconvey/convey"
)

var statusSaverMock *mocks.MockSaver

func initTest() {
	statusSaverMock = mocks.NewMockSaver()
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
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "fileName")
		_, _ = io.Copy(part, strings.NewReader("body"))

		writer.WriteField("email", "a@a.a")
		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter(&ServiceData{MessageSender: testSender{},
				FileSaver:   testSaver{},
				StatusSaver: statusSaverMock}).ServeHTTP(resp, req)

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

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		writer.WriteField("email", "a@a.a")
		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter(&ServiceData{StatusSaver: statusSaverMock,
				MessageSender: testSender{}, FileSaver: testSaver{}}).ServeHTTP(resp, req)

			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
	})
}

func TestPOST_NoEmail(t *testing.T) {
	initTest()
	Convey("Given a HTTP request without email", t, func() {

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "fileName")
		_, _ = io.Copy(part, strings.NewReader("body"))

		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter(&ServiceData{StatusSaver: statusSaverMock, MessageSender: testSender{}, FileSaver: testSaver{}}).ServeHTTP(resp, req)

			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
	})
}

func TestPOST_Sender(t *testing.T) {
	initTest()
	Convey("Given a HTTP request without email", t, func() {

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "fileName")
		_, _ = io.Copy(part, strings.NewReader("body"))
		writer.WriteField("email", "a@a.a")

		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter(&ServiceData{MessageSender: testSenderFunc(
				func(m *messages.QueueMessage, q string, rq string) error {
					return nil
				}), StatusSaver: statusSaverMock, FileSaver: testSaver{}}).ServeHTTP(resp, req)

			Convey("Then the response should be a 200", func() {
				So(resp.Code, ShouldEqual, 200)
			})
		})
	})
}

func TestPOST_SenderFails(t *testing.T) {
	initTest()
	Convey("Given a HTTP request without email", t, func() {

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "fileName")
		_, _ = io.Copy(part, strings.NewReader("body"))
		writer.WriteField("email", "a@a.a")

		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter(&ServiceData{MessageSender: testSenderFunc(
				func(m *messages.QueueMessage, q string, rq string) error {
					return errors.New("Can not send")
				}), StatusSaver: statusSaverMock, FileSaver: testSaver{}}).ServeHTTP(resp, req)

			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
	})
}

func TestPOST_SaverFails(t *testing.T) {
	initTest()
	Convey("Given a HTTP request without email", t, func() {

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "fileName")
		_, _ = io.Copy(part, strings.NewReader("body"))
		writer.WriteField("email", "a@a.a")

		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter(&ServiceData{MessageSender: testSenderFunc(
				func(m *messages.QueueMessage, q string, rq string) error {
					return nil
				}), StatusSaver: statusSaverMock,
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
	Convey("Given a HTTP request without email", t, func() {

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "fileName")
		_, _ = io.Copy(part, strings.NewReader("body"))
		writer.WriteField("email", "a@a.a")

		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			pegomock.When(statusSaverMock.Save(pegomock.AnyString(),
				matchers.AnyStatusStatus())).ThenReturn(errors.New("error"))

			NewRouter(&ServiceData{MessageSender: testSender{},
				StatusSaver: statusSaverMock,
				FileSaver:   testSaver{}}).ServeHTTP(resp, req)

			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
	})
}

type testSenderFunc func(m *messages.QueueMessage, q string, rq string) error

func (f testSenderFunc) Send(m *messages.QueueMessage, q string, rq string) error {
	return f(m, q, rq)
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

type testSender struct{}

func (sender testSender) Send(m *messages.QueueMessage, q string, rq string) error {
	log.Printf("Sending msg %s\n", m.ID)
	return nil
}
