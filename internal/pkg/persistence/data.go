package persistence

const (
	StAudioReady       = "audioReady"
	StAvailableResults = "avResults"
)

type (
	WorkData struct {
		ID        string   `json:"ID"`
		Related   []string `json:"related,omitempty"`
		FileNames []string `json:"fileNames,omitempty"`
	}

	Status struct {
		ID               string   `json:"ID"`
		Status           string   `json:"status,omitempty"`
		Error            string   `json:"error,omitempty"`
		ErrorCode        string   `json:"errorCode,omitempty"`
		AudioReady       bool     `json:"audioReady,omitempty"`
		AvailableResults []string `json:"avResults,omitempty"`
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
