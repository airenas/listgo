package mongo

import (
	"context"
	"errors"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/persistence"
	"go.mongodb.org/mongo-driver/bson"
	mgo "go.mongodb.org/mongo-driver/mongo"
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
	cmdapp.Log.Infof("Getting file name by ID %s", id)

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
	if m.File == "" {
		return "", errors.New("Empty file")
	}
	return m.File, nil
}
