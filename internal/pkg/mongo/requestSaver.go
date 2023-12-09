package mongo

import (
	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/airenas/listgo/internal/pkg/persistence"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RequestSaver saves process request to mongo db
type RequestSaver struct {
	SessionProvider *SessionProvider
}

// NewRequestSaver creates RequestSaver instance
func NewRequestSaver(sessionProvider *SessionProvider) (*RequestSaver, error) {
	f := RequestSaver{SessionProvider: sessionProvider}
	return &f, nil
}

// Save saves resquest to DB
func (ss *RequestSaver) Save(data *persistence.Request) error {
	cmdapp.Log.Infof("Saving request %s: %s", data.ID, data.Email)

	c, ctx, cancel, err := newColl(ss.SessionProvider, requestTable)
	if err != nil {
		return err
	}
	defer cancel()

	return skipNoDocErr(c.FindOneAndUpdate(ctx, bson.M{"ID": sanitize(data.ID)},
		bson.M{"$set": bson.M{"email": data.Email, "file": data.File,
			"externalID": data.ExternalID, "recognizerKey": data.RecognizerKey, "recognizerID": data.RecognizerID}},
		options.FindOneAndUpdate().SetUpsert(true)).Err())
}
