package fs

import (
	"bytes"
	"encoding/json"
	"net/http"

	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"
	"github.com/pkg/errors"
)

//Client comunicates with file server
type Client struct {
	httpclient *http.Client
	url        string
}

//NewClient creates a fs client
func NewClient() (*Client, error) {
	res := Client{}
	var err error
	res.url, err = utils.GetURLFromConfig("fs.url")
	if err != nil {
		return nil, err
	}
	res.httpclient = http.DefaultClient

	return &res, nil
}

type getAudioResponse struct {
	ID       string `json:"id"`
	Data     string `json:"data"`
	FileName string `json:"file_name"`
	JobType  string `json:"job_type"`
}

//GetAudio loads audio from fs
func (sp *Client) GetAudio(kafkaID string) (*kafkaapi.DBEntry, error) {
	urlStr := utils.URLJoin(sp.url, "AudioGetRequest", kafkaID)
	cmdapp.Log.Infof("Get audio: %s", urlStr)
	resp, err := sp.httpclient.Get(urlStr)
	if err != nil {
		return nil, err
	}
	err = utils.ValidateResponse(resp)
	if err != nil {
		return nil, errors.Wrap(err, "Can't get audio")
	}
	var respData getAudioResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return nil, errors.Wrap(err, "Can't decode response")
	}
	var result kafkaapi.DBEntry
	result.ID = respData.ID
	result.Data = respData.Data
	result.FileName = respData.FileName
	result.JobType = respData.JobType

	return &result, nil
}

type transcriptionPostRequest struct {
	ID            string        `json:"id"`
	Event         string        `json:"event"`
	Status        string        `json:"status"`
	Error         trError       `json:"error"`
	Transcription transcription `json:"transcription"`
}

type transcription struct {
	Text   string `json:"text"`
	Latice string `json:"lattice"`
}

type trError struct {
	Code         string `json:"code"`
	DebugMessage string `json:"debug_message"`
}

//SaveResult saves result to fs
func (sp *Client) SaveResult(dataIn *kafkaapi.DBResultEntry) error {
	urlStr := utils.URLJoin(sp.url, "TranscriptionPostRequest")

	var data transcriptionPostRequest
	data.ID = dataIn.ID
	data.Event = "AudioTextReady"
	data.Status = dataIn.Status
	if data.Status == "failed" {
		data.Error.Code = dataIn.Err.Code
		data.Error.DebugMessage = dataIn.Err.Error
	} else {
		data.Transcription.Text = dataIn.Transcription.Text
		data.Transcription.Latice = dataIn.Transcription.ResultFileData
	}

	bytesData, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "Can't marshal data")
	}
	cmdapp.Log.Infof("Post result audio: %s", urlStr)
	resp, err := sp.httpclient.Post(urlStr, "application/json", bytes.NewBuffer(bytesData))
	if err != nil {
		return errors.Wrap(err, "Can't send data to file server")
	}
	err = utils.ValidateResponse(resp)
	if err != nil {
		return errors.Wrap(err, "Can't save result audio")
	}
	return nil
}
