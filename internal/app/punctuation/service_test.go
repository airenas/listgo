package punctuation

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"github.com/heptiolabs/healthcheck"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var punctuatorMock *mocks.MockPunctuator

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	punctuatorMock = mocks.NewMockPunctuator()
}

func TestWrongPath(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("GET", "/invalid", nil)
	resp := httptest.NewRecorder()
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, 404, resp.Code)
}

func TestWrongMethod(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("GET", "/punctuation", nil)
	resp := httptest.NewRecorder()
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, 405, resp.Code)
}

func TestProcess(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/punctuation", newInput("olia"))
	resp := httptest.NewRecorder()
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
}

func TestNoData(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/punctuation", nil)
	resp := httptest.NewRecorder()
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)
}

func TestEmptyText(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/punctuation", newInput(""))
	resp := httptest.NewRecorder()
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)
}

func TestOutput(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/punctuation", newInput("olia olia"))
	resp := httptest.NewRecorder()
	pegomock.When(punctuatorMock.Process(pegomock.AnyString())).ThenReturn("Olia, olia.", nil)
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
	output := getOutput(resp.Body)
	assert.Equal(t, "olia olia", output.Original)
	assert.Equal(t, "Olia, olia.", output.Punctuated)
}

func TestPunctuatorFails(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/punctuation", newInput("olia olia"))
	resp := httptest.NewRecorder()
	pegomock.When(punctuatorMock.Process(pegomock.AnyString())).ThenReturn("", errors.New("error"))
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)
}

func newData() *ServiceData {
	data := ServiceData{}
	data.health = healthcheck.NewHandler()
	data.punctuator = punctuatorMock
	return &data
}

func getOutput(r io.Reader) *Output {
	decoder := json.NewDecoder(r)
	res := Output{}
	decoder.Decode(&res)
	return &res
}

func newInput(text string) *bytes.Buffer {
	result := new(bytes.Buffer)
	json.NewEncoder(result).Encode(Input{Text: text})
	return result
}

func TestLive(t *testing.T) {
	testCode(t, newData(), "/live", 200)
}

func TestLive503(t *testing.T) {
	data := newData()
	data.health.AddLivenessCheck("test", func() error { return errors.New("test") })
	testCode(t, data, "/live", 503)
}

func testCode(t *testing.T, data *ServiceData, path string, code int) {
	initTest(t)
	req := httptest.NewRequest("GET", path, nil)
	resp := httptest.NewRecorder()
	NewRouter(data).ServeHTTP(resp, req)
	assert.Equal(t, code, resp.Code)
}

func TestReady(t *testing.T) {
	testCode(t, newData(), "/ready", 200)
}
