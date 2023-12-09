package rabbit

import (
	"sync"
	"time"

	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/airenas/listgo/internal/pkg/utils"
	"github.com/cenkalti/backoff"
	"github.com/streadway/amqp"

	"github.com/pkg/errors"
)

// ChannelProvider provider amqp channel
type ChannelProvider struct {
	url     string
	conn    *amqp.Connection
	ch      *amqp.Channel
	m       sync.Mutex // struct field mutex
	qPrefix string
}

type runOnChannelFunc func(*amqp.Channel) error

// NewChannelProvider initializes channel provider
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
	prefix := cmdapp.Config.GetString("messageServer.prefix")
	return &ChannelProvider{url: finalURL, qPrefix: prefix}, nil
}

// Channel return cached channel or tries to connect to rabbit broker
func (pr *ChannelProvider) Channel() (*amqp.Channel, error) {
	pr.m.Lock()
	defer pr.m.Unlock()

	if pr.ch != nil {
		return pr.ch, nil
	}
	conn, err := dial(pr.url)
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

// RunOnChannelWithRetry invokes method on channel with retry
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

// Close finalizes ChannelProvider
func (pr *ChannelProvider) Close() {
	pr.m.Lock()
	defer pr.m.Unlock()

	if pr.ch != nil {
		_ = pr.ch.Close()
	}
	if pr.conn != nil {
		_ = pr.conn.Close()
	}
	pr.ch = nil
	pr.conn = nil
}

// QueueName return queue name for channel, may append prefix
func (pr *ChannelProvider) QueueName(name string) string {
	if name == "" {
		return ""
	}
	s := ""
	if pr.qPrefix != "" {
		s = "_"
	}
	return pr.qPrefix + s + name
}

// Healthy checks if rabbit channel is open
func (pr *ChannelProvider) Healthy() error {
	_, err := pr.Channel()
	if err != nil {
		return errors.Wrap(err, "Can't create channel")
	}
	return nil
}

func dial(url string) (*amqp.Connection, error) {
	var res *amqp.Connection
	op := func() error {
		var err error
		cmdapp.Log.Info("Dial " + utils.HidePass(url))
		res, err = amqp.Dial(url)
		return err
	}
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 2 * time.Minute
	err := backoff.Retry(op, bo)
	if err == nil {
		cmdapp.Log.Info("Connected to " + utils.HidePass(url))
	}
	return res, err
}
