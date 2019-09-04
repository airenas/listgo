package kafka

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CheckTrimLen_NoTrim(t *testing.T) {
	ts := "Olia olia"
	assert.Equal(t, ts, checkTrimLen(ts, 10))
}

func Test_CheckTrimLen_Trim(t *testing.T) {
	ts := "Olia olia"
	assert.Contains(t, checkTrimLen(ts, 5), "\nOlia ...")
	assert.Contains(t, checkTrimLen(ts, 1), "\nO...")
}

func Test_CheckTrimLen_UnicodeTrim(t *testing.T) {
	ts := "ĄČĘĖ olia"
	assert.Contains(t, checkTrimLen(ts, 5), "\nĄČĘĖ ...")
}

func Test_CheckTrimLen_Empty(t *testing.T) {
	assert.Contains(t, checkTrimLen("", 5), "")
}
