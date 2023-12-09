package transcriberapi

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"github.com/airenas/listgo/internal/app/status/api"
	"github.com/stretchr/testify/assert"
)

type testResp struct {
	code int
	resp string
}

type testReq struct {
	resp string
	URL  string
}

func newTestR(code int, resp string) testResp {
	return testResp{code: code, resp: resp}
}

func newTestReq(req *http.Request) testReq {
	b, _ := ioutil.ReadAll(req.Body)
	return testReq{URL: req.URL.String(), resp: string(b)}
}

func initTestServer(t *testing.T, rData map[string]testResp) (*Client, *httptest.Server, *[]testReq) {
	t.Helper()
	resRequest := make([]testReq, 0)
	rLock := &sync.Mutex{}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rLock.Lock()
		defer rLock.Unlock()
		resRequest = append(resRequest, newTestReq(req))
		resp, f := rData[req.URL.String()]
		if f {
			rw.WriteHeader(resp.code)
			rw.Write([]byte(resp.resp))
		}
	}))
	// Use Client & URL from our local test server
	api := Client{}
	api.httpclient = server.Client()
	api.statusURL = server.URL
	api.resultURL = server.URL
	api.uploadURL = server.URL
	api.cleanURL = server.URL
	return &api, server, &resRequest
}

func testCalled(t *testing.T, URL string, tReq []testReq) {
	assert.GreaterOrEqual(t, len(tReq), 1)
	str := ""
	for _, r := range tReq {
		str = r.URL
		if str == URL {
			return
		}
	}
	assert.Equal(t, URL, str)
}

func TestStatus(t *testing.T) {
	var resp api.TranscriptionResult
	resp.ID = "k10"
	resp.Status = "COMPLETED"
	resp.RecognizedText = "text"
	rb, _ := json.Marshal(resp)
	api, server, tReq := initTestServer(t, map[string]testResp{"/k10": newTestR(200, string(rb))})
	defer server.Close()

	r, err := api.GetStatus("k10")

	assert.Nil(t, err)
	assert.Equal(t, r.ID, "k10")
	assert.Equal(t, r.Completed, true)
	assert.Equal(t, r.Text, "text")
	assert.Equal(t, r.ErrorCode, "")
	testCalled(t, "/k10", *tReq)
}

func TestStatus_NotCompleted(t *testing.T) {
	var resp api.TranscriptionResult
	resp.ID = "k10"
	resp.Status = "working"
	resp.RecognizedText = "text"
	rb, _ := json.Marshal(resp)
	api, server, tReq := initTestServer(t, map[string]testResp{"/k10": newTestR(200, string(rb))})
	defer server.Close()

	r, err := api.GetStatus("k10")

	assert.Nil(t, err)
	assert.Equal(t, r.ID, "k10")
	assert.Equal(t, r.Completed, false)
	testCalled(t, "/k10", *tReq)
}

func TestStatus_Failed(t *testing.T) {
	var resp api.TranscriptionResult
	resp.ID = "k10"
	resp.Status = "working"
	resp.RecognizedText = ""
	resp.ErrorCode = "ec"
	resp.Error = "e"
	rb, _ := json.Marshal(resp)
	api, server, tReq := initTestServer(t, map[string]testResp{"/k10": newTestR(200, string(rb))})
	defer server.Close()

	r, err := api.GetStatus("k10")

	assert.Nil(t, err)
	assert.Equal(t, r.ID, "k10")
	assert.Equal(t, r.Completed, false)
	assert.Equal(t, r.ErrorCode, "ec")
	assert.Equal(t, r.Error, "e")
	testCalled(t, "/k10", *tReq)
}

func TestStatus_WrongCode_Fails(t *testing.T) {
	api, server, tReq := initTestServer(t, map[string]testResp{"/k10": newTestR(300, "")})
	defer server.Close()

	r, err := api.GetStatus("k10")
	assert.NotNil(t, err)
	assert.Nil(t, r)
	testCalled(t, "/k10", *tReq)
}

func TestStatus_WrongJSON_Fails(t *testing.T) {
	api, server, tReq := initTestServer(t, map[string]testResp{"/k10": newTestR(200, "olia")})
	defer server.Close()

	r, err := api.GetStatus("k10")
	assert.NotNil(t, err)
	assert.Nil(t, r)
	testCalled(t, "/k10", *tReq)
}

func TestResult(t *testing.T) {
	api, server, tReq := initTestServer(t, map[string]testResp{"/result/k10/lat.restored.txt": newTestR(200, "olia"),
		"/result/k10/webvtt.txt": newTestR(200, "webvtt")})
	defer server.Close()

	r, err := api.GetResult("k10")

	assert.Nil(t, err)
	assert.Equal(t, r.ID, "k10")
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte("olia")), r.LatticeData)
	assert.Equal(t, "webvtt", r.WebVTTData)
	testCalled(t, "/result/k10/lat.restored.txt", *tReq)
	testCalled(t, "/result/k10/webvtt.txt", *tReq)
}

