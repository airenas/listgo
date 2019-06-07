package transcriberapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"bitbucket.org/airenas/listgo/internal/app/status/api"
	"github.com/stretchr/testify/assert"
)

func initServer(t *testing.T, urlStr, resp string, code int) (*Client, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.URL.String(), urlStr)
		rw.WriteHeader(code)
		rw.Write([]byte(resp))
	}))
	// Use Client & URL from our local test server
	api := Client{}
	api.httpclient = server.Client()
	api.statusURL = server.URL
	api.resultURL = server.URL
	api.uploadURL = server.URL
	return &api, server
}

func TestStatus(t *testing.T) {
	var resp api.TranscriptionResult
	resp.ID = "k10"
	resp.Status = "COMPLETED"
	resp.RecognizedText = "text"
	rb, _ := json.Marshal(resp)
	api, server := initServer(t, "/k10", string(rb), 200)
	defer server.Close()

	r, err := api.GetStatus("k10")

	assert.Nil(t, err)
	assert.Equal(t, r.ID, "k10")
	assert.Equal(t, r.Completed, true)
	assert.Equal(t, r.Text, "text")
	assert.Equal(t, r.ErrorCode, "")
}

func TestStatus_NotCompleted(t *testing.T) {
	var resp api.TranscriptionResult
	resp.ID = "k10"
	resp.Status = "working"
	resp.RecognizedText = "text"
	rb, _ := json.Marshal(resp)
	api, server := initServer(t, "/k10", string(rb), 200)
	defer server.Close()

	r, err := api.GetStatus("k10")

	assert.Nil(t, err)
	assert.Equal(t, r.ID, "k10")
	assert.Equal(t, r.Completed, false)
}

func TestStatus_Failed(t *testing.T) {
	var resp api.TranscriptionResult
	resp.ID = "k10"
	resp.Status = "working"
	resp.RecognizedText = ""
	resp.ErrorCode = "ec"
	resp.Error = "e"
	rb, _ := json.Marshal(resp)
	api, server := initServer(t, "/k10", string(rb), 200)
	defer server.Close()

	r, err := api.GetStatus("k10")

	assert.Nil(t, err)
	assert.Equal(t, r.ID, "k10")
	assert.Equal(t, r.Completed, false)
	assert.Equal(t, r.ErrorCode, "ec")
	assert.Equal(t, r.Error, "e")
}

func TestStatus_WrongCode_Fails(t *testing.T) {
	api, server := initServer(t, "/k10", string(""), 300)
	defer server.Close()

	r, err := api.GetStatus("k10")
	assert.NotNil(t, err)
	assert.Nil(t, r)
}

func TestStatus_WrongJSON_Fails(t *testing.T) {
	api, server := initServer(t, "/k10", string("olia"), 200)
	defer server.Close()

	r, err := api.GetStatus("k10")
	assert.NotNil(t, err)
	assert.Nil(t, r)
}

func TestResult(t *testing.T) {
	api, server := initServer(t, "/result/k10/result.txt", "olia", 200)
	defer server.Close()

	r, err := api.GetResult("k10")

	assert.Nil(t, err)
	assert.Equal(t, r.ID, "k10")
	assert.Equal(t, r.FileData, "b2xpYQ==")
}

func TestResult_WrongCode_Fails(t *testing.T) {
	api, server := initServer(t, "/result/k10/result.txt", "v", 300)
	defer server.Close()

	r, err := api.GetResult("k10")

	assert.NotNil(t, err)
	assert.Nil(t, r)
}

func TestUpload(t *testing.T) {
	api, server := initServer(t, "/", "{\"id\":\"1\"}", 200)
	defer server.Close()

	r, err := api.Upload(&kafkaapi.UploadData{})

	assert.Nil(t, err)
	assert.Equal(t, r, "1")
}

func TestUpload_NoID_Fails(t *testing.T) {
	api, server := initServer(t, "/", "{\"id\":\"\"}", 200)
	defer server.Close()

	r, err := api.Upload(&kafkaapi.UploadData{})

	assert.NotNil(t, err)
	assert.Equal(t, r, "")
}

func TestUpload_WrongCode_Fails(t *testing.T) {
	api, server := initServer(t, "/", "{\"id\":\"1\"}", 300)
	defer server.Close()

	r, err := api.Upload(&kafkaapi.UploadData{})

	assert.NotNil(t, err)
	assert.Equal(t, "", r)
}

func TestUpload_WrongJSON_Fails(t *testing.T) {
	api, server := initServer(t, "/", "olia", 300)
	defer server.Close()

	r, err := api.Upload(&kafkaapi.UploadData{})

	assert.NotNil(t, err)
	assert.Equal(t, r, "")
}
