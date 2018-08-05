package api

// TranscriptionResult - get method response in JSON
type TranscriptionResult struct {
	ID             string `json:"id"`
	Error          string `json:"error"`
	Status         string `json:"status"`
	RecognizedText string `json:"recognizedText"`
}
