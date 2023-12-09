package punctuation

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/airenas/listgo/internal/app/punctuation/api"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// SettingsDataProviderImpl provides punctuator data from settings
type SettingsDataProviderImpl struct {
	dir  string
	data *api.Data
}

// NewSettingsDataProviderImpl inits SettingsDataProviderImpl from directory
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

// GetData gets data
func (p *SettingsDataProviderImpl) GetData() (*api.Data, error) {
	return p.data, nil
}

// GetVocab return reader to word vocabulary
func (p *SettingsDataProviderImpl) GetVocab() (io.Reader, error) {
	b, err := ioutil.ReadFile(path.Join(p.dir, "vocabulary")) // just pass the file name
	if err != nil {
		return nil, err
	}
	return strings.NewReader(string(b)), nil
}

func loadSettings(file string) (*api.Data, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	d := yaml.NewDecoder(f)
	t := api.Data{}
	err = d.Decode(&t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
