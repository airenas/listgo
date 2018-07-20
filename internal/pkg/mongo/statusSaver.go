package mongo

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
)

// StatusSaver saves process status to mongo db
type StatusSaver struct {
	SessionProvider *SessionProvider
}

//NewStatusSaver creates StatusSaver instance
func NewStatusSaver(sessionProvider *SessionProvider) (*StatusSaver, error) {
	f := StatusSaver{SessionProvider: sessionProvider}
	return &f, nil
}

// Save saves status to DB
func (fs StatusSaver) Save(ID string, status string, errorStr string) error {
	// fileName := fs.StoragePath + name
	// f, err := fs.OpenFileFunc(fileName)
	// if err != nil {
	// 	return errors.New("Can not create file " + fileName + ". " + err.Error())
	// }
	// defer f.Close()
	// savedBytes, err := io.Copy(f, reader)
	// if err != nil {
	// 	return errors.New("Can not save file " + fileName + ". " + err.Error())
	// }
	cmdapp.Log.Infof("Saving status %s: %s (%s)", ID, status, errorStr)
	return nil
}
