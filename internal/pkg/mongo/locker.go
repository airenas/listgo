package mongo

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Locker acquires lock in db
type Locker struct {
	SessionProvider *SessionProvider
}

//NewLocker creates Locker instance
func NewLocker(sessionProvider *SessionProvider) (*Locker, error) {
	f := Locker{SessionProvider: sessionProvider}
	return &f, nil
}

//Lock locks record for sending email
func (ss *Locker) Lock(id string, lockKey string) error {
	cmdapp.Log.Infof("Locking %s: %s", id, lockKey)

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	c := session.DB(store).C(emailTable)

	// make sure we have the record
	_, err = c.Upsert(bson.M{"ID": id, "key": lockKey}, bson.M{"$setOnInsert": bson.M{"status": 0}})
	if err != nil {
		return err
	}

	change := mgo.Change{Update: bson.M{"$set": bson.M{"status": 1}}, ReturnNew: true}
	var lockRecord interface{}
	_, err = c.Find(bson.M{"ID": id, "key": lockKey, "status": 0}).Apply(change, &lockRecord)
	return err
}

//UnLock marks record with specific value
func (ss *Locker) UnLock(id string, lockKey string, value *int) error {
	cmdapp.Log.Infof("Unlocking table %s: %s", id, lockKey)

	session, err := ss.SessionProvider.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	c := session.DB(store).C(emailTable)

	change := mgo.Change{Update: bson.M{"$set": bson.M{"status": *value}}, ReturnNew: true}
	var lockRecord interface{}
	_, err = c.Find(bson.M{"ID": id, "key": lockKey, "status": 1}).Apply(change, &lockRecord)
	cmdapp.LogIf(err)
	return err
}
