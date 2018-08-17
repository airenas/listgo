package upload

import "bitbucket.org/airenas/listgo/internal/pkg/messages"

// MessageSender sends a messages to message broker
type MessageSender interface {
	Send(message *messages.Message) error
}
