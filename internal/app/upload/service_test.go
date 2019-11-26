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
	"time"

	"github.com/heptiolabs/healthcheck"
	"github.com/stretchr/testify/assert"

	"bitbucket.org/airenas/listgo/internal/app/upload/api"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks/matchers"

	"encoding/json"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"github.com/gorilla/mux"
	"github.com/petergtz/pegomock"
)

var statusSaverMock *mocks.MockSaver

var requestSaverMock *mocks.MockRequestSaver

var msgSenderMock *mocks.MockSender

var recognizerMapMock *mocks.MockRecognizerMap

var recognizerProviderMock *mocks.MockRecognizerProvider

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	statusSaverMock = mocks.NewMockSaver()
	requestSaverMock = mocks.NewMockRequestSaver()
	msgSenderMock = mocks.NewMockSender()
	recognizerMapMock = mocks.NewMockRecognizerMap()
	recognizerProviderMock = mocks.NewMockRecognizerProvider()
	pegomock.When(recognizerMapMock.Get(pegomock.AnyString())).ThenReturn("recID", nil)
}

func TestWrongPath(t *testing.T) {
	req := httptest.NewRequest("GET", "/invalid", nil)
	testCode(t, req, 404)
}

func TestNoFilePOST(t *testing.T) {
	test400(t, httptest.NewRequest("POST", "/upload", nil))
}

func TestLive(t *testing.T) {
	req := httptest.NewRequest("GET", "/ready", nil)
	testCode(t, req, 200)
}

func TestLive503(t *testing.T) {
	req := httptest.NewRequest("GET", "/live", nil)
	initTest(t)
	resp := httptest.NewRecorder()

	data := newData()
	data.health.AddLivenessCheck("test", func() error { return errors.New("test") })
	NewRouter(data).ServeHTTP(resp, req)

	assert.Equal(t, 503, resp.Code)
}

func TestReady(t *testing.T) {
	req := httptest.NewRequest("GET", "/live", nil)
	testCode(t, req, 200)
}

func TestPOST(t *testing.T) {
	initTest(t)
	req := newReq("filename.wav", "a@a.a", "")
	resp := httptest.NewRecorder()

	newRouter().ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, 200)
	assert.True(t, strings.HasPrefix(resp.Body.String(), `{"id":"`))
}

func TestPOSTNoFile(t *testing.T) {
	test400(t, newReq("", "a@a.a", ""))
}

func newReq4(file string, email string, externalID string, recID string) *http.Request {
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
	if recID != "" {
		writer.WriteField("recognizer", recID)
	}
	writer.Close()
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func newReq(file string, email string, externalID string) *http.Request {
	return newReq4(file, email, externalID, "recKey")
}

func newRouter() *mux.Router {
	return NewRouter(newData())
}

func newData() *ServiceData {
	return &ServiceData{StatusSaver: statusSaverMock,
		MessageSender:      msgSenderMock,
		RequestSaver:       requestSaverMock,
		FileSaver:          testSaver{},
		RecognizerMap:      recognizerMapMock,
		RecognizerProvider: recognizerProviderMock,
		health:             healthcheck.NewHandler(),
	}
}

func test400(t *testing.T, req *http.Request) {
	testCode(t, req, 400)
}

func testCode(t *testing.T, req *http.Request, code int) {
	initTest(t)
	resp := httptest.NewRecorder()

	newRouter().ServeHTTP(resp, req)

	assert.Equal(t, code, resp.Code)
}
func TestPOST_WrongEmail(t *testing.T) {
	test400(t, newReq("file.wav", "a@", ""))
	test400(t, newReq("file.wav", "@a", ""))
	test400(t, newReq("file.wav", "a_a", ""))
}

func TestPOST_EmptyEmail(t *testing.T) {
	testCode(t, newReq("file.wav", "", ""), 200)
}

func TestPOST_Sender(t *testing.T) {
	initTest(t)
	req := newReq("filename.wav", "a@a.a", "")
	resp := httptest.NewRecorder()

	NewRouter(newData()).ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, 200)
}

func Test_Wav(t *testing.T) {
	testCode(t, newReq("file.wav", "a@a.a", ""), 200)
	testCode(t, newReq("file.Wav", "a@a.a", ""), 200)
	testCode(t, newReq("file.txt.Wav", "a@a.a", ""), 200)
}

func Test_Mp3(t *testing.T) {
	testCode(t, newReq("file.mp3", "a@a.a", ""), 200)
	testCode(t, newReq("file.MP3", "a@a.a", ""), 200)
}

func Test_Mp4(t *testing.T) {
	testCode(t, newReq("file.mp4", "a@a.a", ""), 200)
	testCode(t, newReq("file.MP4", "a@a.a", ""), 200)
}

