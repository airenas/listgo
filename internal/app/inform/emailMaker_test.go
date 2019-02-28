package inform

import (
	"strings"
	"testing"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/inform"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestFailsInit(t *testing.T) {
	m, err := newSimpleEmailMaker(viper.New())
	assert.NotNil(t, err, "Error expected")
	assert.Nil(t, m)
}

func TestInit_OK(t *testing.T) {
	v := viper.New()
	v.Set("mail.url", "url")
	m, err := newSimpleEmailMaker(v)
	assert.Nil(t, err)
	assert.Equal(t, m.url, "url")
}

func TestEmail(t *testing.T) {
	v := viper.New()
	v.Set("mail.url", "url")
	v.Set("mail.x.subject", "subject")
	v.Set("mail.x.text", "text")
	m, _ := newSimpleEmailMaker(v)
	data := inform.Data{}
	data.Email = "email"
	data.ID = "id"
	data.MsgType = "x"
	data.MsgTime = time.Now()

	e, _ := m.Make(&data)
	assert.Equal(t, e.Subject, "subject")
	assert.Contains(t, e.To, "email")
	assert.Equal(t, string(e.Text), "text")

	prepare(v).Set("mail.x.subject", "")
	_, err := m.Make(&data)
	assert.NotNil(t, err, "Error expected")

	prepare(v).Set("mail.x.text", "")
	_, err = m.Make(&data)
	assert.NotNil(t, err, "Error expected")

	prepare(v).Set("mail.x.text", "{{ID}}")
	e, _ = m.Make(&data)
	assert.Equal(t, string(e.Text), "id")

	prepare(v).Set("mail.x.text", "{{URL}}")
	e, _ = m.Make(&data)
	assert.Equal(t, string(e.Text), "url")

	prepare(v).Set("mail.x.text", "{{DATE}}")
	e, _ = m.Make(&data)

	assert.True(t, strings.HasPrefix(string(e.Text), data.MsgTime.Format("2006-01-02 15:04:05")))
}

func prepare(v *viper.Viper) *viper.Viper {
	v.Set("mail.url", "url")
	v.Set("mail.x.subject", "subject")
	v.Set("mail.x.text", "text")
	return v
}
