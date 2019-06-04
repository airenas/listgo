package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURLJoin(t *testing.T) {
	assert.Equal(t, "http://www.delfi.lt/olia", URLJoin("http://www.delfi.lt", "olia"))
	assert.Equal(t, "http://www.delfi.lt/olia/1", URLJoin("http://www.delfi.lt", "olia", "1"))
	assert.Equal(t, "http://www.delfi.lt/olia/1", URLJoin("http://www.delfi.lt/", "/olia/", "1"))
	assert.Equal(t, "http://www.delfi.lt/olia/1", URLJoin("http://www.delfi.lt", "olia", "/1"))
	assert.Equal(t, "http://www.delfi.lt", URLJoin("http://www.delfi.lt"))
	assert.Equal(t, "http://www.delfi.lt:80/olia", URLJoin("http://www.delfi.lt:80/", "olia"))
	assert.Equal(t, "www.delfi.lt:80/olia", URLJoin("www.delfi.lt:80", "olia"))
}

func TestValidateURL(t *testing.T) {
	ut, err := validateConfigURL("http://www.delfi.lt/olia/1", "sn")
	assert.Equal(t, "http://www.delfi.lt/olia/1", ut)
	assert.Nil(t, err)
}

func TestValidateURL_FailEmpty(t *testing.T) {
	ut, err := validateConfigURL("", "sn")
	assert.Equal(t, "", ut)
	assert.NotNil(t, err)
}

func TestValidateURL_Fail(t *testing.T) {
	ut, err := validateConfigURL(":::://", "sn")
	assert.Equal(t, "", ut)
	assert.NotNil(t, err)
}
