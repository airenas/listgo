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
func (sender *Sender) Send(message messages.Message, queue string, replyQueue string) error {
	return sender.SendWithCorreration(message, queue, replyQueue, "")
}

//SendWithCorreration sends the message
func (sender *Sender) SendWithCorreration(message messages.Message, queue string, replyQueue string, corrID string) error {
	realQueue := sender.ChannelProvider.QueueName(queue)
	cmdapp.Log.Infof("Sending message to %s", realQueue)

	msgBytes, err := json.Marshal(message)
	if err != nil {
		return errors.Wrap(err, "Can't marshal message")
	}

	err = sender.ChannelProvider.RunOnChannelWithRetry(func(ch *amqp.Channel) error {
		return ch.Publish(
			"", // exchange
			realQueue,
			false, // mandatory
			false,
			amqp.Publishing{
				DeliveryMode:  amqp.Persistent,
				ContentType:   "application/json",
				Body:          msgBytes,
				ReplyTo:       replyQueue,
				CorrelationId: corrID,
			})
	})
	if err != nil {
		return errors.Wrap(err, "Can't send message")
	}
	return nil
}
