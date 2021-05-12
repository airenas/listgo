package mongo

import (
	"context"
	"errors"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/persistence"
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

	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return "", err
	}
	defer session.EndSession(context.Background())

	c := session.Client().Database(store).Collection(requestTable)

	var m persistence.Request
	err = c.FindOne(ctx, bson.M{"ID": id}).Decode(&m)
	if err == mgo.ErrNoDocuments {
		cmdapp.Log.Infof("ID not found %s", id)
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if m.Email == "" {
		return "", errors.New("Empty email")
	}
	return m.Email, nil
}
