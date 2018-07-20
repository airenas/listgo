package upload

// StatusSaver saves the transcription process status
type StatusSaver interface {
	Save(ID string, status string, errorStr string) error
}
