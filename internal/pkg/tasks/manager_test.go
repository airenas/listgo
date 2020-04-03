package tasks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFailsOnPrefix(t *testing.T) {
	m, err := NewManager("")
	assert.NotNil(t, err)
	assert.Nil(t, m)
}

func TestInit_OK(t *testing.T) {
	m, err := NewManager("pr")
	assert.Nil(t, err)
	assert.NotNil(t, m)
}