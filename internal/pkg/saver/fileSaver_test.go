package saver

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaves(t *testing.T) {
	fakeFile := fakeWriterCloser{bytes.NewBufferString(""), "", false}
	fileSaver := LocalFileSaver{StoragePath: "/data/",
		OpenFileFunc: func(file string) (WriterCloser, error) {
			fakeFile.Name = file
			return &fakeFile, nil
		}}
	err := fileSaver.Save("file", strings.NewReader("body"))
	assert.Nil(t, err)
	assert.Equal(t, fakeFile.String(), "body")
	assert.Equal(t, fakeFile.Name, "/data/file")
	assert.True(t, fakeFile.Closed)
}

func TestFailsOnNoOpen(t *testing.T) {
	fakeFile := fakeWriterCloser{bytes.NewBufferString(""), "", false}
	fileSaver := LocalFileSaver{StoragePath: "",
		OpenFileFunc: func(file string) (WriterCloser, error) {
			return &fakeFile, errors.New("olia")
		}}
	err := fileSaver.Save("file", strings.NewReader("body"))
	assert.NotNil(t, err)
}

func TestChecksDirOnInit(t *testing.T) {
	_, err := NewLocalFileSaver("./")
	assert.Nil(t, err)

	_, err = NewLocalFileSaver("")
	assert.NotNil(t, err)
}

type fakeWriterCloser struct {
	*bytes.Buffer
	Name   string
	Closed bool
}

func (t *fakeWriterCloser) Close() error {
	t.Closed = true
	return nil
}
