package fs

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
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
	api.url = server.URL
	return &api, server
}

func TestGetAudio(t *testing.T) {
	var resp getAudioResponse
	resp.ID = "k10"
	resp.FileName = "f.name"
	resp.Data = "data"
	resp.JobType = "job"
	rb, _ := json.Marshal(resp)
	api, server := initServer(t, "/AudioGetRequest/k10", string(rb), 200)
	defer server.Close()

	r, err := api.GetAudio("k10")

	assert.Nil(t, err)
	assert.Equal(t, r.ID, "k10")
	assert.Equal(t, r.Data, "data")
	assert.Equal(t, r.FileName, "f.name")
	assert.Equal(t, r.JobType, "job")
}

func TestGetAudio_WrongCode_Fails(t *testing.T) {
	api, server := initServer(t, "/AudioGetRequest/k10", "", 400)
	defer server.Close()

	r, err := api.GetAudio("k10")
	assert.NotNil(t, err)
	assert.Nil(t, r)
}

func TestGetAudio_WrongResp_Fails(t *testing.T) {
	api, server := initServer(t, "/AudioGetRequest/k10", "olia", 200)
	defer server.Close()

	r, err := api.GetAudio("k10")
	assert.NotNil(t, err)
	assert.Nil(t, r)
}

func invokeResultPost(t *testing.T, urlStr string, code int, dataIn *kafkaapi.DBResultEntry) (*transcriptionPostRequest, error) {
	var res transcriptionPostRequest
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.URL.String(), urlStr)
		body, _ := ioutil.ReadAll(req.Body)
		json.Unmarshal(body, &res)
		rw.WriteHeader(code)
	}))
	defer server.Close()
	// Use Client & URL from our local test server
	api := Client{}
	api.httpclient = server.Client()
	api.url = server.URL
	err := api.SaveResult(dataIn)
	return &res, err
}

func TestSaveResults(t *testing.T) {
	var dIn kafkaapi.DBResultEntry
	dIn.ID = "k10"
	dIn.Status = "success"
	dIn.Transcription.Text = "tt"
	dIn.Transcription.ResultFileData = "trfd"
	r, err := invokeResultPost(t, "/TranscriptionPostRequest", 200, &dIn)

	assert.Nil(t, err)
	assert.Equal(t, r.ID, "k10")
	assert.Equal(t, r.Status, "success")
	assert.Equal(t, r.Transcription.Text, "tt")
	assert.Equal(t, r.Transcription.Latice, "trfd")
}

func TestSaveResults_ReturnError_Fails(t *testing.T) {
	var dIn kafkaapi.DBResultEntry
	dIn.ID = "k10"
	dIn.Status = "success"
	dIn.Transcription.Text = "tt"
	dIn.Transcription.ResultFileData = "trfd"
	_, err := invokeResultPost(t, "/TranscriptionPostRequest", 400, &dIn)

	assert.NotNil(t, err)
}

func TestSaveResults_PassError_OK(t *testing.T) {
	var dIn kafkaapi.DBResultEntry
	dIn.ID = "k10"
	dIn.Status = "failed"
	dIn.Err.Code = "ec"
	dIn.Err.Error = "ee"
	r, err := invokeResultPost(t, "/TranscriptionPostRequest", 200, &dIn)

	assert.NotNil(t, r)
	assert.Nil(t, err)
	assert.Equal(t, r.ID, "k10")
	assert.Equal(t, r.Status, "failed")
	assert.Equal(t, r.Transcription.Text, "")
	assert.Equal(t, r.Transcription.Latice, "")
	assert.Equal(t, r.Error.Code, "ec")
	assert.Equal(t, r.Error.DebugMessage, "ee")
}
