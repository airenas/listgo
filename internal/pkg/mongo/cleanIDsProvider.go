package mongo

import (
	"errors"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// CleanIDsProvider returns old IDs to remove from system
type CleanIDsProvider struct {
	SessionProvider *SessionProvider
	expireDuration  time.Duration
}

//NewCleanIDsProvider creates CleanIDsProvider instances
func NewCleanIDsProvider(sessionProvider *SessionProvider, expireDuration time.Duration) (*CleanIDsProvider, error) {
	f := CleanIDsProvider{SessionProvider: sessionProvider, expireDuration: expireDuration}
	return &f, nil
}

// Get return expired IDs
func (p *CleanIDsProvider) Get() ([]string, error) {
	expDate := time.Now().Add(-p.expireDuration)
	session, err := p.SessionProvider.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	c := session.DB(store).C(requestTable)
	from := 0
	maxRecords := 10
	result := make([]string, 0)
	var m []bson.M
	for {
		err = c.Find(nil).Sort("_id").Skip(from).Limit(maxRecords).All(&m)
		if err == mgo.ErrNotFound {
			return nil, err
		}
		for _, r := range m {
			if p.isOld(r, expDate) {
				id, err := getID(r)
				if err != nil {
					return nil, err
				}
				result = append(result, id)
			} else {
				return result, nil
			}
		}
		from = from + maxRecords
		if from > len(result) {
			return result, nil
		}
		// do futher selection
	}
}

func (p *CleanIDsProvider) isOld(m bson.M, expireDate time.Time) bool {
	id, ok := m["_id"].(bson.ObjectId)
	if !ok {
		return false
	}
	return id.Time().Before(expireDate)
}

func getID(m bson.M) (string, error) {
	id, ok := m["ID"].(string)
	if !ok || id == "" {
		return "", errors.New("Empty ID")
	}
	return id, nil
}
