package mongo

import (
	"bitbucket.org/airenas/listgo/internal/app/upload/api"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/globalsign/mgo/bson"
)

// RequestSaver saves process request to mongo db
type RequestSaver struct {
	SessionProvider *SessionProvider
}

//NewRequestSaver creates RequestSaver instance
func NewRequestSaver(sessionProvider *SessionProvider) (*RequestSaver, error) {
	f := RequestSaver{SessionProvider: sessionProvider}
	return &f, nil
}

// Save saves resquest to DB
func (ss *RequestSaver) Save(data api.RequestData) error {
	cmdapp.Log.Infof("Saving request %s: %s", data.ID, data.Email)

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	c := session.DB(store).C(requestTable)
	_, err = c.Upsert(bson.M{"ID": data.ID}, bson.M{"$set": bson.M{"email": data.Email, "file": data.File}})
	return err
}
