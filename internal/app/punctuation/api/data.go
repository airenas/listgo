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
	PunctuatedText string
	Original       []string
	Punctuated     []string
	WordIDs        []int32
	PunctIDs       []int32
}
