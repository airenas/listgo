package rabbit

import (
	"sync"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/streadway/amqp"

	"github.com/pkg/errors"
)

//ChannelProvider provider amqp channel
type ChannelProvider struct {
	url  string
	conn *amqp.Connection
	ch   *amqp.Channel
	m    sync.Mutex // struct field mutex
}

type runOnChannelFunc func(*amqp.Channel) error

//NewChannelProvider initializes channel provider
func NewChannelProvider() (*ChannelProvider, error) {
	url := cmdapp.Config.GetString("messageServer.url")
	if url == "" {
		return nil, errors.New("No broker url from messageServer.url")
	}
	user := cmdapp.Config.GetString("messageServer.user")
	pass := cmdapp.Config.GetString("messageServer.pass")
	if user != "" && pass == "" {
		return nil, errors.New("No broker pass from messageServer.pass")
	}
	finalURL := "amqp://"
	if user != "" {
		finalURL = finalURL + user + ":" + pass + "@"
	}
	finalURL = finalURL + url
	return &ChannelProvider{url: finalURL}, nil
}

//Channel return cached channel or tries to connect to rabbit broker
func (pr *ChannelProvider) Channel() (*amqp.Channel, error) {
	pr.m.Lock()
	defer pr.m.Unlock()

	if pr.ch != nil {
		return pr.ch, nil
	}
	conn, err := amqp.Dial(pr.url)
	if err != nil {
		return nil, errors.Wrap(err, "Can't connect to rabbit broker")
	}
	ch, err := conn.Channel()
	if err != nil {
		defer conn.Close()
		return nil, errors.Wrap(err, "Can't create channel")
	}
	pr.conn = conn
	pr.ch = ch
	return pr.ch, nil
}

//RunOnChannelWithRetry invokes method on channel with retry
func (pr *ChannelProvider) RunOnChannelWithRetry(f runOnChannelFunc) error {
	ch, err := pr.Channel()
	if err != nil {
		return errors.Wrap(err, "Can't init channel")
	}
	err = f(ch)
	if err != nil {
		cmdapp.Log.Infof("Retry opening channel")
		pr.Close()
		ch, err = pr.Channel()
		if err != nil {
			return errors.Wrap(err, "Can't init channel")
		}
		err = f(ch)
	}
	return err
}

//Close finalizes ChannelProvider
func (pr *ChannelProvider) Close() {
	pr.m.Lock()
	defer pr.m.Unlock()

	if pr.ch != nil {
		defer pr.ch.Close()
	}
	if pr.conn != nil {
		defer pr.conn.Close()
	}
	pr.ch = nil
	pr.conn = nil
}
