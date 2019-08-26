package punctuation

//Input contains punctuation input text
type Input struct {
	Text string `json:"text"`
}

//Output contains punctuation output text
type Output struct {
	Original   string  `json:"original"`
	Punctuated string  `json:"punctuated"`
	WordIDs    []int32 `json:"wordIDs"`
	PunctIDs   []int32 `json:"punctIDs"`
}
