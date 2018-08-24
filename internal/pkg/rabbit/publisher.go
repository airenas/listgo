package rabbit

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

//Publisher publish events to rabbit mq broker
type Publisher struct {
	ChannelProvider *ChannelProvider
}

//NewPublisher initializes rabbit publisher
func NewPublisher(provider *ChannelProvider) *Publisher {
	return &Publisher{ChannelProvider: provider}
}

//Publish publish the message
func (sender *Publisher) Publish(id string, topic string) error {
	cmdapp.Log.Infof("Publishing event %s(%s)", topic, id)

	err := sender.ChannelProvider.RunOnChannelWithRetry(func(ch *amqp.Channel) error {
		return ch.Publish(
			topic, // exchange
			"",
			false, // mandatory
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(id),
			})
	})
	if err != nil {
		return errors.Wrap(err, "Can't publish event")
	}
	return nil
}
