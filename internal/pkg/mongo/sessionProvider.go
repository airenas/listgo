package mongo

import (
	"errors"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/globalsign/mgo"
)

//SessionProvider connects and provides session for mongo DB
type SessionProvider struct {
	session *mgo.Session
	URL     string
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
