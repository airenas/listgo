package api

// RequestData is a struct for input file data
type RequestData struct {
	ID            string
	Email         string
	File          string
	ExternalID    string
	RecognizerKey string
	RecognizerID  string
}
