package rabbit

import (
	"encoding/json"
	"sync"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/messages"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

//Sender performs messages sending using rabbit mq broker
type Sender struct {
	ChannelProvider *ChannelProvider
	initialized     bool
	initFunc        *initFunc
	m               sync.Mutex
}

type initFunc func(*ChannelProvider) error

//NewSender initializes rabbit sender
func NewSender(provider *ChannelProvider, f initFunc) *Sender {
	return &Sender{ChannelProvider: provider, initialized: false, initFunc: &f}
}

//Send sends the message
func (sender *Sender) Send(message *messages.Message) error {
	err := initialize(sender)
	if err != nil {
		defer sender.ChannelProvider.Close() // lets init sender again
		return errors.Wrap(err, "Can't initialize sender")
	}
	cmdapp.Log.Infof("Sending message %s(%s)", message.Queue, message.ID)

	msgQueue := messages.QueueMessage{ID: message.ID}
	msgBytes, err := json.Marshal(msgQueue)
	if err != nil {
		return errors.Wrap(err, "Can't marshal message")
	}

	ch, err := sender.ChannelProvider.Channel()
	if err != nil {
		return errors.Wrap(err, "Can't init channel")
	}
	err = ch.Publish(
		"", // exchange
		message.Queue,
		false, // mandatory
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         msgBytes,
			ReplyTo:      message.ReplyQueue,
		})

	if err != nil {
		defer sender.ChannelProvider.Close() // lets init sender again
		return errors.Wrap(err, "Can't send message")
	}
	return nil
}

func initialize(sender *Sender) error {
	sender.m.Lock()
	defer sender.m.Unlock()

	if !sender.initialized && sender.initFunc != nil {
		f := *sender.initFunc
		err := f(sender.ChannelProvider)
		if err != nil {
			return err
		}
		sender.initialized = true
	}
	return nil
}
