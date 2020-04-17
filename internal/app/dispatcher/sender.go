package dispatcher

import (
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
)

type msgWithCorrSender struct {
	realSender messages.SenderWithCorr
	replyQName string
}

func newMsgWithCorrSender(realSender messages.SenderWithCorr, replyQName string) (*msgWithCorrSender, error) {
	return &msgWithCorrSender{realSender: realSender, replyQName: replyQName}, nil
}

func (sender *msgWithCorrSender) Send(message messages.Message, queue string, corrID string) error {
	return sender.realSender.SendWithCorr(message, queue, sender.replyQName, corrID)
}
