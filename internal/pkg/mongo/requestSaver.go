package mongo

import (
	"context"

	"bitbucket.org/airenas/listgo/internal/app/upload/api"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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
func (ss *RequestSaver) Save(data *api.RequestData) error {
	cmdapp.Log.Infof("Saving request %s: %s", data.ID, data.Email)

	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	c := session.Client().Database(store).Collection(requestTable)

	return c.FindOneAndUpdate(ctx, bson.M{"ID": sanitize(data.ID)},
		bson.M{"$set": bson.M{"email": data.Email, "file": data.File,
			"externalID": data.ExternalID, "recognizerKey": data.RecognizerKey, "recognizerID": data.RecognizerID}},
		options.FindOneAndUpdate().SetUpsert(true)).Err()
}
