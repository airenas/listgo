package file

import (
	"os"

	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
)

//Filer saves working ids to file system
type Filer struct {
	path string
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
	return &res, nil
}

func (f *Filer) FindWorking(kafkaID string) (*kafkaapi.KafkaTrMap, error) {
	return nil, nil
}
func (f *Filer) SetWorking(krIds *kafkaapi.KafkaTrMap) error {
	return nil
}
func (f *Filer) Delete(trID string) error {
	return nil
}
func (f *Filer) GetPending() ([]*kafkaapi.KafkaTrMap, error) {
	return nil, nil
}
