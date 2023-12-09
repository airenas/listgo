package status

import (
	"errors"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/heptiolabs/healthcheck"
	"github.com/stretchr/testify/assert"

	"github.com/airenas/listgo/internal/app/status/api"
)

func TestWrongPath(t *testing.T) {
	req := httptest.NewRequest("GET", "/invalid", nil)
	resp := httptest.NewRecorder()
	data := newTestData()
	NewRouter(data).ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 404)
}

func TestNoID(t *testing.T) {
	test400(t, "/status")
	test400(t, "/status/")
}

func test400(t *testing.T, path string) {
	req := httptest.NewRequest("GET", path, nil)
	resp := httptest.NewRecorder()
	data := newTestData()
	NewRouter(data).ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 400)
}

func Test_ReturnsResult(t *testing.T) {

	req := httptest.NewRequest("GET", "/status/x", nil)
	resp := httptest.NewRecorder()
	data := newTestData()
	data.StatusProvider = testStatusProvider{}
	NewRouter(data).ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 200)
	assert.True(t, strings.HasPrefix(resp.Body.String(), `{"id":"`))
}

func Test_ProviderFails(t *testing.T) {
	req := httptest.NewRequest("GET", "/status/x", nil)
	resp := httptest.NewRecorder()
	data := newTestData()
	data.StatusProvider = testStatusFunc(
		func(ID string) (*api.TranscriptionResult, error) {
			return nil, errors.New("Can not get")
		})
	NewRouter(data).ServeHTTP(resp, req)
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
	testCode(t, newTestData(), "/live", 200)
}

func TestLive503(t *testing.T) {
	data := newTestData()
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

func newTestData() *ServiceData {
	data := &ServiceData{}
	initMetrics(data)
	data.health = healthcheck.NewHandler()
	return data
}

func TestReady(t *testing.T) {
	testCode(t, newTestData(), "/ready", 200)
}

func TestMetrics(t *testing.T) {
	testCode(t, newTestData(), "/metrics", 200)
}
