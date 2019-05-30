package api

// TranscriptionResult - status method response in JSON
type TranscriptionResult struct {
	ID             string `json:"id"`
	ErrorCode      string `json:"errorCode"`
	Error          string `json:"error"`
	Status         string `json:"status"`
	RecognizedText string `json:"recognizedText"`
	Progress       int32  `json:"progress"`
}
