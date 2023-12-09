package mongo

import (
	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

// CleanRecord deletes mongo table record
type CleanRecord struct {
	SessionProvider *SessionProvider
	Table           string
}

// NewCleanRecords creates CleanRecord instances
func NewCleanRecords(sessionProvider *SessionProvider) ([]*CleanRecord, error) {
	result := make([]*CleanRecord, 0)
	result = append(result, newCleanRecord(sessionProvider, statusTable))
	result = append(result, newCleanRecord(sessionProvider, resultTable))
	result = append(result, newCleanRecord(sessionProvider, emailTable))
	result = append(result, newCleanRecord(sessionProvider, requestTable))
	result = append(result, newCleanRecord(sessionProvider, workTable))
	return result, nil
}

func newCleanRecord(sessionProvider *SessionProvider, table string) *CleanRecord {
	f := CleanRecord{SessionProvider: sessionProvider, Table: table}
	cmdapp.Log.Infof("Init Mongo table Clean for %s", table)
	return &f
}

// Clean deletes record from table by ID
func (fs *CleanRecord) Clean(ID string) error {
	cmdapp.Log.Infof("Cleaning record for for %s[ID=%s]", fs.Table, ID)

	c, ctx, cancel, err := newColl(fs.SessionProvider, fs.Table)
	if err != nil {
		return err
	}
	defer cancel()

	info, err := c.DeleteMany(ctx, bson.M{"ID": ID})
	if err != nil {
		return errors.Wrap(err, "can't delete")
	}
	cmdapp.Log.Infof("Deleted %d", info.DeletedCount)
	return nil
}
