package manager

// ResultSaver saves the transcription result into db
type ResultSaver interface {
	Save(ID string, result string) error
}
