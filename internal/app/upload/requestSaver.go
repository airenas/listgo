package upload

// RequestSaver saves the request info to db
type RequestSaver interface {
	Save(ID string, Email string) error
}
