package persistence

type (
	WorkData struct {
		ID      string   `json:"ID"`
		Related []string `json:"related,omitempty"`
	}

	Status struct {
		ID          string `json:"ID"`
		Status      string `json:"status,omitempty"`
		Error       string `json:"error,omitempty"`
		ErrorCode   string `json:"errorCode,omitempty"`
		InFileReady bool   `json:"inFileReady,omitempty"`
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
