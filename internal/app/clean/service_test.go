package clean

import (
	"errors"
	"net/http/httptest"
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"github.com/heptiolabs/healthcheck"
	"github.com/petergtz/pegomock"
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
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, 404, resp.Code)
}

func TestWrongMethod(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("POST", "/id", nil)
	resp := httptest.NewRecorder()
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, 405, resp.Code)
}

func TestDelete(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("DELETE", "/id", nil)
	resp := httptest.NewRecorder()
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code)
}

func TestNoData(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("DELETE", "/", nil)
	resp := httptest.NewRecorder()
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, 404, resp.Code)
}

func TestCleanerFails(t *testing.T) {
	initTest(t)
	req := httptest.NewRequest("DELETE", "/id", nil)
	resp := httptest.NewRecorder()
	pegomock.When(cleanerMock.Clean(pegomock.AnyString())).ThenReturn(errors.New("error"))
	NewRouter(newData()).ServeHTTP(resp, req)
	assert.Equal(t, 500, resp.Code)
}

func newData() *ServiceData {
	data := ServiceData{}
	data.health = healthcheck.NewHandler()
	data.cleaner = cleanerMock
	return &data
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
