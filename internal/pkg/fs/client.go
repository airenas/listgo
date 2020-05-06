package fs

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

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
	ID               int    `json:"id"`
	Data             string `json:"data"`
	FileName         string `json:"file_name"`
	JobType          string `json:"job_type"`
	NumberOfSpeakers int    `json:"number_of_speakers"`
	RecordQuality    string `json:"record_qualityid"`
}

//GetAudio loads audio from fs
func (sp *Client) GetAudio(kafkaID string) (*kafkaapi.DBEntry, error) {
	urlStr := utils.URLJoin(sp.url, "audio", kafkaID)
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
	result := &kafkaapi.DBEntry{ID: strconv.Itoa(respData.ID),
		Data: respData.Data, FileName: respData.FileName, JobType: respData.JobType,
		RecordQuality: respData.RecordQuality, NumberOfSpeakers: convert(respData.NumberOfSpeakers)}
	return result, nil
}

func convert(i int) string {
	if i == 0 {
		return ""
	}
	return strconv.Itoa(i)
}

const (
	statusFailed = "failed"
	statusDone   = "done"
)

type transcriptionPostRequest struct {
	ID            int            `json:"id"`
	Event         string         `json:"event"`
	Status        string         `json:"status"`
	Error         *trError       `json:"error,omitempty"`
	Transcription *transcription `json:"transcription,omitempty"`
}

type transcription struct {
	Text   string `json:"text"`
	Latice string `json:"lattice,omitempty"`
	WebVTT string `json:"web_vtt,omitempty"`
}

type trError struct {
	Code         string `json:"code"`
	DebugMessage string `json:"debug_message"`
}

//SaveResult saves result to fs
func (sp *Client) SaveResult(dataIn *kafkaapi.DBResultEntry) error {
	urlStr := utils.URLJoin(sp.url, "audio", dataIn.ID, "transcription")
	cmdapp.Log.Infof("Post audio: %s", urlStr)
	var data transcriptionPostRequest
	var err error
	data.ID, err = strconv.Atoi(dataIn.ID)
	if err != nil {
		return errors.Wrap(err, "ID is not number")
	}
	data.Event = "TranscriptionFinished"
	if dataIn.Error != nil {
		data.Status = statusFailed
		data.Error = &trError{Code: dataIn.Error.Code, DebugMessage: dataIn.Error.Error}
	} else {
		data.Status = statusDone
		data.Transcription = &transcription{Text: dataIn.Transcription.Text,
			Latice: dataIn.Transcription.LatticeData, WebVTT: dataIn.Transcription.WebVTT}
	}

	bytesData, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "Can't marshal data")
	}
	resp, err := sp.httpclient.Post(urlStr, "application/json", bytes.NewBuffer(bytesData))
	if err != nil {
		cmdapp.Log.Tracef("JSON: %s", string(bytesData))
		return errors.Wrap(err, "Can't send data to file server")
	}
	err = utils.ValidateResponse(resp)
	if err != nil {
		bodyBytes, err1 := ioutil.ReadAll(resp.Body)
		if err1 != nil {
			bodyBytes = []byte{}
		}
		cmdapp.Log.Tracef("JSON: %s", string(bytesData))
		cmdapp.Log.Debugf("Response: code%d\n%s", resp.StatusCode, string(bodyBytes))
		return errors.Wrap(err, "Can't save transcription")
	}
	return nil
}
