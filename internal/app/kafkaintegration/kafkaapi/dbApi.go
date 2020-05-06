package kafkaapi

//DBEntry keeps structure for AudioGetResponse
type DBEntry struct {
	ID               string
	Data             string
	FileName         string
	JobType          string
	NumberOfSpeakers string
	RecordQuality    string
}

//DBResultEntry keeps structure for TranscriptionPostRequest
type DBResultEntry struct {
	ID            string
	Error         *DBTranscriptionError
	Transcription DBTranscriptionResult
}

//DBTranscriptionError keeps structure for TranscriptionError
type DBTranscriptionError struct {
	Code  string
	Error string
}

//DBTranscriptionResult keeps structure for Result
type DBTranscriptionResult struct {
	Text        string
	LatticeData string
	WebVTT      string
}

//AddDBResultError adds error to object
func AddDBResultError(data *DBResultEntry, code, err string) *DBResultEntry {
	data.Error = &DBTranscriptionError{Code: code, Error: err}
	return data
}
