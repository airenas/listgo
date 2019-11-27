package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"bitbucket.org/airenas/listgo/internal/app/upload/api"
	"bitbucket.org/airenas/listgo/internal/pkg/recognizer"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var recInfoLoaderMock *mocks.MockRecInfoLoader

func initTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	recInfoLoaderMock = mocks.NewMockRecInfoLoader()
	pegomock.When(recInfoLoaderMock.Get(pegomock.AnyString())).ThenReturn(&recognizer.Info{}, nil)
}

func createTempFile(t *testing.T) *os.File {
	f, err := ioutil.TempFile("", "test")
	assert.Nil(t, err)
	return f
}

func load(t *testing.T) (*FileRecognizerMap, *os.File) {
	initTest(t)
	f := createTempFile(t)
	fmt.Fprint(f, "rec: recID")
	r, err := newFileRecognizerMap(f.Name())
	assert.Nil(t, err)
	r.rCache.fileLoader = recInfoLoaderMock
	return r, f
}

func Test_Load(t *testing.T) {
	r, f := load(t)
	defer os.Remove(f.Name())
	assert.NotNil(t, r)
}

func Test_Get(t *testing.T) {
	r, f := load(t)
	defer os.Remove(f.Name())
	assert.NotNil(t, r)
	v, _ := r.Get("rec")
	assert.Equal(t, "recID", v)
}

func Test_GetFails(t *testing.T) {
	r, f := load(t)
	defer os.Remove(f.Name())
	assert.NotNil(t, r)
	v, err := r.Get("rec1")
	assert.Equal(t, "", v)
	assert.Equal(t, api.ErrRecognizerNotFound, err)
	v, err = r.Get("")
	assert.Equal(t, "", v)
	assert.Equal(t, api.ErrRecognizerNotFound, err)
}

func Test_Reload(t *testing.T) {
	f := createTempFile(t)
	defer os.Remove(f.Name())

	fmt.Fprint(f, "rec: recID\n")
	recMap, err := newFileRecognizerMap(f.Name())
	assert.Nil(t, err)
	assert.NotNil(t, recMap)
	v, err := recMap.Get("rec1")
	assert.Equal(t, "", v)

	fmt.Fprint(f, "rec1: recID1")
	time.Sleep(time.Millisecond * 20)
	v, err = recMap.Get("rec1")
	assert.Equal(t, "recID1", v)
}

func Test_ChecksPathOnInit(t *testing.T) {
	_, err := NewFileRecognizerMap("")
	assert.NotNil(t, err)
}

func Test_ReturnDefault(t *testing.T) {
	f := createTempFile(t)
	fmt.Fprint(f, "default: recID\n")
	defer os.Remove(f.Name())
	recMap, err := newFileRecognizerMap(f.Name())
	v, err := recMap.Get("")
	assert.Equal(t, "recID", v)
	assert.Nil(t, err)
}

func Test_GetAll_ReturnsCache(t *testing.T) {
	r, f := load(t)
	defer os.Remove(f.Name())
	assert.NotNil(t, r)
	r.rCache.recognizers = make([]*api.Recognizer, 10)
	r.rCache.needsReload = false

	ri, err := r.GetAll()
	assert.Nil(t, err)
	assert.Equal(t, 10, len(ri))
}

func Test_GetAll_ReturnsCachedError(t *testing.T) {
	r, f := load(t)
	defer os.Remove(f.Name())
	assert.NotNil(t, r)
	r.rCache.lastErr = errors.New("err")
	r.rCache.needsReload = false

	ri, err := r.GetAll()
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(ri))
}

func TestRC_MarkedFor(t *testing.T) {
	r, f := load(t)
	defer os.Remove(f.Name())
	r.rCache.needsReload = false

	fmt.Fprint(f, "rec1: recID1")
	time.Sleep(time.Millisecond * 20)

	assert.Equal(t, true, r.rCache.needsReload)
}

func TestRC_Reloads(t *testing.T) {
	r, f := load(t)
	defer os.Remove(f.Name())
	rc := r.rCache
	pegomock.When(recInfoLoaderMock.Get(pegomock.AnyString())).ThenReturn(&recognizer.Info{Name: "name", Description: "descr"}, nil)
	err := rc.reload(map[string]interface{}{"key1": "v1"})
	assert.Nil(t, err)
	assert.Nil(t, rc.lastErr)
	assert.Equal(t, false, rc.needsReload)
	assert.Equal(t, 1, len(rc.recognizers))
	assert.Equal(t, "key1", rc.recognizers[0].ID)
	assert.Equal(t, "name", rc.recognizers[0].Name)
	assert.Equal(t, "descr", rc.recognizers[0].Description)
}

func TestRC_ReloadsFails(t *testing.T) {
	r, f := load(t)
	defer os.Remove(f.Name())
	rc := r.rCache
	pegomock.When(recInfoLoaderMock.Get(pegomock.AnyString())).ThenReturn(nil, errors.New("err"))
	err := rc.reload(map[string]interface{}{"key1": "v1"})
	assert.NotNil(t, err)
	assert.NotNil(t, rc.lastErr)
	assert.Equal(t, 0, len(rc.recognizers))
}

func TestRC_ReloadsUnique(t *testing.T) {
	r, f := load(t)
	defer os.Remove(f.Name())
	rc := r.rCache
	pegomock.When(recInfoLoaderMock.Get(pegomock.AnyString())).ThenReturn(&recognizer.Info{Name: "name", Description: "descr"}, nil)
	err := rc.reload(map[string]interface{}{"key1": "v1", "key2": "v1"})
	assert.Nil(t, err)
	assert.Nil(t, rc.lastErr)
	assert.Equal(t, 1, len(rc.recognizers))
	recInfoLoaderMock.VerifyWasCalled(pegomock.Once()).Get(pegomock.AnyString())
}

func TestRC_ReloadsSeveral(t *testing.T) {
	r, f := load(t)
	defer os.Remove(f.Name())
	rc := r.rCache
	pegomock.When(recInfoLoaderMock.Get(pegomock.AnyString())).ThenReturn(&recognizer.Info{Name: "name", Description: "descr"}, nil)
	err := rc.reload(map[string]interface{}{"key1": "v1", "key2": "v2"})
	assert.Nil(t, err)
	assert.Nil(t, rc.lastErr)
	assert.Equal(t, 2, len(rc.recognizers))
	recInfoLoaderMock.VerifyWasCalled(pegomock.Twice()).Get(pegomock.AnyString())
}
