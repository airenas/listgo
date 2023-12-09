package upload

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/airenas/listgo/internal/app/upload/api"
	"github.com/airenas/listgo/internal/pkg/messages"
	"github.com/airenas/listgo/internal/pkg/test/mocks"
	"github.com/airenas/listgo/internal/pkg/test/mocks/matchers"
	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

var statusSaverMock *mocks.MockSaver

var fileSaverMock *mocks.MockFileSaver

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
	fileSaverMock = mocks.NewMockFileSaver()
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

	data := newTestData()
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

	newTestRouter().ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, 200)
	assert.True(t, strings.HasPrefix(resp.Body.String(), `{"id":"`))
}

func TestPOSTNoFile(t *testing.T) {
	test400(t, newReq("", "a@a.a", ""))
}

func newReq4(file string, email string, externalID string, recID string) *http.Request {
	return newReqMap([]string{file}, map[string]string{"email": email,
		"externalID": externalID, "recognizer": recID})
}

func newReqMap(files []string, values map[string]string) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for i, file := range files {
		suf := ""
		if i > 0 {
			suf = strconv.Itoa(i + 1)
		}
		if file != "" {
			part, _ := writer.CreateFormFile("file"+suf, file)
			_, _ = io.Copy(part, strings.NewReader("body"))
		}
	}
	for k, v := range values {
		if k != "" {
			writer.WriteField(k, v)
		}
	}
	writer.Close()
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func newReq(file string, email string, externalID string) *http.Request {
	return newReq4(file, email, externalID, "recKey")
}

func newTestRouter() *mux.Router {
	return NewRouter(newTestData())
}

func newTestData() *ServiceData {
	res := &ServiceData{StatusSaver: statusSaverMock,
		MessageSender:      msgSenderMock,
		RequestSaver:       requestSaverMock,
		FileSaver:          fileSaverMock,
		RecognizerMap:      recognizerMapMock,
		RecognizerProvider: recognizerProviderMock,
		health:             healthcheck.NewHandler(),
	}
	initMetrics(res)
	return res
}

func test400(t *testing.T, req *http.Request) {
	testCode(t, req, 400)
}

func testCode(t *testing.T, req *http.Request, code int) {
	initTest(t)
	resp := httptest.NewRecorder()

	newTestRouter().ServeHTTP(resp, req)

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

	NewRouter(newTestData()).ServeHTTP(resp, req)

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

func Test_M4a(t *testing.T) {
	testCode(t, newReq("file.m4a", "a@a.a", ""), 200)
	testCode(t, newReq("file.M4a", "a@a.a", ""), 200)
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

	NewRouter(newTestData()).ServeHTTP(resp, req)

	assert.Equal(t, 500, resp.Code)
}

func TestPOST_RecognizerMethodFails(t *testing.T) {
	initTest(t)
	req := newReq("filename.wav", "a@a.a", "")
	resp := httptest.NewRecorder()
	pegomock.When(recognizerMapMock.Get(pegomock.AnyString())).ThenReturn("", errors.New("Rec map failed"))

	NewRouter(newTestData()).ServeHTTP(resp, req)

	assert.Equal(t, 500, resp.Code)
}

func TestPOST_UnknownRecognizerFails(t *testing.T) {
	initTest(t)
	req := newReq("filename.wav", "a@a.a", "")
	resp := httptest.NewRecorder()
	pegomock.When(recognizerMapMock.Get(pegomock.AnyString())).ThenReturn("", api.ErrRecognizerNotFound)

	NewRouter(newTestData()).ServeHTTP(resp, req)

	assert.Equal(t, 400, resp.Code)
}

func TestPOST_NoRecognizerFails(t *testing.T) {
	initTest(t)
	req := newReq4("filename.wav", "a@a.a", "", "rec123")
	resp := httptest.NewRecorder()
	pegomock.When(recognizerMapMock.Get(pegomock.AnyString())).ThenReturn("", api.ErrRecognizerNotFound)

	NewRouter(newTestData()).ServeHTTP(resp, req)

	assert.Equal(t, 400, resp.Code)
}

func TestPOST_SaverFails(t *testing.T) {
	initTest(t)
	req := newReq("filename.wav", "a@a.a", "")
	resp := httptest.NewRecorder()

	pegomock.When(fileSaverMock.Save(pegomock.AnyString(), matchers.AnyIoReader())).ThenReturn(errors.New("error"))

	data := newTestData()
	NewRouter(data).ServeHTTP(resp, req)

	assert.Equal(t, 500, resp.Code)
}

func TestPOST_StatusSaverFails(t *testing.T) {
	initTest(t)
	req := newReq("filename.wav", "a@a.a", "")
	resp := httptest.NewRecorder()
	pegomock.When(statusSaverMock.SaveF(pegomock.AnyString(),
		matchers.AnyMapOfStringToInterface(), matchers.AnyMapOfStringToInterface())).ThenReturn(errors.New("error"))

	newTestRouter().ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, 500)
}

