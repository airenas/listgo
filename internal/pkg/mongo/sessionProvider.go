package mongo

import (
	"net/url"
	"sync"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/globalsign/mgo"
	"github.com/pkg/errors"
)

//IndexData keeps index creation data
type IndexData struct {
	Table  string
	Field  string
	Unique bool
}

//NewIndexData creates index data
func newIndexData(table string, field string, unique bool) IndexData {
	return IndexData{Table: table, Field: field, Unique: unique}
}

//SessionProvider connects and provides session for mongo DB
type SessionProvider struct {
	session *mgo.Session
	URL     string
	indexes []IndexData
	m       sync.Mutex // struct field mutex
}

//NewSessionProvider creates Mongo session provider
func NewSessionProvider() (*SessionProvider, error) {
	url := cmdapp.Config.GetString("mongo.url")
	if url == "" {
		return nil, errors.New("No Mongo url provided")
	}
	return &SessionProvider{URL: url, indexes: indexData}, nil
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
		cmdapp.Log.Info("Dial mongo: " + hidePass(sp.URL))
		session, err := mgo.Dial(sp.URL)
		if err != nil {
			return nil, errors.Wrap(err, "Can't dial to mongo")
		}
		err = checkIndexes(session, sp.indexes)
		if err != nil {
			return nil, errors.Wrap(err, "Can't create index: "+resultTable)
		}
		sp.session = session
	}
	return sp.session.Copy(), nil
}

func checkIndexes(s *mgo.Session, indexes []IndexData) error {
	session := s.Copy()
	defer session.Close()
	for _, index := range indexes {
		err := checkIndex(s, index)
		if err != nil {
			return errors.Wrap(err, "Can't create index: "+index.Table+":"+index.Field)
		}
	}
	return nil
}

func checkIndex(s *mgo.Session, indexData IndexData) error {
	c := s.DB(store).C(indexData.Table)
	index := mgo.Index{
		Key:        []string{indexData.Field},
		Unique:     indexData.Unique,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	return c.EnsureIndex(index)
}

func hidePass(s string) string {
	u, err := url.Parse(s)
	if err != nil {
		cmdapp.Log.Warn("Can't parse mongo url.")
		return ""
	}
	_, ps := u.User.Password()
	if ps {
		u.User = url.UserPassword(u.User.Username(), "----")
	}
	return u.String()
}
