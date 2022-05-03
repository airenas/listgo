package persistence

const (
	// StAudioReady status table field for audioReady
	StAudioReady = "audioReady"
	// StError status table field for error
	StError = "error"
	// StErrorCode status table field for erorCode
	StErrorCode = "errorCode"
	// StAvailableResults status table field for available Results
	StAvailableResults = "avResults"
)

type (
	// WorkData keeps related IDs for multi transcription job
	WorkData struct {
		ID        string   `json:"ID"`
		Related   []string `json:"related,omitempty"`
		FileNames []string `json:"fileNames,omitempty"`
	}

	// Status keeps job status
	Status struct {
		ID               string   `bson:"ID"`
		Status           string   `bson:"status,omitempty"`
		Error            string   `bson:"error,omitempty"`
		ErrorCode        string   `bson:"errorCode,omitempty"`
		AudioReady       bool     `bson:"audioReady,omitempty"`
		AvailableResults []string `bson:"avResults,omitempty"`
	}

	// Result is table for the final text
	Result struct {
		ID   string `json:"ID"`
		Text string `json:"text,omitempty"`
	}
	// Request is table for initial request info
	Request struct {
		ID            string `json:"ID"`
		Email         string `json:"email,omitempty"`
		File          string `json:"file,omitempty"`
		ExternalID    string `json:"externalID,omitempty"`
		RecognizerKey string `json:"recognizerKey,omitempty"`
		RecognizerID  string `json:"recognizerID,omitempty"`
	}
)
