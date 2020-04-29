package punctuation

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
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
	testFail(t, httptest.NewRequest("GET", "/invalid", nil), 404)
}

func TestWrongMethod(t *testing.T) {
	initTest(t)
	testFail(t, httptest.NewRequest("GET", "/punctuation", nil), 405)
	testFail(t, httptest.NewRequest("GET", "/punctuationArray", nil), 405)
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
	testFail(t, httptest.NewRequest("POST", "/punctuation", nil), 400)
	testFail(t, httptest.NewRequest("POST", "/punctuationArray", nil), 400)
}

func TestEmptyText(t *testing.T) {
	initTest(t)
	testFail(t, httptest.NewRequest("POST", "/punctuation", newInput("")), 400)
}

func TestOutput(t *testing.T) {
	initTest(t)
	testOutput(t, httptest.NewRequest("POST", "/punctuation", newInput("olia olia")))
}

func TestOutputArray(t *testing.T) {
	initTest(t)
	testOutput(t, httptest.NewRequest("POST", "/punctuationArray", newArrInput("olia olia")))
}

func testOutput(t *testing.T, req *http.Request) {
	initTest(t)
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
	testNoDebugData(t, httptest.NewRequest("POST", "/punctuation", newInput("olia olia")))
}

func TestArray_NoDebugData(t *testing.T) {
	initTest(t)
	testNoDebugData(t, httptest.NewRequest("POST", "/punctuationArray", newArrInput("olia olia")))
}

func testNoDebugData(t *testing.T, req *http.Request) {
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
	testDebugData(t, httptest.NewRequest("POST", "/punctuation?debug=1", newInput("olia olia")))
}

func TestArray_DebugData(t *testing.T) {
	initTest(t)
	testDebugData(t, httptest.NewRequest("POST", "/punctuationArray?debug=1", newArrInput("olia olia")))
}

func testDebugData(t *testing.T, req *http.Request) {
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
	pegomock.When(punctuatorMock.Process(pegomock.AnyStringSlice())).ThenReturn(nil, errors.New("error"))
	testFail(t, httptest.NewRequest("POST", "/punctuation", newInput("olia olia")), 500)
}

func TestPunctuatorWrongInput(t *testing.T) {
	initTest(t)
	testFail(t, httptest.NewRequest("POST", "/punctuation", newArrInput("olia olia")), 400)
}

func TestPunctuatorArrayWrongInput(t *testing.T) {
	initTest(t)
	testFail(t, httptest.NewRequest("POST", "/punctuationArray", newInput("olia olia")), 400)
	testFail(t, httptest.NewRequest("POST", "/punctuationArray", newArrInput("")), 400)
}

func TestPunctuatorArrayFails(t *testing.T) {
	initTest(t)
	pegomock.When(punctuatorMock.Process(pegomock.AnyStringSlice())).ThenReturn(nil, errors.New("error"))
	testFail(t, httptest.NewRequest("POST", "/punctuationArray", newArrInput("olia olia")), 500)
}

func testFail(t *testing.T, req *http.Request, expectedCode int) {
	resp := httptest.NewRecorder()
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, expectedCode, resp.Code)
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
	arr := []string{}
	if len(text) > 0 {
		arr = strings.Split(text, " ")
	}
	json.NewEncoder(result).Encode(InputArray{Words: arr})
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
