package manager

import (
	"bitbucket.org/airenas/listgo/internal/pkg/msgsender"
	"github.com/RichardKnop/machinery/v1/backends/result"
)

// MessageSender sends a messages to message broker
type MessageSender interface {
	Send(message *msgsender.Message) (*result.AsyncResult, error)
}
