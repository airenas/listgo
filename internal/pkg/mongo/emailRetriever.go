package mongo

import (
	"errors"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// EmailRetriever returns email by transcription ID
type EmailRetriever struct {
	SessionProvider *SessionProvider
}

//NewEmailRetriever creates EmailRetriever instance
func NewEmailRetriever(sessionProvider *SessionProvider) (*EmailRetriever, error) {
	f := EmailRetriever{SessionProvider: sessionProvider}
	return &f, nil
}

//Get returns email by ID
func (ss *EmailRetriever) Get(id string) (string, error) {
	cmdapp.Log.Infof("Getting email by ID %s", id)

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
	email, ok := m["email"].(string)
	if !ok || email == "" {
		return "", errors.New("Empty email")
	}
	return email, nil
}