func TestPOST_RequestSaverFails(t *testing.T) {
	initTest(t)
	req := newReq("filename.wav", "a@a.a", "")
	resp := httptest.NewRecorder()
	pegomock.When(requestSaverMock.Save(matchers.AnyPtrToPersistenceRequest())).ThenReturn(errors.New("error"))

	newTestRouter().ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, 500)
}

func TestPOST_RequestSaverCalled(t *testing.T) {
	initTest(t)
	req := newReq("filename.wav", "a@a.a", "externalID")
	resp := httptest.NewRecorder()
	pegomock.When(requestSaverMock.Save(matchers.AnyPtrToPersistenceRequest())).ThenReturn(nil)

	newTestRouter().ServeHTTP(resp, req)

	rd := requestSaverMock.VerifyWasCalled(pegomock.Once()).Save(matchers.AnyPtrToPersistenceRequest()).GetCapturedArguments()
	assert.Equal(t, rd.Email, "a@a.a")
	assert.Equal(t, rd.ExternalID, "externalID")
	assert.Equal(t, "recKey", rd.RecognizerKey)
	assert.Equal(t, "recID", rd.RecognizerID)
	assert.True(t, strings.HasSuffix(rd.File, ".wav"))
	assert.NotEmpty(t, rd.ID)
}

func TestPOST_RequestSaverMultiFiles(t *testing.T) {
	initTest(t)
	req := newReqMap([]string{"file.wav", "file2.wav", "olia.mp4"}, map[string]string{"email": "a@a.lt",
		"recognizer": "rec"})
	resp := httptest.NewRecorder()
	pegomock.When(requestSaverMock.Save(matchers.AnyPtrToPersistenceRequest())).ThenReturn(nil)

	newTestRouter().ServeHTTP(resp, req)

	rd := requestSaverMock.VerifyWasCalled(pegomock.Once()).Save(matchers.AnyPtrToPersistenceRequest()).GetCapturedArguments()
	assert.Equal(t, rd.Email, "a@a.lt")
	assert.Equal(t, "rec", rd.RecognizerKey)
	assert.Equal(t, "recID", rd.RecognizerID)
	assert.True(t, strings.HasSuffix(rd.File, ""))
	assert.NotEmpty(t, rd.ID)
	files, _ := fileSaverMock.VerifyWasCalled(pegomock.Times(3)).Save(pegomock.AnyString(), matchers.AnyIoReader()).
		GetAllCapturedArguments()
	if assert.Equal(t, 3, len(files)) {
		assert.Equal(t, rd.ID+"/file.wav", files[0])
		assert.Equal(t, rd.ID+"/file2.wav", files[1])
		assert.Equal(t, rd.ID+"/olia.mp4", files[2])
	}
	_, q, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(),
		pegomock.AnyString()).GetCapturedArguments()
	assert.Equal(t, messages.DecodeMultiple, q)
}

func TestPOST_NumberOfSpeakersPassed(t *testing.T) {
	initTest(t)
	req := newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt",
		"recognizer": "rec", "numberOfSpeakers": "2"})
	resp := httptest.NewRecorder()
	newTestRouter().ServeHTTP(resp, req)

	msg, q, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(),
		pegomock.AnyString()).GetCapturedArguments()

	assert.Equal(t, messages.Decode, q)
	qmsg, ok := msg.(*messages.QueueMessage)
	assert.True(t, ok)
	assert.Equal(t, messages.Decode, q)
	assert.NotNil(t, qmsg)
	assert.NotNil(t, qmsg.Tags)
	assert.Equal(t, "2", getTag(qmsg.Tags, messages.TagNumberOfSpeakers))
}

