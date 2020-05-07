package clean

import (
	"errors"
	"net/http/httptest"
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"github.com/heptiolabs/healthcheck"
	"github.com/petergtz/pegomock"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

var cleanerMock *mocks.MockCleaner

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	cleanerMock = mocks.NewMockCleaner()
	pegomock.When(cleanerMock.Clean(pegomock.AnyString())).ThenReturn(nil)
}

func TestWrongPath(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("GET", "/olia/id", nil)
	resp := httptest.NewRecorder()
	NewRouter(newTestData()).ServeHTTP(resp, req)
	assert.Equal(t, 404, resp.Code)
}

func TestWrongMethod(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/id", nil)
	resp := httptest.NewRecorder()
	NewRouter(newTestData()).ServeHTTP(resp, req)
	assert.Equal(t, 405, resp.Code)
}

func TestDelete(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("DELETE", "/id", nil)
	resp := httptest.NewRecorder()
	NewRouter(newTestData()).ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
}

func TestNoData(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("DELETE", "/", nil)
	resp := httptest.NewRecorder()
	NewRouter(newTestData()).ServeHTTP(resp, req)
	assert.Equal(t, 404, resp.Code)
}

func TestCleanerFails(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("DELETE", "/id", nil)
	resp := httptest.NewRecorder()
	pegomock.When(cleanerMock.Clean(pegomock.AnyString())).ThenReturn(errors.New("error"))
	NewRouter(newTestData()).ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)
}

func newTestData() *ServiceData {
	data := &ServiceData{}
	data.health = healthcheck.NewHandler()
	data.cleaner = cleanerMock
	initMetrics(data)
	return data
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

func TestReady(t *testing.T) {
	testCode(t, newTestData(), "/ready", 200)
}

func TestMetrics(t *testing.T) {
	testCode(t, newTestData(), "/metrics", 200)
}

func TestAddMetric(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("DELETE", "/id", nil)
	resp := httptest.NewRecorder()
	data := newTestData()
	NewRouter(data).ServeHTTP(resp, req)
	assert.Equal(t, 1, testutil.CollectAndCount(data.metrics.responseDur))

}
