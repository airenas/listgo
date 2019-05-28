package result

import (
	"errors"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"github.com/gorilla/mux"
	"github.com/petergtz/pegomock"
)

var audioFileLoaderMock *mocks.MockFileLoader
var resultFileLoaderMock *mocks.MockFileLoader
var fileMock *mocks.MockFile
var fileNameProviderMock *mocks.MockFileNameProvider

func initTest() {
	audioFileLoaderMock = mocks.NewMockFileLoader()
	resultFileLoaderMock = mocks.NewMockFileLoader()
	fileMock = mocks.NewMockFile()
	fileNameProviderMock = mocks.NewMockFileNameProvider()
}

func TestWrongPath(t *testing.T) {
	req := httptest.NewRequest("GET", "/invalid", nil)
	resp := httptest.NewRecorder()
	newRouter().ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 404)
}

func TestNoID(t *testing.T) {
	req := httptest.NewRequest("GET", "/audio/", nil)
	resp := httptest.NewRecorder()
	newRouter().ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 404)
}

func TestGET(t *testing.T) {
	initTest()
	req := httptest.NewRequest("GET", "/audio/id", nil)
	resp := httptest.NewRecorder()
	pegomock.When(fileNameProviderMock.Get(pegomock.AnyString())).ThenReturn("olia", nil)
	pegomock.When(audioFileLoaderMock.Load(pegomock.AnyString())).ThenReturn(fileMock, nil)
	pegomock.When(fileMock.Stat()).ThenReturn(mockedFileInfo{}, nil)
	pegomock.When(fileMock.Seek(pegomock.AnyInt64(), pegomock.AnyInt())).ThenReturn(int64(2), nil)
	pegomock.When(fileMock.Read(anyByteArray())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			return []pegomock.ReturnValue{2, nil}
		})
	newRouter().ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 200)
	assert.NotEmpty(t, resp.Body.String())
}

func newRouter() *mux.Router {
	return NewRouter(&ServiceData{audioFileLoader: audioFileLoaderMock,
		resultFileLoader: resultFileLoaderMock,
		fileNameProvider: fileNameProviderMock})
}

func Test_FileNameProviderFails(t *testing.T) {
	initTest()
	req := httptest.NewRequest("GET", "/audio/id", nil)
	resp := httptest.NewRecorder()
	pegomock.When(fileNameProviderMock.Get(pegomock.AnyString())).ThenReturn("", errors.New("Can not get"))

	newRouter().ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 404)
}

func Test_FileLoaderFails(t *testing.T) {
	initTest()
	req := httptest.NewRequest("GET", "/audio/id", nil)
	resp := httptest.NewRecorder()
	pegomock.When(fileNameProviderMock.Get(pegomock.AnyString())).ThenReturn("olia", nil)
	pegomock.When(audioFileLoaderMock.Load(pegomock.AnyString())).ThenReturn(nil, errors.New("Can not get"))

	newRouter().ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 404)
}

func Test_FileStatFails(t *testing.T) {
	initTest()
	req := httptest.NewRequest("GET", "/audio/id", nil)
	resp := httptest.NewRecorder()
	pegomock.When(fileNameProviderMock.Get(pegomock.AnyString())).ThenReturn("olia", nil)
	pegomock.When(audioFileLoaderMock.Load(pegomock.AnyString())).ThenReturn(fileMock, nil)
	pegomock.When(fileMock.Stat()).ThenReturn(mockedFileInfo{}, errors.New("Can not get"))

	newRouter().ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 404)
}

func TestResultNoID(t *testing.T) {
	req := httptest.NewRequest("GET", "/result/", nil)
	resp := httptest.NewRecorder()

	newRouter().ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 404)
}

func TestResultNoFile(t *testing.T) {
	req := httptest.NewRequest("GET", "/result/id/", nil)
	resp := httptest.NewRecorder()
	newRouter().ServeHTTP(resp, req)

	assert.Equal(t, resp.Code, 404)
}

func TestResultWrongFile(t *testing.T) {
	req := httptest.NewRequest("GET", "/result/id/..file", nil)
	resp := httptest.NewRecorder()
	newRouter().ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 400)
}

func TestResultGET(t *testing.T) {
	initTest()
	req := httptest.NewRequest("GET", "/result/id/file", nil)
	resp := httptest.NewRecorder()
	pegomock.When(resultFileLoaderMock.Load(pegomock.AnyString())).ThenReturn(fileMock, nil)
	pegomock.When(fileMock.Stat()).ThenReturn(mockedFileInfo{}, nil)
	pegomock.When(fileMock.Seek(pegomock.AnyInt64(), pegomock.AnyInt())).ThenReturn(int64(2), nil)
	pegomock.When(fileMock.Read(anyByteArray())).Then(
		func(params []pegomock.Param) pegomock.ReturnValues {
			return []pegomock.ReturnValue{2, nil}
		})
	newRouter().ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 200)
	assert.NotEmpty(t, resp.Body.String())
}

func TestResult_FileLoaderFails(t *testing.T) {
	initTest()
	req := httptest.NewRequest("GET", "/result/id/file", nil)
	resp := httptest.NewRecorder()
	pegomock.When(resultFileLoaderMock.Load(pegomock.AnyString())).ThenReturn(nil, errors.New("Can not get"))

	newRouter().ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 404)
}

func TestResult_FileStatFails(t *testing.T) {
	initTest()
	req := httptest.NewRequest("GET", "/result/id/file", nil)
	resp := httptest.NewRecorder()
	pegomock.When(resultFileLoaderMock.Load(pegomock.AnyString())).ThenReturn(fileMock, nil)
	pegomock.When(fileMock.Stat()).ThenReturn(mockedFileInfo{}, errors.New("Can not get"))

	newRouter().ServeHTTP(resp, req)
	assert.Equal(t, resp.Code, 404)
}

type mockedFileInfo struct {
	os.FileInfo
}

func (mockedFileInfo) Name() string {
	return "olia.wav"
}

func (mockedFileInfo) ModTime() time.Time {
	return time.Now()
}

func anyByteArray() []byte {
	pegomock.RegisterMatcher(pegomock.NewAnyMatcher(reflect.TypeOf([]byte{})))
	return []byte{}
}
