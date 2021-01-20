package inform

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"bitbucket.org/airenas/listgo/internal/app/inform/auth"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/jordan-wright/email"
	"github.com/pkg/errors"
)

const (
	//SMTP_NOAUTH login using no authentication
	SMTP_NOAUTH = "NO_AUTH"
	//SMTP_PLAIN login using PLAIN authentication for google
	SMTP_PLAIN = "PLAIN_AUTH"
	//SMTP_LOGIN login using no authentication for other
	SMTP_LOGIN = "LOGIN"
)

//SimpleEmailSender uses standard esmtp lib to send emails
type SimpleEmailSender struct {
	sendPool *email.Pool
	authType string
	host     string
	port     int
}

func newSimpleEmailSender() (*SimpleEmailSender, error) {
	r := SimpleEmailSender{}
	var err error
	r.authType, err = getType(cmdapp.Config.GetString("smtp.type"))
	if err != nil {
		return nil, errors.Wrap(err, "Can't init smtp authentication type")
	}
	r.host = cmdapp.Config.GetString("smtp.host")
	if r.host == "" {
		return nil, errors.New("No smtp host")
	}
	r.port = cmdapp.Config.GetInt("smtp.port")
	if r.port <= 0 {
		return nil, errors.New("No smtp port")
	}
	if r.authType != SMTP_NOAUTH {
		r.sendPool, err = email.NewPool(r.getFullHost(), 1, newAuth())
		if err != nil {
			return nil, err
		}
	}
	cmdapp.Log.Infof("SMTP auth type: %s", r.authType)
	cmdapp.Log.Infof("SMTP server: %s", r.getFullHost())
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
	if s.authType == SMTP_NOAUTH {
		return email.Send(s.getFullHost(), nil)
	}
	return s.sendPool.Send(email, 10*time.Second)
}

func (s *SimpleEmailSender) getFullHost() string {
	return fmt.Sprintf("%s:%d", s.host, s.port)
}

func getType(s string) (string, error) {
	su := strings.TrimSpace(strings.ToUpper(s))
	if su == "" {
		return SMTP_PLAIN, nil
	}
	values := []string{SMTP_NOAUTH, SMTP_PLAIN, SMTP_LOGIN}
	for _, st := range values {
		if st == su {
			return su, nil
		}
	}
	return "", errors.Errorf("Unknown smtp type '%s'. Allowed values: %v", s, values)
}
