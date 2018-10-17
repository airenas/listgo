package inform

import (
	"errors"
	"strings"

	"github.com/spf13/viper"

	"github.com/jordan-wright/email"
)

type SimpleEmailMaker struct {
	url string
	c   *viper.Viper
}

func newSimpleEmailMaker(c *viper.Viper) (*SimpleEmailMaker, error) {
	r := SimpleEmailMaker{c: c}
	var err error
	r.url, err = getStringNonNil(c, "mail.url")
	if err != nil {
		return nil, err
	}
	return &r, nil
}

//Make prepares the email for ID
func (maker *SimpleEmailMaker) Make(data *Data) (*email.Email, error) {
	return maker.make(data, maker.c)
}

func (maker *SimpleEmailMaker) getText(data *Data, c *viper.Viper) (string, error) {
	r, err := getStringNonNil(c, "mail."+data.msgType+".text")
	if err != nil {
		return "", err
	}
	url := strings.Replace(maker.url, "{{ID}}", data.id, -1)
	r = strings.Replace(r, "{{ID}}", data.id, -1)
	r = strings.Replace(r, "{{URL}}", url, -1)
	t := data.msgTime.Format("2006-01-02 15:04:05")
	r = strings.Replace(r, "{{DATE}}", t, -1)
	return r, nil
}

func (maker *SimpleEmailMaker) make(data *Data, c *viper.Viper) (*email.Email, error) {
	r := email.NewEmail()
	var err error
	r.Subject, err = getStringNonNil(c, "mail."+data.msgType+".subject")
	if err != nil {
		return nil, err
	}
	text, err := maker.getText(data, c)
	if err != nil {
		return nil, err
	}
	r.Text = []byte(text)
	r.To = []string{data.email}
	r.From, err = getStringNonNil(c, "smtp.username")
	return r, err
}

func getStringNonNil(c *viper.Viper, key string) (string, error) {
	r := c.GetString(key)
	if r == "" {
		return "", errors.New("No setting " + key)
	}
	return r, nil
}
