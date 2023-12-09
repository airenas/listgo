package transcriberapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"

	"github.com/airenas/listgo/internal/pkg/status"

	"github.com/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"github.com/airenas/listgo/internal/app/status/api"
	uparams "github.com/airenas/listgo/internal/app/upload/api"
	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/airenas/listgo/internal/pkg/utils"
	"github.com/pkg/errors"
)

// Client comunicates with transcriber service
type Client struct {
	httpclient *http.Client
	uploadURL  string
	statusURL  string
	resultURL  string
	cleanURL   string
}

// NewClient creates a transcriber client
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
	res.cleanURL, err = utils.GetURLFromConfig("transcriber.url.clean")
	if err != nil {
		return nil, err
	}
	res.httpclient = &http.Client{}

	return &res, nil
}

// GetStatus get status from the server
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
	res.Completed = status.From(result.Status) == status.Completed

	return &res, nil
}

// GetResult gets result file from transcrinber
func (sp *Client) GetResult(ID string) (*kafkaapi.Result, error) {
	var err error
	var lock sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)
	var res kafkaapi.Result
	res.ID = ID
	go func() {
		defer wg.Done()
		url := utils.URLJoin(sp.resultURL, "result", ID, "lat.restored.txt")
		b, errF := getStringResult(sp.httpclient, url)
		if errF != nil {
			lock.Lock()
			defer lock.Unlock()
			err = errF
		}
		res.LatticeData = base64.StdEncoding.EncodeToString(b)
	}()
	go func() {
		defer wg.Done()
		url := utils.URLJoin(sp.resultURL, "result", ID, "webvtt.txt")
		b, errF := getStringResult(sp.httpclient, url)
		if errF != nil {
			lock.Lock()
			defer lock.Unlock()
			err = errF
		}
		res.WebVTTData = string(b)
	}()
	wg.Wait()
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func getStringResult(httpclient *http.Client, urlStr string) ([]byte, error) {
	cmdapp.Log.Debugf("Calling %s", urlStr)
	resp, err := httpclient.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = utils.ValidateResponse(resp)
	if err != nil {
		return nil, errors.Wrap(err, "Can't get result")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Can't read response")
	}
	return body, nil
}

type uploadResponse struct {
	ID string `json:"id"`
}

// Upload uploads audio to transcriber service
func (sp *Client) Upload(audio *kafkaapi.UploadData) (string, error) {
	dataDecoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(audio.AudioData))
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(uparams.PrmFile, audio.FileName)
	if err != nil {
		return "", errors.Wrap(err, "Can't add file to request")
	}
	_, err = io.Copy(part, dataDecoder)
	if err != nil {
		return "", errors.Wrap(err, "Can't add file to request")
	}
	writer.WriteField(uparams.PrmExternalID, audio.ExternalID)
	if audio.NumberOfSpeakers != "" {
		writer.WriteField(uparams.PrmNumberOfSpeakers, audio.NumberOfSpeakers)
	}
	rec := getRecognizer(audio)
	writer.WriteField(uparams.PrmRecognizer, rec)
	writer.Close()
	req, err := http.NewRequest("POST", sp.uploadURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	cmdapp.Log.Debugf("Sending audio to: %s for model %s", sp.uploadURL, rec)
	resp, err := sp.httpclient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	err = utils.ValidateResponse(resp)
	if err != nil {
		return "", errors.Wrap(err, "Can't upload")
	}
	var respData uploadResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return "", errors.Wrap(err, "Can't decode response")
	}
	if respData.ID == "" {
		return "", errors.New("Can't get ID from response")
	}
	return respData.ID, nil
}

// Delete removes all transcription data related with ID
func (sp *Client) Delete(ID string) error {
	urlStr := utils.URLJoin(sp.cleanURL, ID)
	cmdapp.Log.Infof("Invoke clean data request to: %s", urlStr)

	req, err := http.NewRequest("DELETE", urlStr, nil)
	if err != nil {
		return err
	}
	resp, err := sp.httpclient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = utils.ValidateResponse(resp)
	if err != nil {
		return errors.Wrap(err, "Can't delete information")
	}
	return nil
}

func getRecognizer(audio *kafkaapi.UploadData) string {
	rc := strings.TrimSpace(audio.RecordQuality)
	if rc == "" {
		return strings.TrimSpace(audio.JobType)
	}
	return strings.TrimSpace(audio.JobType) + "_" + rc
}
