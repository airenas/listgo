package mongo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHidePass_NoPassword(t *testing.T) {
	url := "mongodb://mongo:27017"
	assert.Equal(t, hidePass(url), "mongodb://mongo:27017")
}

func TestHidePassword_Hidden(t *testing.T) {
	url := "mongodb://l:olia@mongo:27017"
	assert.Equal(t, hidePass(url), "mongodb://l:----@mongo:27017")
}
