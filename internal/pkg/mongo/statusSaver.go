package mongo

import (
	"context"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/err"
	"bitbucket.org/airenas/listgo/internal/pkg/status"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	f := StatusSaver{SessionProvider: sessionProvider, errCodeExtractor: err.CodeExtractor{}}
	return &f, nil
}

// Save saves status to DB
func (ss *StatusSaver) Save(ID string, st status.Status) error {
	cmdapp.Log.Infof("Saving status %s: %s", ID, status.Name(st))

	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	c := session.Client().Database(store).Collection(statusTable)

	return c.FindOneAndUpdate(ctx, bson.M{"ID": sanitize(ID)},
		bson.M{"$set": bson.M{"status": status.Name(st)}, "$unset": bson.M{"error": 1, "errorCode": 1}},
		options.FindOneAndUpdate().SetUpsert(true)).Err()
}

//SaveError saves error to DB
func (ss *StatusSaver) SaveError(ID string, errorStr string) error {
	cmdapp.Log.Infof("Saving error %s: %s", ID, errorStr)

	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	c := session.Client().Database(store).Collection(statusTable)
	errorCode := ss.errCodeExtractor.Get(errorStr)

	return c.FindOneAndUpdate(ctx, bson.M{"ID": sanitize(ID)},
		bson.M{"$set": bson.M{"error": errorStr, "errorCode": errorCode}},
		options.FindOneAndUpdate().SetUpsert(true)).Err()
}
