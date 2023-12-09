package mongo

import (
	"context"
	"time"

	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mgo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CleanIDsProvider returns old IDs to remove from system
type CleanIDsProvider struct {
	SessionProvider *SessionProvider
	expireDuration  time.Duration
}

// NewCleanIDsProvider creates CleanIDsProvider instances
func NewCleanIDsProvider(sessionProvider *SessionProvider, expireDuration time.Duration) (*CleanIDsProvider, error) {
	f := CleanIDsProvider{SessionProvider: sessionProvider, expireDuration: expireDuration}
	return &f, nil
}

// Get return expired IDs
func (p *CleanIDsProvider) Get() ([]string, error) {
	expDate := time.Now().Add(-p.expireDuration)
	cmdapp.Log.Infof("Getting old records, time < %s", expDate.String())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	session, err := p.SessionProvider.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(context.Background())

	c := session.Client().Database(store).Collection(requestTable)
	from := int64(0)
	maxRecords := int64(10)
	res := make([]string, 0)
	for {
		cursor, err := c.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"_id": 1}).SetSkip(from).SetLimit(maxRecords))
		if err != nil {
			if err != mgo.ErrNoDocuments {
				return nil, errors.Wrap(err, "can't select from "+requestTable)
			}
			return res, nil
		}
		var recs []bson.M
		if err := cursor.All(ctx, &recs); err != nil {
			return nil, errors.Wrap(err, "can't get data")
		}
		cmdapp.Log.Debugf("Loaded %d records", len(recs))
		for _, r := range recs {
			if p.isOld(r, expDate) {
				id, err := getID(r)
				if err != nil {
					return nil, err
				}
				res = append(res, id)
			} else {
				return res, nil
			}
		}

		from = from + maxRecords
		if from > int64(len(res)) {
			return res, nil
		}
		// do futher selection
	}
}

func (p *CleanIDsProvider) isOld(m bson.M, expireDate time.Time) bool {
	id, ok := m["_id"].(primitive.ObjectID)
	if !ok {
		cmdapp.Log.Warn("_id not found in record")
		return false
	}
	cmdapp.Log.Debugf("_id time %s", id.Timestamp().String())
	return id.Timestamp().Before(expireDate)
}

func getID(m bson.M) (string, error) {
	id, ok := m["ID"].(string)
	if !ok || id == "" {
		return "", errors.New("empty ID")
	}
	return id, nil
}
