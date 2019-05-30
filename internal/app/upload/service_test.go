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

	"github.com/stretchr/testify/assert"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"github.com/gorilla/mux"
	"github.com/petergtz/pegomock"
)

var statusSaverMock *mocks.MockSaver

var requestSaverMock *mocks.MockRequestSaver

var msgSenderMock *mocks.MockSender

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	statusSaverMock = mocks.NewMockSaver()
	requestSaverMock = mocks.NewMockRequestSaver()
	msgSenderMock = mocks.NewMockSender()
}

func TestWrongPath(t *testing.T) {
	req := httptest.NewRequest("GET", "/invalid", nil)
	testCode(t, req, 404)
}

func TestNoFilePOST(t *testing.T) {
	test400(t, httptest.NewRequest("POST", "/upload", nil))
}

func TestPOST(t *testing.T) {
	initTest(t)
	req := newReq("filename", "a@a.a", "")
	resp := httptest.NewRecorder()

	newRouter().ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, 200)
	assert.True(t, strings.HasPrefix(resp.Body.String(), `{"id":"`))
}

func TestPOSTNoFile(t *testing.T) {
	test400(t, newReq("", "a@a.a", ""))
}

func newReq(file string, email string, externalID string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if file != "" {
		part, _ := writer.CreateFormFile("file", file)
		_, _ = io.Copy(part, strings.NewReader("body"))
	}
	if email != "" {
		writer.WriteField("email", email)
	}
	if externalID != "" {
		writer.WriteField("externalID", externalID)
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

func test400(t *testing.T, req *http.Request) {
	testCode(t, req, 400)
}

func testCode(t *testing.T, req *http.Request, code int) {
	initTest(t)
	resp := httptest.NewRecorder()

	newRouter().ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, code)
}
func TestPOST_WrongEmail(t *testing.T) {
	test400(t, newReq("file", "a@", ""))
	test400(t, newReq("file", "@a", ""))
	test400(t, newReq("file", "a_a", ""))
}

func TestPOST_EmptyEmail(t *testing.T) {
	testCode(t, newReq("file", "", ""), 200)
}

func TestPOST_Sender(t *testing.T) {
	initTest(t)
	req := newReq("filename", "a@a.a", "")
	resp := httptest.NewRecorder()

	NewRouter(&ServiceData{MessageSender: msgSenderMock, StatusSaver: statusSaverMock,
		RequestSaver: requestSaverMock,
		FileSaver:    testSaver{}}).ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, 200)
}

func TestPOST_SenderFails(t *testing.T) {
	initTest(t)
	req := newReq("filename", "a@a.a", "")
	resp := httptest.NewRecorder()
	pegomock.When(msgSenderMock.Send(matchers.AnyMessagesMessage(), pegomock.AnyString(),
		pegomock.AnyString())).ThenReturn(errors.New("Can not send"))

	NewRouter(&ServiceData{MessageSender: msgSenderMock,
		StatusSaver:  statusSaverMock,
		RequestSaver: requestSaverMock,
		FileSaver:    testSaver{}}).ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, 400)
}

func TestPOST_SaverFails(t *testing.T) {
	initTest(t)
	req := newReq("filename", "a@a.a", "")
	resp := httptest.NewRecorder()

	NewRouter(&ServiceData{MessageSender: msgSenderMock, StatusSaver: statusSaverMock,
		RequestSaver: requestSaverMock,
		FileSaver: testSaverFunc(
			func(id string, reader io.Reader) error {
				return errors.New("Can not send")
			})}).ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, 400)
}

func TestPOST_StatusSaverFails(t *testing.T) {
	initTest(t)
	req := newReq("filename", "a@a.a", "")
	resp := httptest.NewRecorder()
	pegomock.When(statusSaverMock.Save(pegomock.AnyString(),
		matchers.AnyStatusStatus())).ThenReturn(errors.New("error"))

	newRouter().ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, 400)
}

func TestPOST_RequestSaverFails(t *testing.T) {
	initTest(t)
	req := newReq("filename", "a@a.a", "")
	resp := httptest.NewRecorder()
	pegomock.When(requestSaverMock.Save(matchers.AnyApiRequestData())).ThenReturn(errors.New("error"))

	newRouter().ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, 400)
}

func TestPOST_RequestSaverCalled(t *testing.T) {
	initTest(t)
	req := newReq("filename.wav", "a@a.a", "externalID")
	resp := httptest.NewRecorder()
	pegomock.When(requestSaverMock.Save(matchers.AnyApiRequestData())).ThenReturn(nil)

	newRouter().ServeHTTP(resp, req)

	rd := requestSaverMock.VerifyWasCalled(pegomock.Once()).Save(matchers.AnyApiRequestData()).GetCapturedArguments()
	assert.Equal(t, rd.Email, "a@a.a")
	assert.Equal(t, rd.ExternalID, "externalID")
	assert.True(t, strings.HasSuffix(rd.File, ".wav"))
	assert.NotEmpty(t, rd.ID)
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
