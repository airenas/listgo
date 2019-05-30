package err

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var ece CodeExtractor

func TestDefault(t *testing.T) {
	assert.Equal(t, DefaultCode, ece.Get(""))
	assert.Equal(t, DefaultCode, ece.Get("error"))
	assert.Equal(t, DefaultCode, ece.Get("[[ErrorCode:"))
	assert.Equal(t, DefaultCode, ece.Get(errorCodeStart+"olia"))
	assert.Equal(t, DefaultCode, ece.Get("olia"+errorCodeEnd))
	assert.Equal(t, DefaultCode, ece.Get(errorCodeStart+""+errorCodeEnd))
}

func TestExtract(t *testing.T) {
	assert.Equal(t, "olia", ece.Get(errorCodeStart+"olia"+errorCodeEnd))
	assert.Equal(t, "olia", ece.Get("error\n\n"+errorCodeStart+"olia"+errorCodeEnd))
	assert.Equal(t, "olia", ece.Get("error\n\n"+errorCodeStart+"olia"+errorCodeEnd+"\naaaa"))
}

func TestTrims(t *testing.T) {
	assert.Equal(t, "errorCode", ece.Get(errorCodeStart+"  errorCode \n\t"+errorCodeEnd))
}
