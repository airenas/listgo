package mongo

import (
	"bitbucket.org/airenas/listgo/internal/pkg/persistence"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// WorkSaver saves process work data to mongo db
type WorkSaver struct {
	SessionProvider *SessionProvider
}

//NewStatusSaver creates StatusSaver instance
func NewWorkSaver(sessionProvider *SessionProvider) (*WorkSaver, error) {
	f := WorkSaver{SessionProvider: sessionProvider}
	return &f, nil
}

// Save saves data to DB
func (ws *WorkSaver) Save(data *persistence.WorkData) error {
	c, ctx, cancel, err := newColl(ws.SessionProvider, workTable)
	if err != nil {
		return err
	}
	defer cancel()

	return skipNoDocErr(c.FindOneAndUpdate(ctx, bson.M{"ID": sanitize(data.ID)},
		bson.M{"$set": bson.M{"related": data.Related, "fileNames": data.FileNames}},
		options.FindOneAndUpdate().SetUpsert(true)).Err())
}

// Get retrieves status from DB
func (ws *WorkSaver) Get(id string) (*persistence.WorkData, error) {
	c, ctx, cancel, err := newColl(ws.SessionProvider, workTable)
	if err != nil {
		return nil, err
	}
	defer cancel()

	var res persistence.WorkData
	err = c.FindOne(ctx, bson.M{"ID": sanitize(id)}).Decode(&res)
	if err != nil {
		return nil, errors.Wrapf(err, "can't get data by ID = %s", id)
	}
	return &res, nil
}
