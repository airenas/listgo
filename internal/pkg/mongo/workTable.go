package mongo

import (
	"context"

	"bitbucket.org/airenas/listgo/internal/pkg/persistence"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
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
	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ws.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	c := session.Client().Database(store).Collection(workTable)

	_, err = c.InsertOne(ctx, data)
	return err
}

// Get retrieves status from DB
func (ws *WorkSaver) Get(id string) (*persistence.WorkData, error) {
	ctx, cancel := mongoContext()
	defer cancel()

	session, err := ws.SessionProvider.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())
	c := session.Client().Database(store).Collection(workTable)
	var res persistence.WorkData
	err = c.FindOne(ctx, bson.M{"ID": sanitize(id)}).Decode(&res)
	if err != nil {
		return nil, errors.Wrap(err, "can't get data")
	}
	return &res, nil
}
