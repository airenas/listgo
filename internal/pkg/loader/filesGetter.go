package loader

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
)

// LocalFileList loads file list from local disk dir.
type LocalFileList struct {
	// Path is the main folder to start look from
	Path string
}

//NewLocalFileList creates LocalFileList instance.
func NewLocalFileList(path string) (*LocalFileList, error) {
	cmdapp.Log.Infof("Init LocalFileList at: %s", path)
	if path == "" {
		return nil, errors.New("no path provided")
	}
	f := LocalFileList{Path: path}
	return &f, nil
}

// Load loads file from disk
func (fs *LocalFileList) List(dir string) ([]string, error) {
	fileName := filepath.Join(fs.Path, dir)
	var files []string
	fileInfo, err := ioutil.ReadDir(fileName)
	if err != nil {
		return files, err
	}
	for _, file := range fileInfo {
		ext := filepath.Ext(file.Name())
		if !file.IsDir() && (ext == ".mp4" || ext == ".wav") {
			files = append(files, filepath.Join(dir, file.Name()))
		}
	}
	return files, nil
}
