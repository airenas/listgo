package config

import (
	"path/filepath"

	"bitbucket.org/airenas/listgo/internal/app/upload/api"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// FileRecognizerMap struct loads config from provided path
type FileRecognizerMap struct {
	Path string
	v    *viper.Viper
}

//NewFileRecognizerMap creates FileRecognizerMap instance
func NewFileRecognizerMap(path string) (*FileRecognizerMap, error) {
	cmdapp.Log.Infof("Init Recognizer Map from: %s", path)
	if path == "" {
		return nil, errors.New("No path provided")
	}
	file := filepath.Join(path, "recognizers.map.yml")
	return newFileRecognizerMap(file)
}

func newFileRecognizerMap(file string) (*FileRecognizerMap, error) {
	cmdapp.Log.Infof("Init Recognizer Map from: %s", file)
	if file == "" {
		return nil, errors.New("No recognizer map file provided")
	}
	f := FileRecognizerMap{}
	f.v = viper.New()
	f.v.SetConfigFile(file)
	f.v.SetConfigType("yml")
	err := f.v.ReadInConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Can't read recognizers map file: "+file)
	}

	f.v.WatchConfig()
	f.v.OnConfigChange(func(e fsnotify.Event) {
		cmdapp.Log.Infof("Config reloaded from: %s", file)
	})
	return &f, nil
}

// Get return recognizer ID by provided key
func (fs *FileRecognizerMap) Get(name string) (string, error) {
	var id string
	if name == "" {
		id = fs.v.GetString("default")
	} else {
		id = fs.v.GetString(name)
	}
	if id == "" {
		return "", api.ErrRecognizerNotFound
	}
	return id, nil
}
