package kafkaapi;

//DBEntry keeps structure for AudioGetResponse
type DBEntry struct {
	ID string
	Data string
	FileName string
	JobType string
}

//DBResultEntry keeps structure for TranscriptionPostRequest
type DBResultEntry struct {
	ID string
	Status string
	Err DBTranscriptionError
	Transcription DBTranscriptionResult
}

//DBTranscriptionError keeps structure for TranscriptionError
type DBTranscriptionError struct {
	Code string
	Error string
}

//DBTranscriptionResult keeps structure for Result
type DBTranscriptionResult struct {
	Text string
	ResultFile string
}
