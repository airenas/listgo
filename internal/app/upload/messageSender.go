package upload

import "bitbucket.org/airenas/listgo/internal/pkg/msgsender"

// MessageSender sends a messages to message broker
type MessageSender interface {
	Send(message msgsender.Message) error
}
