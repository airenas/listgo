package mongo

import (
	"sync"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/globalsign/mgo"
	"github.com/pkg/errors"
)

//SessionProvider connects and provides session for mongo DB
type SessionProvider struct {
	session *mgo.Session
	URL     string
	m       sync.Mutex // struct field mutex
}

//NewSessionProvider creates Mongo session provider
func NewSessionProvider() (*SessionProvider, error) {
	url := cmdapp.Config.GetString("mongo.url")
	if url == "" {
		return nil, errors.New("No Mongo url provided")
	}
	return &SessionProvider{URL: url}, nil
}

//Close closes mongo session
func (sp *SessionProvider) Close() {
	if sp.session != nil {
		sp.session.Close()
	}
}

//NewSession creates mongo session
func (sp *SessionProvider) NewSession() (*mgo.Session, error) {
	sp.m.Lock()
	defer sp.m.Unlock()

	if sp.session == nil {
		cmdapp.Log.Info("Dial mongo. URL: " + sp.URL)
		session, err := mgo.Dial(sp.URL)
		if err != nil {
			return nil, errors.Wrap(err, "Can't dial to mongo: "+sp.URL)
		}
		err = checkIndex(session, statusTable)
		if err != nil {
			return nil, errors.Wrap(err, "Can't create index: "+statusTable)
		}
		err = checkIndex(session, resultTable)
		if err != nil {
			return nil, errors.Wrap(err, "Can't create index: "+resultTable)
		}
		sp.session = session
	}
	return sp.session.Copy(), nil
}

func checkIndex(s *mgo.Session, table string) error {
	session := s.Copy()
	defer session.Close()
	c := session.DB(store).C(table)
	index := mgo.Index{
		Key:        []string{"ID"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	return c.EnsureIndex(index)
}
