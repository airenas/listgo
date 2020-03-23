package kafkaapi

//UploadData keeps structure for upload method
type UploadData struct {
	ExternalID       string
	AudioData        string
	FileName         string
	JobType          string
	NumberOfSpeakers string
	RecordQuality    string
}

//Status keeps structure for transcription status
type Status struct {
	ID        string
	Text      string
	Completed bool
	ErrorCode string
	Error     string
}

//Result keeps structure for transcription result
type Result struct {
	ID       string
	FileData string
}
