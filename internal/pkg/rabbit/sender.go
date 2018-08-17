package rabbit

import (
	"encoding/json"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/messages"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

//Sender performs messages sending using rabbit mq broker
type Sender struct {
	ChannelProvider *ChannelProvider
}

type initFunc func(*ChannelProvider) error

//NewSender initializes rabbit sender
func NewSender(provider *ChannelProvider) *Sender {
	return &Sender{ChannelProvider: provider}
}

//Send sends the message
func (sender *Sender) Send(message *messages.QueueMessage, queue string, replyQueue string) error {
	cmdapp.Log.Infof("Sending message %s(%s)", queue, message.ID)

	msgBytes, err := json.Marshal(message)
	if err != nil {
		return errors.Wrap(err, "Can't marshal message")
	}

	err = sender.ChannelProvider.RunOnChannelWithRetry(func(ch *amqp.Channel) error {
		return ch.Publish(
			"", // exchange
			queue,
			false, // mandatory
			false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "application/json",
				Body:         msgBytes,
				ReplyTo:      replyQueue,
			})
	})
	if err != nil {
		return errors.Wrap(err, "Can't send message")
	}
	return nil
}
