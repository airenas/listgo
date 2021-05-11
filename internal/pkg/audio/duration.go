package audio

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
)

//Duration comunicates with duration service
type Duration struct {
	httpclient *http.Client
	url        string
}

//NewDurationClient creates a transcriber client
func NewDurationClient(urlStr string) (*Duration, error) {
	res := Duration{}
	var err error
	urlRes, err := url.Parse(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Can't parse url "+urlStr)
	}
	if urlRes.Host == "" {
		return nil, errors.New("Can't parse url " + urlStr)
	}
	res.url = urlRes.String()
	res.httpclient = &http.Client{}
	return &res, nil
}

//Get return duration by calling the service
func (dc *Duration) Get(name string, file io.Reader) (time.Duration, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		return 0, errors.Wrap(err, "Can't add file to request")
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return 0, errors.Wrap(err, "Can't add file to request")
	}
	writer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", dc.url, body)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	cmdapp.Log.Debugf("Sending audio to: %s", dc.url)
	resp, err := dc.httpclient.Do(req)
	if err != nil {
		return 0, err
	}
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return 0, errors.New("Can't get duration")
	}
	var respData durationResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return 0, errors.Wrap(err, "Can't decode response")
	}
	return time.Millisecond * time.Duration(int32(respData.Duration*1000)), nil
}

type durationResponse struct {
	Duration float64 `json:"duration"`
}