func TestResult_WrongCode_Fails(t *testing.T) {
	api, server, tReq := initTestServer(t, map[string]testResp{"/result/k10/lat.restored.txt": newTestR(300, "v"),
		"/result/k10/webvtt.txt": newTestR(300, "webvtt")})
	defer server.Close()

	r, err := api.GetResult("k10")

	assert.NotNil(t, err)
	assert.Nil(t, r)
	testCalled(t, "/result/k10/lat.restored.txt", *tReq)
	testCalled(t, "/result/k10/webvtt.txt", *tReq)
}

func TestResult_WrongCode_Lat_Fails(t *testing.T) {
	api, server, tReq := initTestServer(t, map[string]testResp{"/result/k10/lat.restored.txt": newTestR(300, "olia"),
		"/result/k10/webvtt.txt": newTestR(200, "webvtt")})
	defer server.Close()
	r, err := api.GetResult("k10")

	assert.NotNil(t, err)
	assert.Nil(t, r)
	testCalled(t, "/result/k10/lat.restored.txt", *tReq)
	testCalled(t, "/result/k10/webvtt.txt", *tReq)
}

func TestResult_WrongCode_WebVTT_Fails(t *testing.T) {
	api, server, tReq := initTestServer(t, map[string]testResp{"/result/k10/lat.restored.txt": newTestR(200, "olia"),
		"/result/k10/webvtt.txt": newTestR(300, "webvtt")})
	defer server.Close()
	r, err := api.GetResult("k10")

	assert.NotNil(t, err)
	assert.Nil(t, r)
	testCalled(t, "/result/k10/lat.restored.txt", *tReq)
	testCalled(t, "/result/k10/webvtt.txt", *tReq)
}

func TestUpload(t *testing.T) {
	api, server, tReq := initTestServer(t, map[string]testResp{"/": newTestR(200, "{\"id\":\"1\"}")})
	defer server.Close()

	r, err := api.Upload(&kafkaapi.UploadData{})

	assert.Nil(t, err)
	assert.Equal(t, r, "1")
	testCalled(t, "/", *tReq)
}

func TestUpload_NoID_Fails(t *testing.T) {
	api, server, tReq := initTestServer(t, map[string]testResp{"/": newTestR(200, "{\"id\":\"\"}")})
	defer server.Close()

	r, err := api.Upload(&kafkaapi.UploadData{})

	assert.NotNil(t, err)
	assert.Equal(t, r, "")
	testCalled(t, "/", *tReq)
}

func TestUpload_WrongCode_Fails(t *testing.T) {
	api, server, tReq := initTestServer(t, map[string]testResp{"/": newTestR(300, "{\"id\":\"1\"}")})
	defer server.Close()

	r, err := api.Upload(&kafkaapi.UploadData{})

	assert.NotNil(t, err)
	assert.Equal(t, "", r)
	testCalled(t, "/", *tReq)
}

func TestUpload_WrongJSON_Fails(t *testing.T) {
	api, server, tReq := initTestServer(t, map[string]testResp{"/": newTestR(300, "olia")})
	defer server.Close()

	r, err := api.Upload(&kafkaapi.UploadData{})

	assert.NotNil(t, err)
	assert.Equal(t, r, "")
	testCalled(t, "/", *tReq)
}

func TestUpload_PassNumberOfSpeakers(t *testing.T) {
	api, server, tReq := initTestServer(t, map[string]testResp{"/": newTestR(300, "olia")})
	defer server.Close()

	r, err := api.Upload(&kafkaapi.UploadData{NumberOfSpeakers: "__numberOfSpeakers__"})

	assert.NotNil(t, err)
	assert.Equal(t, "", r)
	testCalled(t, "/", *tReq)
	bs := (*tReq)[0].resp
	assert.Contains(t, bs, "numberOfSpeakers")
	assert.Contains(t, bs, "__numberOfSpeakers__")
}

func TestUpload_PassRecognizer(t *testing.T) {
	api, server, tReq := initTestServer(t, map[string]testResp{"/": newTestR(300, "olia")})
	defer server.Close()

	r, err := api.Upload(&kafkaapi.UploadData{JobType: "law", RecordQuality: "standard"})

	assert.NotNil(t, err)
	assert.Equal(t, r, "")
	testCalled(t, "/", *tReq)
	bs := (*tReq)[0].resp
	assert.Contains(t, bs, "recognizer")
	assert.Contains(t, bs, "law_standard")
}

func TestDelete(t *testing.T) {
	api, server, tReq := initTestServer(t, map[string]testResp{"/10": newTestR(200, "OK")})
	defer server.Close()

	err := api.Delete("10")

	assert.Nil(t, err)
	testCalled(t, "/10", *tReq)
}

func TestDelete_Fails(t *testing.T) {
	api, server, tReq := initTestServer(t, map[string]testResp{"/10": newTestR(500, "Error")})
	defer server.Close()

	err := api.Delete("10")

	assert.NotNil(t, err)
	testCalled(t, "/10", *tReq)
}
