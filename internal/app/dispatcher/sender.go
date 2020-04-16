package dispatcher

import (
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/rabbit"
)

type msgWithCorrSender struct {
	realSender *rabbit.Sender
	replyQName string
}

func newMsgWithCorrSender(realSender *rabbit.Sender, replyQName string) (*msgWithCorrSender, error) {
	return &msgWithCorrSender{realSender: realSender, replyQName: replyQName}, nil
}

func (sender *msgWithCorrSender) Send(message messages.Message, queue string, corrID string) error {
	return sender.realSender.SendWithCorreration(message, queue, sender.replyQName, corrID)
}
