package file

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
)

// Filer saves working ids to file system
type Filer struct {
	path string
}

// NewFiler creates a Filer
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
	return &res, nil
}

// Find returns existing working map by kafkaID or nil
func (f *Filer) Find(kafkaID string) (*kafkaapi.KafkaTrMap, error) {
	fn := f.makePath(kafkaID)
	if _, err := os.Stat(fn); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
	}
	trID, err := getFileData(fn)
	if err != nil {
		return nil, err
	}
	return &kafkaapi.KafkaTrMap{TrID: trID, KafkaID: kafkaID}, nil
}

func (f *Filer) makePath(kafkaID string) string {
	return path.Join(f.path, kafkaID)
}

// SetWorking creates file with id as working one
func (f *Filer) SetWorking(krIds *kafkaapi.KafkaTrMap) error {
	fn := f.makePath(krIds.KafkaID)
	cmdapp.Log.Info("Creating file:" + fn)
	mf, err := os.Create(fn)
	if err != nil {
		return errors.Wrap(err, "Can't create file "+fn)
	}
	defer mf.Close()
	_, err = mf.WriteString(krIds.TrID)
	if err != nil {
		return errors.Wrap(err, "Can't write file "+fn)
	}
	return nil
}

// Delete removes transcription ID from working file indicators
func (f *Filer) Delete(kafkaID string) error {
	fp := f.makePath(kafkaID)
	cmdapp.Log.Info("Deleting file:" + fp)
	err := os.Remove(fp)
	if err != nil {
		return errors.Wrap(err, "Can't delete file "+fp)
	}
	return nil
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
