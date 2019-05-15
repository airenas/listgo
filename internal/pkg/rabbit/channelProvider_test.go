package rabbit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyQueueName(t *testing.T) {
	var prv ChannelProvider
	assert.Equal(t, "", prv.QueueName(""))
}

func TestNoPrefix(t *testing.T) {
	var prv ChannelProvider
	assert.Equal(t, "olia", prv.QueueName("olia"))
}

func TestPrefix(t *testing.T) {
	var prv ChannelProvider
	prv.qPrefix = "prefix"
	assert.Equal(t, "prefix_olia", prv.QueueName("olia"))
}
