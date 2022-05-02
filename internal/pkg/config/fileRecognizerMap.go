package config

import (
	"path/filepath"
	"sync"

	"bitbucket.org/airenas/listgo/internal/app/upload/api"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/recognizer"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// FileRecognizerMap struct loads config from provided path
type FileRecognizerMap struct {
	Path  string
	v     *viper.Viper
	vLock sync.RWMutex

	rCache *RecognizersCache
}

type infoLoader interface {
	Get(key string) (*recognizer.Info, error)
}

// RecognizersCache struct keeps current recognizer settings
type RecognizersCache struct {
	recognizers []*api.Recognizer
	lastErr     error

	needsReload bool
	lock        sync.Mutex
	fileLoader  infoLoader
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
	cmdapp.Log.Infof("Init Recognizer Map config from: %s", file)
	if file == "" {
		return nil, errors.New("No recognizer map file provided")
	}
	f := FileRecognizerMap{}
	rc := &RecognizersCache{needsReload: true}
	f.rCache = rc
	f.vLock = sync.RWMutex{}
	var err error
	fp := filepath.Dir(file)
	rc.fileLoader, err = NewFileRecognizerInfoLoader(fp)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init recognizers info loader for recognizer cache. Path: "+fp)
	}
	f.v, err = initViper(file)
	if err != nil {
		return nil, err
	}

	// configure reload
	if err := f.addWatcher(file); err != nil {
		return nil, err
	}
	return &f, nil
}

func (f *FileRecognizerMap) addWatcher(file string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					cmdapp.Log.Println("modified file:", event.Name)
					f.onConfigChange(file)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				cmdapp.Log.Error("error:", err)
			}
		}
	}()
	cmdapp.Log.Info("Add watch for :", file)
	return watcher.Add(file)
}

func initViper(file string) (*viper.Viper, error) {
	res := viper.New()
	res.SetConfigFile(file)
	res.SetConfigType("yml")
	if err := res.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "can't read recognizers map file: "+file)
	}
	return res, nil
}

// Get return recognizer ID by provided key
func (fs *FileRecognizerMap) onConfigChange(file string) {
	cmdapp.Log.Infof("Config reload started from '%s'", file)
	fs.vLock.Lock()
	defer fs.vLock.Unlock()

	copyV, err := initViper(file)
	if err != nil {
		cmdapp.Log.Error(err)
		return
	}
	// cache access only with lock
	rc := fs.rCache
	fs.rCache.lock.Lock()
	defer rc.lock.Unlock()
	fs.v = copyV
	rc.needsReload = true
	cmdapp.Log.Infof("Config reloaded")
}

// Get return recognizer ID by provided key
func (fs *FileRecognizerMap) Get(name string) (string, error) {
	fs.vLock.RLock()
	defer fs.vLock.RUnlock()

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

// GetAll returns all information about installed recognizers
func (fs *FileRecognizerMap) GetAll() ([]*api.Recognizer, error) {
	rc := fs.rCache
	rc.lock.Lock()
	defer rc.lock.Unlock()

	if rc.needsReload {
		cmdapp.Log.Info("Reloading recognizers")
		err := rc.reload(fs.v.AllSettings())
		if err != nil {
			return nil, err
		}
	}
	return rc.recognizers, rc.lastErr
}

func (rc *RecognizersCache) reload(m map[string]interface{}) error {
	rc.lastErr = nil
	rc.recognizers = nil
	rc.needsReload = false
	rm := make(map[string]*recognizer.Info)
	res := make([]*api.Recognizer, 0)
	for k, v := range m {
		vs, ok := v.(string)
		if !ok {
			rc.lastErr = errors.New("can't convert vipers value to string")
			return rc.lastErr
		}
		if _, f := rm[vs]; !f {
			var err error
			r, err := rc.fileLoader.Get(vs)
			if err != nil {
				rc.lastErr = errors.Wrap(err, "can't load recognizer for key "+vs)
				return rc.lastErr
			}
			rm[vs] = r
			res = append(res, mapRecognizer(k, r))
		}
	}
	rc.recognizers = res
	return nil
}

func mapRecognizer(k string, r *recognizer.Info) *api.Recognizer {
	res := api.Recognizer{}
	res.ID = k
	res.Name = r.Name
	res.Description = r.Description
	res.DateCreated = r.DateCreated
	return &res
}
