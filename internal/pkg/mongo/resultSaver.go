package mongo

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/globalsign/mgo/bson"
)

// ResultSaver saves process status to mongo db
type ResultSaver struct {
	SessionProvider *SessionProvider
}

//NewResultSaver creates ResultSaver instance
func NewResultSaver(sessionProvider *SessionProvider) (*ResultSaver, error) {
	f := ResultSaver{SessionProvider: sessionProvider}
	return &f, nil
}

// Save saves result to DB
func (fs *ResultSaver) Save(ID string, result string) error {
	cmdapp.Log.Infof("Saving result for %s", ID)

	session, err := fs.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	c := session.DB(store).C(resultTable)
	_, err = c.Upsert(
		bson.M{"ID": ID},
		bson.M{"$set": bson.M{"text": result}},
	)
	return err
}