func TestPOST_SkipNumJoinPassed(t *testing.T) {
	initTest(t)
	req := newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt",
		"recognizer": "rec", api.PrmSkipNumJoin: "1"})
	resp := httptest.NewRecorder()
	newTestRouter().ServeHTTP(resp, req)

	msg, q, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(),
		pegomock.AnyString()).GetCapturedArguments()

	assert.Equal(t, messages.Decode, q)
	qmsg, ok := msg.(*messages.QueueMessage)
	assert.True(t, ok)
	assert.Equal(t, messages.Decode, q)
	assert.NotNil(t, qmsg)
	assert.NotNil(t, qmsg.Tags)
	assert.Equal(t, "1", getTag(qmsg.Tags, messages.TagSkipNumJoin))
}

func TestPOST_TimestampAdded(t *testing.T) {
	initTest(t)
	req := newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt",
		"recognizer": "rec"})
	resp := httptest.NewRecorder()
	newTestRouter().ServeHTTP(resp, req)

	msg, q, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(),
		pegomock.AnyString()).GetCapturedArguments()

	assert.Equal(t, messages.Decode, q)
	qmsg, ok := msg.(*messages.QueueMessage)
	assert.True(t, ok)
	assert.Equal(t, messages.Decode, q)
	assert.NotNil(t, qmsg)
	assert.NotNil(t, qmsg.Tags)
	s := getTag(qmsg.Tags, messages.TagTimestamp)
	ti, _ := strconv.Atoi(s)
	sTime := time.Unix(int64(ti), 0)
	assert.True(t, time.Now().Add(time.Second).After(sTime))
	assert.True(t, time.Now().Add(-time.Second).Before(sTime))
}

func TestPOST_NumberOfSpeakersNotPassed(t *testing.T) {
	initTest(t)
	req := newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt",
		"recognizer": "rec"})
	resp := httptest.NewRecorder()
	newTestRouter().ServeHTTP(resp, req)

	msg, q, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(),
		pegomock.AnyString()).GetCapturedArguments()

	assert.Equal(t, messages.Decode, q)
	qmsg, ok := msg.(*messages.QueueMessage)
	assert.True(t, ok)
	assert.NotNil(t, qmsg)
	assert.NotNil(t, qmsg.Tags)
	assert.Equal(t, "", getTag(qmsg.Tags, messages.TagNumberOfSpeakers))
}

func TestPOST_SkipNumJoinNotPassed(t *testing.T) {
	initTest(t)
	req := newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt",
		"recognizer": "rec"})
	resp := httptest.NewRecorder()
	newTestRouter().ServeHTTP(resp, req)

	msg, q, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(),
		pegomock.AnyString()).GetCapturedArguments()

	assert.Equal(t, messages.Decode, q)
	qmsg, ok := msg.(*messages.QueueMessage)
	assert.True(t, ok)
	assert.NotNil(t, qmsg)
	assert.NotNil(t, qmsg.Tags)
	assert.Equal(t, "", getTag(qmsg.Tags, messages.TagSkipNumJoin))
}

func TestPOST_SpOnChannelPassed(t *testing.T) {
	initTest(t)
	req := newReqMap([]string{"file.wav"}, map[string]string{"sepSpeakersOnChannel": "1"})
	resp := httptest.NewRecorder()
	newTestRouter().ServeHTTP(resp, req)

	msg, q, _ := msgSenderMock.VerifyWasCalled(pegomock.Once()).Send(matchers.AnyMessagesMessage(), pegomock.AnyString(),
		pegomock.AnyString()).GetCapturedArguments()

	assert.Equal(t, messages.Decode, q)
	qmsg, ok := msg.(*messages.QueueMessage)
	assert.True(t, ok)
	assert.NotNil(t, qmsg)
	assert.NotNil(t, qmsg.Tags)
	assert.Equal(t, "1", getTag(qmsg.Tags, messages.TagSepSpeakersOnChannel))
}

