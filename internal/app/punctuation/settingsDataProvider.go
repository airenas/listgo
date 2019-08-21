package punctuation

import (
	"io"
	"os"
	"path"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

//SettingsDataProviderImpl provides punctuator data from settings
type SettingsDataProviderImpl struct {
	dir  string
	data *Data
}

//NewSettingsDataProviderImpl inits SettingsDataProviderImpl from directory
func NewSettingsDataProviderImpl(dir string) (*SettingsDataProviderImpl, error) {
	res := SettingsDataProviderImpl{}
	res.dir = dir
	var err error
	res.data, err = loadSettings(path.Join(dir, "settings.yml"))
	if err != nil {
		return nil, errors.Wrap(err, "Cannot load settings")
	}
	return &res, nil
}

//GetData gets data
func (p *SettingsDataProviderImpl) GetData() (*Data, error) {
	return p.data, nil
}

//GetVocab return reader to wird vocabulary
func (p *SettingsDataProviderImpl) GetVocab() (io.ReadCloser, error) {
	f, err := os.Open(path.Join(p.dir, "vocabulary"))
	if err != nil {
		return nil, err
	}
	return f, nil
}

func loadSettings(file string) (*Data, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	d := yaml.NewDecoder(f)
	t := Data{}
	err = d.Decode(&t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
