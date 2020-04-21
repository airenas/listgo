package strategy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	c, err := newCost(time.Second, 2, 3)
	assert.Nil(t, err)
	assert.NotNil(t, c)
}

func TestInit_Fails(t *testing.T) {
	_, err := newCost(time.Second*0, 2, 3)
	assert.NotNil(t, err)
	_, err = newCost(time.Second, 0, 3)
	assert.NotNil(t, err)
	_, err = newCost(time.Second, 500, 3)
	assert.NotNil(t, err)
	_, err = newCost(time.Second, 2, 0)
	assert.NotNil(t, err)
}
