package mongo

import (
	"errors"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// FileNameProvider returns file name by transcription ID
type FileNameProvider struct {
	SessionProvider *SessionProvider
}

//NewFileNameProvider creates FileNameProvider instance
func NewFileNameProvider(sessionProvider *SessionProvider) (*FileNameProvider, error) {
	f := FileNameProvider{SessionProvider: sessionProvider}
	return &f, nil
}

//Get returns filename by ID
func (ss *FileNameProvider) Get(id string) (string, error) {
	cmdapp.Log.Infof("Geting email by ID %s", id)

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	c := session.DB(store).C(requestTable)
	var m bson.M
	err = c.Find(bson.M{"ID": id}).One(&m)
	if err == mgo.ErrNotFound {
		cmdapp.Log.Infof("ID not found %s", id)
		return "", nil
	}
	if err != nil {
		return "", err
	}
	result, ok := m["file"].(string)
	if !ok || result == "" {
		return "", errors.New("Empty file")
	}
	return result, nil
}
