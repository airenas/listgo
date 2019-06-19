package status

import (
	"errors"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/heptiolabs/healthcheck"
	"github.com/stretchr/testify/assert"

	"bitbucket.org/airenas/listgo/internal/app/status/api"
)

func TestWrongPath(t *testing.T) {
	req := httptest.NewRequest("GET", "/invalid", nil)
	resp := httptest.NewRecorder()
	NewRouter(&ServiceData{}).ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 404)
}

func TestNoID(t *testing.T) {
	test400(t, "/status")
	test400(t, "/status/")
}

func test400(t *testing.T, path string) {
	req := httptest.NewRequest("GET", path, nil)
	resp := httptest.NewRecorder()
	NewRouter(&ServiceData{}).ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 400)
}

func Test_ReturnsResult(t *testing.T) {

	req := httptest.NewRequest("GET", "/status/x", nil)
	resp := httptest.NewRecorder()

	NewRouter(&ServiceData{StatusProvider: testStatusProvider{}}).ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 200)
	assert.True(t, strings.HasPrefix(resp.Body.String(), `{"id":"`))
}

func Test_ProviderFails(t *testing.T) {
	req := httptest.NewRequest("GET", "/status/x", nil)
	resp := httptest.NewRecorder()

	NewRouter(&ServiceData{StatusProvider: testStatusFunc(
		func(ID string) (*api.TranscriptionResult, error) {
			return nil, errors.New("Can not get")
		})}).ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 400)
}

type testStatusFunc func(ID string) (*api.TranscriptionResult, error)

func (f testStatusFunc) Get(ID string) (*api.TranscriptionResult, error) {
	return f(ID)
}

type testStatusProvider struct{}

func (p testStatusProvider) Get(ID string) (*api.TranscriptionResult, error) {
	log.Printf("Get status %s \n", ID)
	return &api.TranscriptionResult{}, nil
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

func newData() *ServiceData {
	data := ServiceData{}
	data.health = healthcheck.NewHandler()
	return &data
}

func TestReady(t *testing.T) {
	testCode(t, newData(), "/ready", 200)
}
