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

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
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
				StatusSaver: testStatusSaver{}}).ServeHTTP(resp, req)

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
	Convey("Given a HTTP request for /upload", t, func() {

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		writer.WriteField("email", "a@a.a")
		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp := httptest.NewRecorder()

		Convey("When the request is handled by the Router", func() {
			NewRouter(&ServiceData{StatusSaver: testStatusSaver{}, MessageSender: testSender{}, FileSaver: testSaver{}}).ServeHTTP(resp, req)

			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
	})
}

func TestPOST_NoEmail(t *testing.T) {
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
			NewRouter(&ServiceData{StatusSaver: testStatusSaver{}, MessageSender: testSender{}, FileSaver: testSaver{}}).ServeHTTP(resp, req)

			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
	})
}

func TestPOST_Sender(t *testing.T) {
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
				func(message *messages.Message) error {
					return nil
				}), StatusSaver: testStatusSaver{}, FileSaver: testSaver{}}).ServeHTTP(resp, req)

			Convey("Then the response should be a 200", func() {
				So(resp.Code, ShouldEqual, 200)
			})
		})
	})
}

func TestPOST_SenderFails(t *testing.T) {
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
				func(message *messages.Message) error {
					return errors.New("Can not send")
				}), StatusSaver: testStatusSaver{}, FileSaver: testSaver{}}).ServeHTTP(resp, req)

			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
	})
}

func TestPOST_SaverFails(t *testing.T) {
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
				func(message *messages.Message) error {
					return nil
				}), StatusSaver: testStatusSaver{},
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
			NewRouter(&ServiceData{MessageSender: testSender{},
				StatusSaver: testStatusSaverFunc(
					func(ID string, status string, errorStr string) error {
						return errors.New("Can not send")
					}),
				FileSaver: testSaver{}}).ServeHTTP(resp, req)

			Convey("Then the response should be a 400", func() {
				So(resp.Code, ShouldEqual, 400)
			})
		})
	})
}

type testSenderFunc func(message *messages.Message) error

func (f testSenderFunc) Send(message *messages.Message) error {
	return f(message)
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

func (sender testSender) Send(message *messages.Message) error {
	log.Printf("Sending msg %s\n", message.ID)
	return nil
}

type testStatusSaverFunc func(ID string, status string, errorStr string) error

func (f testStatusSaverFunc) Save(ID string, status string, errorStr string) error {
	return f(ID, status, errorStr)
}

type testStatusSaver struct{}

func (saver testStatusSaver) Save(ID string, status string, errorStr string) error {
	log.Printf("Saving status %s %s\n", ID, status)
	return nil
}
