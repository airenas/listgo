package mongo

import (
	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ResultSaver saves process status to mongo db
type ResultSaver struct {
	SessionProvider *SessionProvider
}

// NewResultSaver creates ResultSaver instance
func NewResultSaver(sessionProvider *SessionProvider) (*ResultSaver, error) {
	f := ResultSaver{SessionProvider: sessionProvider}
	return &f, nil
}

// Save saves result to DB
func (fs *ResultSaver) Save(ID string, result string) error {
	cmdapp.Log.Infof("Saving result for %s", ID)

	c, ctx, cancel, err := newColl(fs.SessionProvider, resultTable)
	if err != nil {
		return err
	}
	defer cancel()

	return skipNoDocErr(c.FindOneAndUpdate(ctx, bson.M{"ID": sanitize(ID)},
		bson.M{"$set": bson.M{"text": result}},
		options.FindOneAndUpdate().SetUpsert(true)).Err())
}
