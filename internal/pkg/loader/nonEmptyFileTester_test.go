package loader

import (
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileLoads(t *testing.T) {
	fileTester := NonEmptyFileTester{Path: "/data/{ID}",
		fileFunc: func(file string) (io.ReadCloser, error) {
			assert.Equal(t, "/data/file", file)
			return ioutil.NopCloser(strings.NewReader("olia")), nil
		}}
	ft, err := fileTester.Test("file")
	assert.Nil(t, err)
	assert.True(t, ft)
}

func TestFileLoads_InitialEmpty(t *testing.T) {
	fileTester := NonEmptyFileTester{Path: "/data/{ID}",
		fileFunc: func(file string) (io.ReadCloser, error) {
			assert.Equal(t, "/data/file", file)
			return ioutil.NopCloser(strings.NewReader("\n\n\n\nolia\n")), nil
		}}
	ft, err := fileTester.Test("file")
	assert.Nil(t, err)
	assert.True(t, ft)
}

func TestFileLoads_Fail(t *testing.T) {
	fileTester := NonEmptyFileTester{Path: "/data/{ID}",
		fileFunc: func(file string) (io.ReadCloser, error) {
			return nil, errors.New("error")
		}}
	_, err := fileTester.Test("file")
	assert.NotNil(t, err)
}

func TestFileLoads_Empty(t *testing.T) {
	var r io.ReadCloser
	fileTester := NonEmptyFileTester{Path: "/data/{ID}",
		fileFunc: func(file string) (io.ReadCloser, error) {
			assert.Equal(t, "/data/file", file)
			return r, nil
		}}
	r = ioutil.NopCloser(strings.NewReader(""))
	ft, _ := fileTester.Test("file")
	assert.False(t, ft)

	r = ioutil.NopCloser(strings.NewReader("   "))
	ft, _ = fileTester.Test("file")
	assert.False(t, ft)

	r = ioutil.NopCloser(strings.NewReader("\n\n\n\n\n   "))
	ft, _ = fileTester.Test("file")
	assert.False(t, ft)
}

func TestFileChecksDirOnInit(t *testing.T) {
	ft, err := NewNonEmptyFileTester("path/{ID}")
	assert.Nil(t, err)
	assert.NotNil(t, ft)
}

func TestFileChecksDirOnInit_Fail(t *testing.T) {
	_, err := NewNonEmptyFileTester("path")
	assert.NotNil(t, err)
	_, err = NewNonEmptyFileTester("path/olia")
	assert.NotNil(t, err)
	_, err = NewLocalFileLoader("")
	assert.NotNil(t, err)
}
