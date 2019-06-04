package file

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
)

const fileExt string = ".working"

//Filer saves working ids to file system
type Filer struct {
	path         string
	trToKafkaMap map[string]string
	kafkaToTrMap map[string]string
	mutex        *sync.Mutex
}

//NewFiler creates a Filer
func NewFiler() (*Filer, error) {
	res := Filer{}
	res.path = cmdapp.Config.GetString("ids.path")
	if res.path == "" {
		return nil, errors.New("No ids.path setting provided")
	}
	err := os.MkdirAll(res.path, os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err, "Can't init ids directory")
	}
	res.mutex = &sync.Mutex{}
	res.trToKafkaMap = make(map[string]string)
	res.kafkaToTrMap = make(map[string]string)
	return &res, nil
}

//FindWorking return existing working map by kafkaID or nil
func (f *Filer) FindWorking(kafkaID string) (*kafkaapi.KafkaTrMap, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	v, ok := f.kafkaToTrMap[kafkaID]
	if ok {
		return &kafkaapi.KafkaTrMap{TrID: v, KafkaID: kafkaID}, nil
	}
	return nil, nil
}

//SetWorking creates file with id as working one
func (f *Filer) SetWorking(krIds *kafkaapi.KafkaTrMap) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	fp := path.Join(f.path, krIds.TrID+fileExt)
	cmdapp.Log.Info("Creating file:" + fp)
	mf, err := os.Create(fp)
	if err != nil {
		return errors.Wrap(err, "Can't create file "+fp)
	}
	defer mf.Close()
	_, err = mf.WriteString(krIds.KafkaID)
	if err != nil {
		return errors.Wrap(err, "Can't write file "+fp)
	}
	f.trToKafkaMap[krIds.TrID] = krIds.KafkaID
	f.kafkaToTrMap[krIds.KafkaID] = krIds.TrID
	return nil
}

//Delete removes transcription ID from working file indicators
func (f *Filer) Delete(trID string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	v, ok := f.trToKafkaMap[trID]
	if !ok {
		return errors.New("Can't find ids map for " + trID)
	}
	fp := path.Join(f.path, trID+fileExt)
	cmdapp.Log.Info("Deleting file:" + fp)
	err := os.Remove(fp)
	if err != nil {
		return errors.Wrap(err, "Can't delete file "+fp)
	}
	delete(f.trToKafkaMap, trID)
	delete(f.kafkaToTrMap, v)
	return nil
}

//GetPending return all pending tasks. It read all files *.working in directory.
func (f *Filer) GetPending() ([]*kafkaapi.KafkaTrMap, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	res := make([]*kafkaapi.KafkaTrMap, 0)

	err := filepath.Walk(f.path, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), fileExt) {
			trID := info.Name()[:len(info.Name())-len(fileExt)]
			kafkaID, err := getFileData(path)
			if err != nil {
				return err
			}
			res = append(res, &kafkaapi.KafkaTrMap{TrID: trID, KafkaID: kafkaID})
			f.trToKafkaMap[trID] = kafkaID
			f.kafkaToTrMap[kafkaID] = trID
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "Can't read directory "+f.path)
	}
	return res, nil
}

func getFileData(fn string) (string, error) {
	dat, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", errors.Wrap(err, "Can't read "+fn)
	}
	res := string(dat)
	res = strings.TrimSpace(res)
	if res == "" {
		return "", errors.New("No data in " + fn)
	}
	return res, nil
}
