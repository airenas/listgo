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

//PResult keeps punctuation result
type PResult struct {
	WordIDs    []int32
	Punctuated string
	PunctIDs   []int32
}
