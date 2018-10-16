package inform

import (
	"net/smtp"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/jordan-wright/email"
)

type SimpleEmailSender struct {
	sendPool *email.Pool
}

func newSimpleEmailSender() (*SimpleEmailSender, error) {
	r := SimpleEmailSender{}
	var err error
	r.sendPool, err = email.NewPool(cmdapp.Config.GetString("smtp.host")+":"+cmdapp.Config.GetString("smtp.port"), 1,
		smtp.PlainAuth("", cmdapp.Config.GetString("smtp.username"), cmdapp.Config.GetString("smtp.password"), cmdapp.Config.GetString("smtp.host")))
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *SimpleEmailSender) Send(email *email.Email) error {
	return s.sendPool.Send(email, 10*time.Second)
}
