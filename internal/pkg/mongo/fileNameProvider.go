package mongo

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/persistence"
	"github.com/pkg/errors"
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
	if m.File == "" {
		return "", errors.New("empty file")
	}
	return m.File, nil
}
