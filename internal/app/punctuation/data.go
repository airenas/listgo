package punctuation

//Input contains punctuation input text
type Input struct {
	Text string `json:"text"`
}

//InputArray contains punctuation input words array
type InputArray struct {
	Words []string `json:"input"`
}

//Output contains punctuation output
type Output struct {
	PunctuatedText string   `json:"punctuatedText"`
	Original       []string `json:"original"`
	Punctuated     []string `json:"punctuated"`
	WordIDs        []int32  `json:"wordIDs"`
	PunctIDs       []int32  `json:"punctIDs"`
}
