package loader

import (
	"errors"
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"

	"bitbucket.org/airenas/listgo/internal/app/result/api"
	"github.com/stretchr/testify/assert"
)

func TestLoads(t *testing.T) {
	fakeFile := fakeFile("content")
	fileLoader := LocalFileLoader{Path: "/data/",
		OpenFileFunc: func(file string) (api.File, error) {
			return fakeFile, nil
		}}
	f, err := fileLoader.Load("file")
	assert.Nil(t, err)
	assert.NotNil(t, f)
}

func TestFailsOnNoOpen(t *testing.T) {
	fileLoader := LocalFileLoader{Path: "",
		OpenFileFunc: func(file string) (api.File, error) {
			return nil, errors.New("olia")
		}}
	_, err := fileLoader.Load("file")
	assert.NotNil(t, err)
}

func TestChecksDirOnInit(t *testing.T) {
	_, err := NewLocalFileLoader("./")
	assert.Nil(t, err)
	_, err = NewLocalFileLoader("")
	assert.NotNil(t, err)
}

func fakeFile(c string) api.File {
	return mocks.NewMockFile()
}
