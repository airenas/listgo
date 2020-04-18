package dispatcher

import (
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"github.com/pkg/errors"
)

type msgWithCorrSender struct {
	realSender messages.SenderWithCorr
	replyQName string
}

func newMsgWithCorrSender(realSender messages.SenderWithCorr, replyQName string) (*msgWithCorrSender, error) {
	res := &msgWithCorrSender{}
	if realSender == nil {
		return nil, errors.New("No realSender provided")
	}
	res.realSender = realSender
	if replyQName == "" {
		return nil, errors.New("No replyQName provided")
	}
	res.replyQName = replyQName
	return res, nil
}

func (sender *msgWithCorrSender) Send(message messages.Message, queue string, corrID string) error {
	return sender.realSender.SendWithCorr(message, queue, sender.replyQName, corrID)
}
