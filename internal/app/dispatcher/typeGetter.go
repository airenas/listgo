package dispatcher

import (
	"bitbucket.org/airenas/listgo/internal/pkg/config"
	"github.com/pkg/errors"
)

type typeGetter struct {
	recognizerInfo *config.FileRecognizerInfoLoader
	key            string
}

func newTypeGetter(recognizerInfo *config.FileRecognizerInfoLoader, key string) (*typeGetter, error) {
	if recognizerInfo == nil {
		return nil, errors.New("No recognizer Info loader provided")
	}
	if key == "" {
		return nil, errors.New("No key for model type getter")
	}
	return &typeGetter{recognizerInfo: recognizerInfo, key: key}, nil
}

func (g *typeGetter) Get(rec string) (string, error) {
	rd, err := g.recognizerInfo.Get(rec)
	if err != nil {
		return "", err
	}
	mt, f := rd.Settings[g.key]
	if !f {
		return "", errors.Errorf("Key '%s' not found", g.key)
	}
	return mt, nil
}
