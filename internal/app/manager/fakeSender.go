package manager

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
)

type fakeMessageSender struct {
}

func newFakeMessageSender() *fakeMessageSender {
	return &fakeMessageSender{}
}

func (fms *fakeMessageSender) Send(message messages.Message, queue string, replyQueue string) error {
	cmdapp.Log.Debug("Skip sending message")
	return nil
}
