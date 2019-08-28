package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFailsInit_StoragePath(t *testing.T) {
	f, err := newLocalFile("", "path")
	assert.Nil(t, f)
	assert.NotNil(t, err)
}

func TestFailsInit_Patern(t *testing.T) {
	f, err := newLocalFile("/path", "")
	assert.Nil(t, f)
	assert.NotNil(t, err)
	f, err = newLocalFile("/path", "olia")
	assert.Nil(t, f)
	assert.NotNil(t, err)
}

func TestInit(t *testing.T) {
	f, err := newLocalFile("/path", "olia/{ID}")
	assert.Nil(t, err)
	assert.NotNil(t, f)
}
