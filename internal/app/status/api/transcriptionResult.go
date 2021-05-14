package api

// TranscriptionResult - status method response in JSON
type TranscriptionResult struct {
	ID               string   `json:"id"`
	ErrorCode        string   `json:"errorCode,omitempty"`
	Error            string   `json:"error,omitempty"`
	Status           string   `json:"status"`
	RecognizedText   string   `json:"recognizedText,omitempty"`
	Progress         int32    `json:"progress,omitempty"`
	AudioReady       bool     `json:"audioReady,omitempty"`
	AvailableResults []string `json:"avResults,omitempty"`
}
