package loader

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/airenas/listgo/internal/app/result/api"
	"github.com/airenas/listgo/internal/pkg/cmdapp"
)

// OpenFileFunc declares function to open file by name and return Reader
type OpenFileFunc func(fileName string) (api.File, error)

// LocalFileLoader loads file on local disk
type LocalFileLoader struct {
	// StoragePath is the main folder to save into
	Path         string
	OpenFileFunc OpenFileFunc
}

// NewLocalFileLoader creates LocalFileLoader instance
func NewLocalFileLoader(path string) (*LocalFileLoader, error) {
	cmdapp.Log.Infof("Init Local File Storage at: %s", path)
	if path == "" {
		return nil, errors.New("no path provided")
	}
	f := LocalFileLoader{Path: path, OpenFileFunc: openFile}
	return &f, nil
}

// Load loads file from disk
func (fs LocalFileLoader) Load(name string) (api.File, error) {
	fileName := filepath.Join(fs.Path, name)
	f, err := fs.OpenFileFunc(fileName)
	if err != nil {
		return nil, errors.New("Can not open file " + fileName + ". " + err.Error())
	}
	return f, nil
}

func openFile(fileName string) (api.File, error) {
	return os.Open(fileName)
}
