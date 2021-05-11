package loader

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLocalFileList(t *testing.T) {
	fl, err := NewLocalFileList("/data/")
	assert.Nil(t, err)
	assert.NotNil(t, fl)

	fl, err = NewLocalFileList("")
	assert.NotNil(t, err)
	assert.Nil(t, fl)
}
