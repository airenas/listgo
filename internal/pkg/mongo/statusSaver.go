package mongo

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/err"
	"bitbucket.org/airenas/listgo/internal/pkg/persistence"
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

	c, ctx, cancel, err := newColl(ss.SessionProvider, statusTable)
	if err != nil {
		return err
	}
	defer cancel()

	return skipNoDocErr(c.FindOneAndUpdate(ctx, bson.M{"ID": sanitize(ID)},
		bson.M{"$set": bson.M{"status": status.Name(st)}, "$unset": bson.M{
			persistence.StError:     1,
			persistence.StErrorCode: 1}},
		options.FindOneAndUpdate().SetUpsert(true)).Err())
}

// SaveF saves status to DB fields
func (ss *StatusSaver) SaveF(id string, set, unset map[string]interface{}) error {
	cmdapp.Log.Infof("Saving status %s", id)

	c, ctx, cancel, err := newColl(ss.SessionProvider, statusTable)
	if err != nil {
		return err
	}
	defer cancel()

	update, err := makeUpdate(set, unset)
	if err != nil {
		return err
	}

	return skipNoDocErr(c.FindOneAndUpdate(ctx, bson.M{"ID": sanitize(id)}, update,
		options.FindOneAndUpdate().SetUpsert(true)).Err())
}

func makeUpdate(set, unset map[string]interface{}) (bson.M, error) {
	res := bson.M{}
	if (len(set)) > 0 {
		res["$set"] = set
	}
	if (len(unset)) > 0 {
		res["$unset"] = unset
	}
	return res, nil
}

//SaveError saves error to DB
func (ss *StatusSaver) SaveError(ID string, errorStr string) error {
	cmdapp.Log.Infof("Saving error %s: %s", ID, errorStr)

	c, ctx, cancel, err := newColl(ss.SessionProvider, statusTable)
	if err != nil {
		return err
	}
	defer cancel()

	errorCode := ss.errCodeExtractor.Get(errorStr)

	return skipNoDocErr(c.FindOneAndUpdate(ctx, bson.M{"ID": sanitize(ID)},
		bson.M{"$set": bson.M{persistence.StError: errorStr, persistence.StErrorCode: errorCode}},
		options.FindOneAndUpdate().SetUpsert(true)).Err())
}
