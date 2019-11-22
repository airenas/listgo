package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"bitbucket.org/airenas/listgo/internal/app/upload/api"
	"github.com/stretchr/testify/assert"
)

func createTempFile(t *testing.T) *os.File {
	f, err := ioutil.TempFile("", "test")
	assert.Nil(t, err)
	return f
}

func load(t *testing.T) (*FileRecognizerMap, *os.File) {
	f := createTempFile(t)
	fmt.Fprint(f, "rec: recID")
	r, err := newFileRecognizerMap(f.Name())
	assert.Nil(t, err)
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
	time.Sleep(time.Millisecond * 10)
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
