package mongo

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/err"
	"bitbucket.org/airenas/listgo/internal/pkg/status"
	"github.com/globalsign/mgo/bson"
)

type errCodeExtractor interface {
	Get(string) string
}

// StatusSaver saves process status to mongo db
type StatusSaver struct {
	SessionProvider  *SessionProvider
	errCodeExtractor errCodeExtractor
}

//NewStatusSaver creates StatusSaver instance
func NewStatusSaver(sessionProvider *SessionProvider) (*StatusSaver, error) {
	f := StatusSaver{SessionProvider: sessionProvider, errCodeExtractor: err.ErrCodeExtractor{}}
	return &f, nil
}

// Save saves status to DB
func (ss *StatusSaver) Save(ID string, status status.Status) error {
	cmdapp.Log.Infof("Saving status %s: %s", ID, status.Name)

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	c := session.DB(store).C(statusTable)
	_, err = c.Upsert(bson.M{"ID": ID},
		bson.M{"$set": bson.M{"status": status.Name}, "$unset": bson.M{"error": 1, "errorCode": 1}})
	return err
}

//SaveError saves error to DB
func (ss *StatusSaver) SaveError(id string, errorStr string) error {
	cmdapp.Log.Infof("Saving error %s: %s", id, errorStr)

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	errorCode := ss.errCodeExtractor.Get(errorStr)
	c := session.DB(store).C(statusTable)
	_, err = c.Upsert(
		bson.M{"ID": id},
		bson.M{"$set": bson.M{"error": errorStr, "errorCode": errorCode}},
	)
	return err
}