func TestPOST_SpOnChannelPassedFail(t *testing.T) {
	req := newReqMap([]string{"file.wav", "file1.wav"}, map[string]string{"sepSpeakersOnChannel": "1"})
	testCode(t, req, http.StatusBadRequest)
}

func TestPOST_FailOnUnknownFormData(t *testing.T) {
	testCode(t, newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt", "recognizer": "rec"}), 200)
	testCode(t, newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt", "rec": "rec"}), 400)
	testCode(t, newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt", "numSpeak": "rec"}), 400)
}

func TestPOST_FailOnWrongNumberOfSpeakers(t *testing.T) {
	testCode(t, newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt", api.PrmNumberOfSpeakers: "olia"}), 200)
	testCode(t, newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt", api.PrmNumberOfSpeakers: "$ olia"}), 400)
	testCode(t, newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt", api.PrmNumberOfSpeakers: "shell olia"}), 400)
	testCode(t, newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt", api.PrmNumberOfSpeakers: "eval olia"}), 400)
	testCode(t, newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt", api.PrmNumberOfSpeakers: "(olia"}), 400)
	testCode(t, newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt", api.PrmNumberOfSpeakers: "olia)"}), 400)
}

func TestPOST_FailOnWrongSkipNumJoin(t *testing.T) {
	testCode(t, newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt", api.PrmSkipNumJoin: "1"}), 200)
	testCode(t, newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt", api.PrmNumberOfSpeakers: "$ olia"}), 400)
	testCode(t, newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt", api.PrmSkipNumJoin: "shell olia"}), 400)
	testCode(t, newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt", api.PrmSkipNumJoin: "eval olia"}), 400)
	testCode(t, newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt", api.PrmSkipNumJoin: "(olia"}), 400)
	testCode(t, newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt", api.PrmSkipNumJoin: "olia)"}), 400)
	testCode(t, newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt", api.PrmSkipNumJoin: "{olia"}), 400)
	testCode(t, newReqMap([]string{"file.wav"}, map[string]string{"email": "a@a.lt", api.PrmSkipNumJoin: "olia}"}), 400)
}

func getTag(tags []messages.Tag, key string) string {
	for _, t := range tags {
		if t.Key == key {
			return t.Value
		}
	}
	return ""
}

func TestGET_Recognizers(t *testing.T) {
	initTest(t)
	req, _ := http.NewRequest("GET", "/recognizers", nil)
	resp := httptest.NewRecorder()
	var ri []*api.Recognizer
	ttime := time.Now().Truncate(24 * time.Hour)
	ri = append(ri, &api.Recognizer{ID: "ID", Name: "name", Description: "descr", DateCreated: ttime})
	pegomock.When(recognizerProviderMock.GetAll()).ThenReturn(ri, nil)

	newTestRouter().ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)

	var r []*api.Recognizer
	err := json.Unmarshal(resp.Body.Bytes(), &r)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(r))
	assert.Equal(t, "ID", r[0].ID)
	assert.Equal(t, "name", r[0].Name)
	assert.Equal(t, "descr", r[0].Description)
	assert.Equal(t, ttime.UTC(), r[0].DateCreated.UTC())
}

func TestGET_Recognizers_Fails(t *testing.T) {
	initTest(t)
	req, _ := http.NewRequest("GET", "/recognizers", nil)
	resp := httptest.NewRecorder()
	pegomock.When(recognizerProviderMock.GetAll()).ThenReturn(nil, errors.New("err"))

	newTestRouter().ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)
}

func TestRecognizers_MetricCollected(t *testing.T) {
	initTest(t)
	req, _ := http.NewRequest("GET", "/recognizers", nil)
	var ri []*api.Recognizer
	ri = append(ri, &api.Recognizer{ID: "ID", Name: "name", Description: "descr", DateCreated: time.Now()})
	pegomock.When(recognizerProviderMock.GetAll()).ThenReturn(ri, nil)
	resp := httptest.NewRecorder()
	data := newTestData()
	NewRouter(data).ServeHTTP(resp, req)
	assert.Equal(t, 1, testutil.CollectAndCount(data.metrics.recResponseDur))
}

func TestPOST_MetricsCollected(t *testing.T) {
	initTest(t)
	req := newReq("filename.wav", "a@a.a", "")
	resp := httptest.NewRecorder()

	data := newTestData()
	NewRouter(data).ServeHTTP(resp, req)
	assert.Equal(t, 1, testutil.CollectAndCount(data.metrics.uploadResponseDur))
	assert.Equal(t, 1, testutil.CollectAndCount(data.metrics.uploadRequestSize))
	assert.Equal(t, 0, testutil.CollectAndCount(data.metrics.recResponseDur))
}

func TestMetrics(t *testing.T) {
	req, _ := http.NewRequest("GET", "/metrics", nil)
	testCode(t, req, 200)
}

func TestToLowerExt(t *testing.T) {
	assert.Equal(t, "olia.wav", toLowerExt("olia.wav"))
	assert.Equal(t, "olia.wav", toLowerExt("olia.WAV"))
	assert.Equal(t, "OLIA.wav", toLowerExt("OLIA.WAV"))
	assert.Equal(t, "ipdasidpasidp/olia.wav", toLowerExt("ipdasidpasidp/olia.Wav"))
}

func TestValidateHeaders(t *testing.T) {
	assert.NotNil(t, validateFiles([]*multipart.FileHeader{{Filename: "oo.ooo"}}))
	assert.NotNil(t, validateFiles([]*multipart.FileHeader{{Filename: "oo.wav"}, {Filename: "oo.ooo"}}))
	assert.NotNil(t, validateFiles([]*multipart.FileHeader{{Filename: "../oo.wav"}}))
	assert.NotNil(t, validateFiles([]*multipart.FileHeader{{Filename: "oo.wav"}, {Filename: "../oo.wav"}}))

	assert.Nil(t, validateFiles([]*multipart.FileHeader{{Filename: "oo.wav"}}))
	assert.Nil(t, validateFiles([]*multipart.FileHeader{{Filename: "oo.mp4"}, {Filename: "o1.mp4"}}))
}

func TestValidateFormFiles(t *testing.T) {
	assert.NotNil(t, validateFormFiles(&multipart.Form{File: map[string][]*multipart.FileHeader{"file2": {}}}))
	assert.NotNil(t, validateFormFiles(&multipart.Form{File: map[string][]*multipart.FileHeader{"file": {}, "file1": {}}}))
	assert.NotNil(t, validateFormFiles(&multipart.Form{File: map[string][]*multipart.FileHeader{"file": {}, "file2": {}, "file4": {}}}))

	assert.Nil(t, validateFormFiles(&multipart.Form{File: map[string][]*multipart.FileHeader{"file": {}}}))
	assert.Nil(t, validateFormFiles(&multipart.Form{File: map[string][]*multipart.FileHeader{"file": {}, "file2": {}}}))
	assert.Nil(t, validateFormFiles(&multipart.Form{File: map[string][]*multipart.FileHeader{"file": {}, "file2": {},
		"file3": {}, "file4": {}}}))
}

func Test(t *testing.T) {
	assert.Equal(t, "olia.txt", filepath.Base("../../olia.txt"))
	assert.Equal(t, "olia.txt", filepath.Base("olia.txt"))
	assert.Equal(t, "olia.txt", filepath.Base("aaa/../../olia.txt"))
	assert.Equal(t, "olia.txt", filepath.Base("/home/aaa/aaa/../../olia.txt"))

	assert.Equal(t, "olia.txt", filepath.Clean("olia.txt"))
	assert.Equal(t, "../olia.txt", filepath.Clean("aaa/../../olia.txt"))
	assert.Equal(t, "/home/olia.txt", filepath.Clean("/home/aaa/aaa/../../olia.txt"))
	assert.Equal(t, "../../olia.txt", filepath.Clean("../../olia.txt"))
}

func Test_sanitizeName(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{name: "Good", args: "file.wav", want: "file.wav"},
		{name: "Space", args: "file 2.wav", want: "file_2.wav"},
		{name: "Several spaces", args: "file  2 2.wav", want: "file__2_2.wav"},
		{name: "basePaths", args: "../../file  2 2.wav", want: "file__2_2.wav"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeName(tt.args); got != tt.want {
				t.Errorf("sanitizeName() = %v, want %v", got, tt.want)
			}
		})
	}
}
