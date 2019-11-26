package config

import (
	"io/ioutil"
	"path/filepath"

	"bitbucket.org/airenas/listgo/internal/pkg/recognizer"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// FileRecognizerInfoLoader struct loads config from provided path
type FileRecognizerInfoLoader struct {
	Path string
}

//NewFileRecognizerInfoLoader creates FileRecognizerInfoLoader instance
func NewFileRecognizerInfoLoader(path string) (*FileRecognizerInfoLoader, error) {
	cmdapp.Log.Infof("Init Recognizer Info Loader from: %s", path)
	if path == "" {
		return nil, errors.New("No path provided")
	}
	f := FileRecognizerInfoLoader{Path: path}
	return &f, nil
}

// Get return recognizer Info by provided key from file key + '.yml'
func (fs *FileRecognizerInfoLoader) Get(key string) (*recognizer.Info, error) {
	if key == "" {
		return nil, errors.New("No recognizer key provided")
	}
	file := filepath.Join(fs.Path, key+".yml")
	return loadFile(file)
}

func loadFile(file string) (*recognizer.Info, error) {
	fData, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "Can't load: "+file)
	}
	ri, err := loadYaml(fData)
	if err != nil {
		return nil, errors.Wrap(err, "Can't load: "+file)
	}
	return ri, nil
}

func loadYaml(data []byte) (*recognizer.Info, error) {
	ri := recognizer.Info{}
	err := yaml.Unmarshal(data, &ri)
	if err != nil {
		return nil, errors.Wrap(err, "Can't unmarshal")
	}
	if ri.Name == "" {
		return nil, errors.New("No recognizer name in yaml")
	}
	return &ri, nil
}
