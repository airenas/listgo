package fs

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
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
	resp.ID = 10
	resp.FileName = "f.name"
	resp.Data = "data"
	resp.JobType = "job"
	resp.NumberOfSpeakers = 2
	resp.RecordQuality = "good"
	rb, _ := json.Marshal(resp)
	api, server := initServer(t, "/audio/10", string(rb), 200)
	defer server.Close()

	r, err := api.GetAudio("10")

	assert.Nil(t, err)
	assert.Equal(t, r.ID, "10")
	assert.Equal(t, r.Data, "data")
	assert.Equal(t, r.FileName, "f.name")
	assert.Equal(t, r.JobType, "job")
	assert.Equal(t, "good", r.RecordQuality)
	assert.Equal(t, "2", r.NumberOfSpeakers)
}

func TestGetAudio_EmptyNumberOfSpeakers(t *testing.T) {
	var resp getAudioResponse
	resp.NumberOfSpeakers = 0
	rb, _ := json.Marshal(resp)
	api, server := initServer(t, "/audio/10", string(rb), 200)
	defer server.Close()

	r, err := api.GetAudio("10")

	assert.Nil(t, err)
	assert.Equal(t, "", r.NumberOfSpeakers)
}

func TestGetAudio_WrongCode_Fails(t *testing.T) {
	api, server := initServer(t, "/audio/10", "", 400)
	defer server.Close()

	r, err := api.GetAudio("10")
	assert.NotNil(t, err)
	assert.Nil(t, r)
}

func TestGetAudio_WrongResp_Fails(t *testing.T) {
	api, server := initServer(t, "/audio/10", "olia", 200)
	defer server.Close()

	r, err := api.GetAudio("10")
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
	dIn.ID = "10"
	dIn.Transcription.Text = "tt"
	dIn.Transcription.LatticeData = "trfd"
	dIn.Transcription.WebVTT = "trvtt"
	r, err := invokeResultPost(t, "/audio/10/transcription", 200, &dIn)

	assert.Nil(t, err)
	assert.Equal(t, r.ID, 10)
	assert.Equal(t, statusDone, r.Status)
	assert.Equal(t, "tt", r.Transcription.Text)
	assert.Equal(t, "trfd", r.Transcription.Latice)
	assert.Equal(t, "trvtt", r.Transcription.WebVTT)
}

func TestSaveResults_ReturnError_Fails(t *testing.T) {
	var dIn kafkaapi.DBResultEntry
	dIn.ID = "10"
	dIn.Transcription.Text = "tt"
	dIn.Transcription.LatticeData = "trfd"
	_, err := invokeResultPost(t, "/audio/10/transcription", 400, &dIn)

	assert.NotNil(t, err)
}

func TestSaveResults_PassError_OK(t *testing.T) {
	var dIn kafkaapi.DBResultEntry
	dIn.ID = "10"
	dIn.Error = &kafkaapi.DBTranscriptionError{Code: "ec", Error: "ee"}
	r, err := invokeResultPost(t, "/audio/10/transcription", 200, &dIn)

	assert.NotNil(t, r)
	assert.Nil(t, err)
	assert.Equal(t, r.ID, 10)
	assert.Equal(t, statusFailed, r.Status)
	assert.Nil(t, r.Transcription)
	assert.Equal(t, "ec", r.Error.Code)
	assert.Equal(t, "ee", r.Error.DebugMessage)
}
