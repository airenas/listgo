package persistence

const (
	StAudioReady       = "audioReady"
	StError            = "error"
	StErrorCode        = "errorCode"
	StAvailableResults = "avResults"
)

type (
	WorkData struct {
		ID        string   `json:"ID"`
		Related   []string `json:"related,omitempty"`
		FileNames []string `json:"fileNames,omitempty"`
	}

	Status struct {
		ID               string   `bson:"ID"`
		Status           string   `bson:"status,omitempty"`
		Error            string   `bson:"error,omitempty"`
		ErrorCode        string   `bson:"errorCode,omitempty"`
		AudioReady       bool     `bson:"audioReady,omitempty"`
		AvailableResults []string `bson:"avResults,omitempty"`
	}

	Result struct {
		ID   string `json:"ID"`
		Text string `json:"text,omitempty"`
	}

	Request struct {
		ID            string `json:"ID"`
		Email         string `json:"email,omitempty"`
		File          string `json:"file,omitempty"`
		ExternalID    string `json:"externalID,omitempty"`
		RecognizerKey string `json:"recognizerKey,omitempty"`
		RecognizerID  string `json:"recognizerID,omitempty"`
	}
)
