package mongo

import (
	"context"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	ctx, cancel := mongoContext()
	defer cancel()

	session, err := fs.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	c := session.Client().Database(store).Collection(resultTable)

	return c.FindOneAndUpdate(ctx, bson.M{"ID": sanitize(ID)},
		bson.M{"$set": bson.M{"text": result}},
		options.FindOneAndUpdate().SetUpsert(true)).Err()
}
