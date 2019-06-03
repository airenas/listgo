package transcriberapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/url"
	"path"
	"strings"

	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
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
	res.uploadURL, err = getURL("transcriber.url.upload")
	if err != nil {
		return nil, err
	}
	res.statusURL, err = getURL("transcriber.url.status")
	if err != nil {
		return nil, err
	}
	res.resultURL, err = getURL("transcriber.url.result")
	if err != nil {
		return nil, err
	}
	res.httpclient = retryablehttp.NewClient()
	res.httpclient.RetryMax = 3

	return &res, nil
}

// Upload(audio *kafkaapi.UploadData) (string, error)
// 	GetStatus(ID string) (*kafkaapi.Status, error)
// 	GetResult(ID string) (*kafkaapi.Result, error)

//GetStatus get status from the server
func (sp *Client) GetStatus(ID string) (*kafkaapi.Status, error) {
	urlStr := path.Join(sp.statusURL, "status", ID)
	cmdapp.Log.Infof("Get status: %s", urlStr)
	resp, err := sp.httpclient.Get(urlStr)
	if err != nil {
		return nil, err
	}
	var result kafkaapi.Status
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrap(err, "Can't decode response")
	}

	return &result, nil
}

//GetResult gets result file from transcrinber
func (sp *Client) GetResult(ID string) (*kafkaapi.Result, error) {
	// bytesData, err := json.Marshal(data)
	// if err != nil {
	// 	return errors.Wrap(err, "Can't marshal data")
	// }
	// urlStr := path.Join(sp.url.Path, "TranscriptionPostRequest")
	// cmdapp.Log.Infof("Post result audio: %s", urlStr)
	// _, err = sp.httpclient.Post(urlStr, "application/json", bytes.NewBuffer(bytesData))
	// if err != nil {
	// 	return errors.Wrap(err, "Can't send data to file server")
	// }
	return nil, nil
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
	if err != nil {
		return "", err
	}
	var respData uploadResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return "", errors.Wrap(err, "Can't decode response")
	}
	return respData.ID, nil
}

func getURL(name string) (string, error) {
	urlStr := cmdapp.Config.GetString(name)
	if urlStr == "" {
		return "", errors.New("No " + name + " setting provided")
	}
	url, err := url.Parse(urlStr)
	if err != nil {
		return "", errors.Wrap(err, "Can't parse url "+urlStr)
	}
	return url.String(), nil
}
