package manager

import (
	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/airenas/listgo/internal/pkg/messages"
)

// FakeMessageSender is mail sender doing nothing
type FakeMessageSender struct {
}

// NewFakeMessageSender inits new instance
func NewFakeMessageSender() *FakeMessageSender {
	return &FakeMessageSender{}
}

// Send does nothing
func (fms *FakeMessageSender) Send(message messages.Message, queue string, replyQueue string) error {
	cmdapp.Log.Debug("Skip sending message")
	return nil
}
