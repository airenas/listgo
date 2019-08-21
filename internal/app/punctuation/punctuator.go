package punctuation

import "io"

//Data keeps punctuation settings
type Data struct {
	Info                  string
	PunctuationVovabulary []string `yaml:"puctuationVocabulary,flow"`
	SentenceEnd           []string `yaml:"sentenceEnd,flow"`
	Timesteps             int      `yaml:"timesteps"`
}

//DataProvider provides data to initializer
type DataProvider interface {
	GetVocab() (io.ReadCloser, error)
	GetData() (*Data, error)
}

//PunctuatorImpl implements punctuator service
type PunctuatorImpl struct {
	vocab     map[string]int32
	puncVocab map[int32]string
	timesteps int32
}

//NewPunctuatorImpl creates instance
func NewPunctuatorImpl(d DataProvider) (*PunctuatorImpl, error) {
	return &PunctuatorImpl{}, nil
}

//Process is main Punctuator method
func (p *PunctuatorImpl) Process(text string) (string, error) {
	return text, nil
}
