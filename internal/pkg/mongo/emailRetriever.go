package mongo

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/persistence"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	mgo "go.mongodb.org/mongo-driver/mongo"
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

	c, ctx, cancel, err := newColl(ss.SessionProvider, requestTable)
	if err != nil {
		return "", err
	}
	defer cancel()

	var m persistence.Request
	err = c.FindOne(ctx, bson.M{"ID": id}).Decode(&m)
	if err == mgo.ErrNoDocuments {
		cmdapp.Log.Infof("ID not found %s", id)
		return "", nil
	}
	if err != nil {
		return "", errors.Wrap(err, "can't get request record")
	}
	if m.Email == "" {
		return "", errors.New("Empty email")
	}
	return m.Email, nil
}
