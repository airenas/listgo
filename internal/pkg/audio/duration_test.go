package audio

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInit_FailOnWronURL(t *testing.T) {
	_, err := NewDurationClient("http://")
	assert.NotNil(t, err)
	_, err = NewDurationClient("")
	assert.NotNil(t, err)
}

func TestInit(t *testing.T) {
	d, err := NewDurationClient("http://localhost:8000")
	assert.Nil(t, err)
	assert.NotNil(t, d)
}

func initTestServer(t *testing.T, rCode int, body string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(rCode)
		rw.Write([]byte(body))
	}))
	return server
}

func TestClient(t *testing.T) {
	var resp durationResponse
	resp.Duration = 10
	rb, _ := json.Marshal(resp)
	server := initTestServer(t, 200, string(rb))
	defer server.Close()
	d, _ := NewDurationClient(server.URL)
	d.httpclient = server.Client()

	r, err := d.Get("1.wav", strings.NewReader("olia"))

	assert.Nil(t, err)
	assert.Equal(t, time.Second*10, r)
}

func TestClient_PassFile(t *testing.T) {
	var resp durationResponse
	resp.Duration = 10
	rb, _ := json.Marshal(resp)
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		req.ParseMultipartForm(32 << 20)
		file, handler, _ := req.FormFile("file")
		defer file.Close()
		assert.Equal(t, "1.wav", handler.Filename)
		buf := new(strings.Builder)
		io.Copy(buf, file)
		assert.Equal(t, "olia", buf.String())
		rw.WriteHeader(200)
		rw.Write(rb)
	}))
	defer server.Close()
	d, _ := NewDurationClient(server.URL)
	d.httpclient = server.Client()

	_, err := d.Get("1.wav", strings.NewReader("olia"))

	assert.Nil(t, err)
}

func TestClient_Fail(t *testing.T) {
	var resp durationResponse
	resp.Duration = 10
	rb, _ := json.Marshal(resp)
	server := initTestServer(t, 400, string(rb))
	defer server.Close()
	d, _ := NewDurationClient(server.URL)
	d.httpclient = server.Client()

	_, err := d.Get("1.wav", strings.NewReader("olia"))

	assert.NotNil(t, err)
}
