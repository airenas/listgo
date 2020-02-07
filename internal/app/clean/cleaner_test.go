package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFileCleaners(t *testing.T) {
	f, err := newFileCleaners("/path", "path1{ID}")
	assert.Nil(t, err)
	assert.NotNil(t, f)
}

func TestSeveralFileCleaners(t *testing.T) {
	f, err := newFileCleaners("/path", "path1{ID}\nPath2{ID}\n  \n")
	assert.Nil(t, err)
	assert.NotNil(t, f)
	assert.Equal(t, 2, len(f))
}

func TestNewFileCleanersPath(t *testing.T) {
	f, err := newFileCleaners("/path", "path1{ID}")
	assert.Nil(t, err)
	assert.NotNil(t, f)
	assert.Equal(t, 1, len(f))
	assert.Equal(t, "/path", f[0].StoragePath)
	assert.Equal(t, "path1{ID}", f[0].pattern)
}
