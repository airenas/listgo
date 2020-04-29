package punctuation

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"bitbucket.org/airenas/listgo/internal/app/punctuation/api"
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
	pegomock.When(punctuatorMock.Process(pegomock.AnyStringSlice())).ThenReturn(&api.PResult{}, nil)
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
	pegomock.When(punctuatorMock.Process(pegomock.AnyStringSlice())).ThenReturn(&api.PResult{PunctuatedText: "Olia, olia.",
		Punctuated: []string{"Olia,", "olia."}}, nil)
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
	output := getOutput(resp.Body)
	assert.Equal(t, []string{"olia", "olia"}, output.Original)
	assert.Equal(t, []string{"Olia,", "olia."}, output.Punctuated)
	assert.Equal(t, "Olia, olia.", output.PunctuatedText)
}

func TestOutput_NoDebugData(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/punctuation", newInput("olia olia"))
	resp := httptest.NewRecorder()
	pegomock.When(punctuatorMock.Process(pegomock.AnyStringSlice())).ThenReturn(&api.PResult{PunctuatedText: "Olia, olia.",
		WordIDs: []int32{1, 2}, PunctIDs: []int32{0, 1}}, nil)
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
	output := getOutput(resp.Body)
	assert.Equal(t, []string{"olia", "olia"}, output.Original)
	assert.Equal(t, "Olia, olia.", output.PunctuatedText)
	assert.Empty(t, output.WordIDs)
	assert.Empty(t, output.PunctIDs)
}

func TestOutput_DebugData(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/punctuation?debug=1", newInput("olia olia"))
	resp := httptest.NewRecorder()
	pegomock.When(punctuatorMock.Process(pegomock.AnyStringSlice())).ThenReturn(&api.PResult{PunctuatedText: "Olia, olia.",
		WordIDs: []int32{1, 2}, PunctIDs: []int32{0, 1}}, nil)
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
	output := getOutput(resp.Body)
	assert.Equal(t, []string{"olia", "olia"}, output.Original)
	assert.Equal(t, "Olia, olia.", output.PunctuatedText)
	assert.Equal(t, []int32{1, 2}, output.WordIDs)
	assert.Equal(t, []int32{0, 1}, output.PunctIDs)
}

func TestPunctuatorFails(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/punctuation", newInput("olia olia"))
	resp := httptest.NewRecorder()
	pegomock.When(punctuatorMock.Process(pegomock.AnyStringSlice())).ThenReturn(nil, errors.New("error"))
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)
}

func TestPunctuatorArrayWrongInput(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/punctuationArray", newInput("olia olia"))
	resp := httptest.NewRecorder()
	pegomock.When(punctuatorMock.Process(pegomock.AnyStringSlice())).ThenReturn(nil, errors.New("error"))
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, 400, resp.Code)
}

func TestPunctuatorArrayFails(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/punctuationArray", newArrInput("olia olia"))
	resp := httptest.NewRecorder()
	pegomock.When(punctuatorMock.Process(pegomock.AnyStringSlice())).ThenReturn(nil, errors.New("error"))
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

func newArrInput(text string) *bytes.Buffer {
	result := new(bytes.Buffer)
	json.NewEncoder(result).Encode(InputArray{Words: strings.Split(text, " ")})
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
