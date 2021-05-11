package manager

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
)

type FakeMessageSender struct {
}

func NewFakeMessageSender() *FakeMessageSender {
	return &FakeMessageSender{}
}

func (fms *FakeMessageSender) Send(message messages.Message, queue string, replyQueue string) error {
	cmdapp.Log.Debug("Skip sending message")
	return nil
}
