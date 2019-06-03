package fs

import (
	"bytes"
	"encoding/json"
	"net/url"
	"path"

	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"

	"github.com/hashicorp/go-retryablehttp"
)

//Client comunicates with file server
type Client struct {
	httpclient *retryablehttp.Client
	url        string
}

//NewClient creates a fs client
func NewClient() (*Client, error) {
	res := Client{}
	urlStr := cmdapp.Config.GetString("fs.url")
	if urlStr == "" {
		return nil, errors.New("No fs.url provided")
	}
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't parse url "+urlStr)
	}
	res.url = url.String()
	res.httpclient = retryablehttp.NewClient()
	res.httpclient.RetryMax = 3

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
	u, _ := url.Parse(sp.url)
	u.Path = path.Join(u.Path, "AudioGetRequest", kafkaID)
	urlStr := u.String()
	cmdapp.Log.Infof("Get audio: %s", urlStr)
	resp, err := sp.httpclient.Get(urlStr)
	if err != nil {
		return nil, err
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

//SaveResult saves result to fs
func (sp *Client) SaveResult(data *kafkaapi.DBResultEntry) error {
	bytesData, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "Can't marshal data")
	}
	urlStr := path.Join(sp.url, "TranscriptionPostRequest")
	cmdapp.Log.Infof("Post result audio: %s", urlStr)
	_, err = sp.httpclient.Post(urlStr, "application/json", bytes.NewBuffer(bytesData))
	if err != nil {
		return errors.Wrap(err, "Can't send data to file server")
	}
	return nil
}
