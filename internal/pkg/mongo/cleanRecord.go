package mongo

import (
	"context"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"go.mongodb.org/mongo-driver/bson"
)

// CleanRecord deletes mongo table record
type CleanRecord struct {
	SessionProvider *SessionProvider
	Table           string
}

//NewCleanRecords creates CleanRecord instances
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

	ctx, cancel := mongoContext()
	defer cancel()

	session, err := fs.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.Background())

	c := session.Client().Database(store).Collection(fs.Table)

	info, err := c.DeleteMany(ctx, bson.M{"ID": ID})
	if err != nil {
		return err
	}
	cmdapp.Log.Infof("Deleted %d", info.DeletedCount)
	return nil
}
