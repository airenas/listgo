package transcriberapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"strings"

	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"bitbucket.org/airenas/listgo/internal/app/status/api"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"
	"github.com/pkg/errors"

	"github.com/hashicorp/go-retryablehttp"
)

//Client comunicates with transcriber service
type Client struct {
	httpclient *retryablehttp.Client
	uploadURL  string
	statusURL  string
	resultURL  string
}

//NewClient creates a transcriber client
func NewClient() (*Client, error) {
	res := Client{}
	var err error
	res.uploadURL, err = utils.GetURLFromConfig("transcriber.url.upload")
	if err != nil {
		return nil, err
	}
	res.statusURL, err = utils.GetURLFromConfig("transcriber.url.status")
	if err != nil {
		return nil, err
	}
	res.resultURL, err = utils.GetURLFromConfig("transcriber.url.result")
	if err != nil {
		return nil, err
	}
	res.httpclient = retryablehttp.NewClient()
	res.httpclient.RetryMax = 3

	return &res, nil
}

//GetStatus get status from the server
func (sp *Client) GetStatus(ID string) (*kafkaapi.Status, error) {
	urlStr := utils.URLJoin(sp.statusURL, ID)
	cmdapp.Log.Infof("Get status: %s", urlStr)
	resp, err := sp.httpclient.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = utils.ValidateResponse(resp)
	if err != nil {
		return nil, errors.Wrap(err, "Can't get status")
	}

	var result api.TranscriptionResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrap(err, "Can't decode response")
	}

	var res kafkaapi.Status
	res.ID = result.ID
	res.ErrorCode = result.ErrorCode
	res.Error = result.Error
	res.Text = result.RecognizedText
	res.Completed = result.Status == "COMPLETED"

	return &res, nil
}

//GetResult gets result file from transcrinber
func (sp *Client) GetResult(ID string) (*kafkaapi.Result, error) {
	urlStr := utils.URLJoin(sp.resultURL, "result", ID, "result.txt")
	resp, err := sp.httpclient.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = utils.ValidateResponse(resp)
	if err != nil {
		return nil, errors.Wrap(err, "Can't get status")
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		return nil, errors.Errorf("Can't get result. Code: %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Can't read response")
	}
	var res kafkaapi.Result
	res.ID = ID
	res.FileData = base64.StdEncoding.EncodeToString(body)
	return &res, nil
}

type uploadResponse struct {
	ID string `json:"id"`
}

//Upload uploads audio to transcriber service
func (sp *Client) Upload(audio *kafkaapi.UploadData) (string, error) {
	cmdapp.Log.Infof("Sending audio to: %s", sp.uploadURL)

	dataDecoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(audio.AudioData))
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", audio.FileName)
	if err != nil {
		return "", errors.Wrap(err, "Can't add file to request")
	}
	_, err = io.Copy(part, dataDecoder)
	if err != nil {
		return "", errors.Wrap(err, "Can't add file to request")
	}
	writer.WriteField("externalID", audio.ExternalID)
	writer.Close()
	req, err := retryablehttp.NewRequest("POST", sp.uploadURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := sp.httpclient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	var respData uploadResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return "", errors.Wrap(err, "Can't decode response")
	}
	if respData.ID == "" {
		return "", errors.Wrap(err, "Can't get ID from response")
	}
	return respData.ID, nil
}