func Test_ExtensionFails(t *testing.T) {
	testCode(t, newReq("file.txt", "a@a.a", ""), 400)
	testCode(t, newReq("file.mp", "a@a.a", ""), 400)
	testCode(t, newReq("file.wave", "a@a.a", ""), 400)
	testCode(t, newReq("file.mp5", "a@a.a", ""), 400)
}

func TestPOST_SenderFails(t *testing.T) {
	initTest(t)
	req := newReq("filename.wav", "a@a.a", "")
	resp := httptest.NewRecorder()
	pegomock.When(msgSenderMock.Send(matchers.AnyMessagesMessage(), pegomock.AnyString(),
		pegomock.AnyString())).ThenReturn(errors.New("Can not send"))

	NewRouter(newData()).ServeHTTP(resp, req)

	assert.Equal(t, 500, resp.Code)
}

func TestPOST_RecognizerMethodFails(t *testing.T) {
	initTest(t)
	req := newReq("filename.wav", "a@a.a", "")
	resp := httptest.NewRecorder()
	pegomock.When(recognizerMapMock.Get(pegomock.AnyString())).ThenReturn("", errors.New("Rec map failed"))

	NewRouter(newData()).ServeHTTP(resp, req)

	assert.Equal(t, 500, resp.Code)
}

func TestPOST_UnknownRecognizerFails(t *testing.T) {
	initTest(t)
	req := newReq("filename.wav", "a@a.a", "")
	resp := httptest.NewRecorder()
	pegomock.When(recognizerMapMock.Get(pegomock.AnyString())).ThenReturn("", api.ErrRecognizerNotFound)

	NewRouter(newData()).ServeHTTP(resp, req)

	assert.Equal(t, 400, resp.Code)
}

func TestPOST_NoRecognizerFails(t *testing.T) {
	initTest(t)
	req := newReq4("filename.wav", "a@a.a", "", "rec123")
	resp := httptest.NewRecorder()
	pegomock.When(recognizerMapMock.Get(pegomock.AnyString())).ThenReturn("", api.ErrRecognizerNotFound)

	NewRouter(newData()).ServeHTTP(resp, req)

	assert.Equal(t, 400, resp.Code)
}

func TestPOST_SaverFails(t *testing.T) {
	initTest(t)
	req := newReq("filename.wav", "a@a.a", "")
	resp := httptest.NewRecorder()

	data := newData()
	data.FileSaver = testSaverFunc(
		func(id string, reader io.Reader) error {
			return errors.New("Can not send")
		})
	NewRouter(data).ServeHTTP(resp, req)

	assert.Equal(t, 500, resp.Code)
}

func TestPOST_StatusSaverFails(t *testing.T) {
	initTest(t)
	req := newReq("filename.wav", "a@a.a", "")
	resp := httptest.NewRecorder()
	pegomock.When(statusSaverMock.Save(pegomock.AnyString(),
		matchers.AnyStatusStatus())).ThenReturn(errors.New("error"))

	newRouter().ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, 500)
}

func TestPOST_RequestSaverFails(t *testing.T) {
	initTest(t)
	req := newReq("filename.wav", "a@a.a", "")
	resp := httptest.NewRecorder()
	pegomock.When(requestSaverMock.Save(matchers.AnyApiRequestData())).ThenReturn(errors.New("error"))

	newRouter().ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, 500)
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
	assert.Equal(t, "recKey", rd.RecognizerKey)
	assert.Equal(t, "recID", rd.RecognizerID)
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

func TestGET_Recognizers(t *testing.T) {
	initTest(t)
	req, _ := http.NewRequest("GET", "/recognizers", nil)
	resp := httptest.NewRecorder()
	var ri []*api.Recognizer
	ttime := time.Now().Truncate(24 * time.Hour)
	ri = append(ri, &api.Recognizer{ID: "ID", Name: "name", Description: "descr", DateCreated: ttime})
	pegomock.When(recognizerProviderMock.GetAll()).ThenReturn(ri, nil)

	newRouter().ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)

	var r []*api.Recognizer
	err := json.Unmarshal(resp.Body.Bytes(), &r)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(r))
	assert.Equal(t, "ID", r[0].ID)
	assert.Equal(t, "name", r[0].Name)
	assert.Equal(t, "descr", r[0].Description)
	assert.Equal(t, ttime, r[0].DateCreated)
}

func TestGET_Recognizers_Fails(t *testing.T) {
	initTest(t)
	req, _ := http.NewRequest("GET", "/recognizers", nil)
	resp := httptest.NewRecorder()
	pegomock.When(recognizerProviderMock.GetAll()).ThenReturn(nil, errors.New("err"))

	newRouter().ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)
}
