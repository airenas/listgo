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

func TestFailsInit_ID(t *testing.T) {
	f, err := newLocalFile("/olia", "path")
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

func TestInitWithoutPath(t *testing.T) {
	f, err := newLocalFile("", "/path1{ID}")
	assert.Nil(t, err)
	assert.NotNil(t, f)
	assert.Equal(t, "", f.StoragePath)
	assert.Equal(t, "/path1{ID}", f.pattern)
}

func TestGetPath_NoStoragePath(t *testing.T) {
	f, _ := newLocalFile("", "/path{ID}")
	assert.NotNil(t, f)
	assert.Equal(t, "/path222", f.getPath("222"))
}

func TestGetPath_WithStoragePath(t *testing.T) {
	f, _ := newLocalFile("/haha", "path{ID}")
	assert.NotNil(t, f)
	assert.Equal(t, "/haha/path222", f.getPath("222"))
}
