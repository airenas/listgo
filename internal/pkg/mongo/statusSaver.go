package mongo

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/globalsign/mgo/bson"
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
	cmdapp.Log.Infof("Saving status %s: %s (%s)", ID, status, errorStr)

	session, err := fs.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	c := session.DB("store").C("status")
	_, err = c.Upsert(
		bson.M{"ID": ID},
		bson.M{"$set": bson.M{"status": status}},
	)
	return err
}
