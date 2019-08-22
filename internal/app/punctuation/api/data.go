package api

//Data keeps punctuation settings
type Data struct {
	Info                  string
	PunctuationVovabulary []string `yaml:"puctuationVocabulary,flow"`
	SentenceEnd           []string `yaml:"sentenceEnd,flow"`
	Timesteps             int      `yaml:"timesteps"`
	UnknownWord           string   `yaml:"unknownWord"`
	SequenceEndWord       string   `yaml:"sequenceEndWord"`
}
