package mongo

import (
	"bitbucket.org/airenas/listgo/internal/app/status/api"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/err"
	"bitbucket.org/airenas/listgo/internal/pkg/progress"
	"bitbucket.org/airenas/listgo/internal/pkg/status"
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
func (fs StatusProvider) Get(id string) (*api.TranscriptionResult, error) {
	cmdapp.Log.Infof("Retrieving status %s", id)

	session, err := fs.SessionProvider.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	c := session.DB(store).C(statusTable)

	var m bson.M
	err = c.Find(bson.M{"ID": id}).One(&m)
	if err == mgo.ErrNotFound {
		cmdapp.Log.Infof("ID not found %s", id)
		return newNotFoundResult(id), nil
	}

	if err != nil {
		return nil, err
	}

	result := api.TranscriptionResult{ID: id}

	st, ok := m["status"].(string)
	if ok {
		result.Status = st
	}
	errorCodeStr, ok := m["errorCode"].(string)
	if ok {
		result.ErrorCode = errorCodeStr
	}
	errorStr, ok := m["error"].(string)
	if ok {
		result.Error = errorStr
	}
	result.Progress = progress.Convert(result.Status)
	if result.Status == status.Completed.Name {
		result.RecognizedText, err = getResultText(session, id)
	}

	return &result, err
}

// Get retrieves status from DB
func getResultText(session *mgo.Session, id string) (string, error) {
	cmdapp.Log.Infof("Retrieving result %s", id)

	c := session.DB(store).C(resultTable)

	var m bson.M
	err := c.Find(bson.M{"ID": id}).One(&m)
	if err == mgo.ErrNotFound {
		cmdapp.Log.Infof("ID not found %s", id)
		return "", nil
	}
	if err != nil {
		return "", err
	}
	text, ok := m["text"].(string)
	if ok {
		return text, nil
	}
	return "", nil
}

func newNotFoundResult(ID string) *api.TranscriptionResult {
	result := api.TranscriptionResult{ID: ID}
	result.Status = "NOT_FOUND"
	result.ErrorCode = err.NotFoundCode
	result.Error = "Nežinomas ID: " + ID
	return &result
}
