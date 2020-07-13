package api

//Data keeps punctuation settings
type Data struct {
	Info                  string
	PunctuationVocabulary []string `yaml:"puctuationVocabulary,flow"`
	SentenceEnd           []string `yaml:"sentenceEnd,flow"`
	Timesteps             int      `yaml:"timesteps"`
	UnknownWord           string   `yaml:"unknownWord"`
	SequenceEndWord       string   `yaml:"sequenceEndWord"`
	NumdWord              string   `yaml:"numWord"`
	Model                 string   `yaml:"model"`
	Version               int      `yaml:"version"`
}

//PResult keeps punctuation result
type PResult struct {
	PunctuatedText string
	Original       []string
	Punctuated     []string
	WordIDs        []int32
	PunctIDs       []int32
}
