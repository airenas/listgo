package rabbit

import (
	"sync"

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

//NewChannelProvider initializes channel provider
func NewChannelProvider(url string) (*ChannelProvider, error) {
	if url == "" {
		return nil, errors.New("No broker url")
	}
	return &ChannelProvider{url: url}, nil
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
