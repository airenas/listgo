package mongo

import (
	"bitbucket.org/airenas/listgo/internal/app/status/api"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// StatusProvider provides transcription status from mongo db
type StatusProvider struct {
	SessionProvider *SessionProvider
}

//NewStatusProvider creates StatusProvider instance
func NewStatusProvider(sessionProvider *SessionProvider) (*StatusProvider, error) {
	f := StatusProvider{SessionProvider: sessionProvider}
	return &f, nil
}

// Get retrieves status from DB
func (fs StatusProvider) Get(ID string) (*api.TranscriptionResult, error) {
	cmdapp.Log.Infof("Retrieving status %s", ID)

	session, err := fs.SessionProvider.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	c := session.DB("store").C("status")

	var m bson.M
	err = c.Find(bson.M{"ID": ID}).One(&m)
	if err == mgo.ErrNotFound {
		cmdapp.Log.Infof("ID not found %s", ID)
		return newNotFoundResult(ID), nil
	}

	if err != nil {
		return nil, err
	}

	result := api.TranscriptionResult{ID: ID}

	status, ok := m["status"].(string)
	if ok {
		result.Status = status
	}
	return &result, err
}

func newNotFoundResult(ID string) *api.TranscriptionResult {
	result := api.TranscriptionResult{ID: ID}
	result.Status = "NOT_FOUND"
	result.Error = "Nežinomas ID: " + ID
	return &result
}
