package status

//Saver saves the transcription process status
type Saver interface {
	Save(id string, st Status) error
	SaveError(id string, errorStr string) error
	SaveF(id string, set, unset map[string]interface{}) error
}
