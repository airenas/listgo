package inform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetType(t *testing.T) {
	testGetType(t, "", SMTP_PLAIN, false)
	testGetType(t, "olia", "", true)
	testGetType(t, SMTP_PLAIN, SMTP_PLAIN, false)
	testGetType(t, SMTP_LOGIN, SMTP_LOGIN, false)
	testGetType(t, SMTP_NOAUTH, SMTP_NOAUTH, false)
	testGetType(t, "no_auth", SMTP_NOAUTH, false)
}

func TestFullHost(t *testing.T) {
	se := SimpleEmailSender{}
	se.host = "net.olia"
	se.port = 445
	assert.Equal(t, "net.olia:445", se.getFullHost())
}

func testGetType(t *testing.T, v, exp string, expErr bool) {
	m, err := getType(v)
	assert.Equal(t, expErr, err != nil)
	assert.Equal(t, exp, m)
}
