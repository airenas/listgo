package inform

import (
	"net/smtp"
	"strings"
	"time"

	"bitbucket.org/airenas/listgo/internal/app/inform/auth"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/jordan-wright/email"
)

//SimpleEmailSender uses standard esmtp lib to send emails
type SimpleEmailSender struct {
	sendPool *email.Pool
}

func newSimpleEmailSender() (*SimpleEmailSender, error) {
	r := SimpleEmailSender{}
	var err error
	r.sendPool, err = email.NewPool(cmdapp.Config.GetString("smtp.host")+":"+cmdapp.Config.GetString("smtp.port"), 1, newAuth())
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func newAuth() smtp.Auth {
	if strings.ToLower(cmdapp.Config.GetString("smtp.useLogin")) == "true" {
		cmdapp.Log.Infof("Using custom login auth")
		return auth.LoginAuth(cmdapp.Config.GetString("smtp.username"), cmdapp.Config.GetString("smtp.password"))
	}
	cmdapp.Log.Infof("Using plain login auth ")
	return smtp.PlainAuth("", cmdapp.Config.GetString("smtp.username"), cmdapp.Config.GetString("smtp.password"),
		cmdapp.Config.GetString("smtp.host"))
}

//Send sends email
func (s *SimpleEmailSender) Send(email *email.Email) error {
	return s.sendPool.Send(email, 10*time.Second)
}
