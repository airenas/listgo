package progress_test

import (
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/progress"
	"bitbucket.org/airenas/listgo/internal/pkg/status"
	"github.com/stretchr/testify/assert"
)

func TestConvert(t *testing.T) {
	pr := progress.Convert(status.AudioConvert)
	assert.True(t, pr > 0)

	pr = progress.Convert(status.From("olia"))
	assert.Equal(t, int32(0), pr)

	pr = progress.Convert(status.Completed)
	assert.Equal(t, int32(100), pr)
}

func TestConvert_Rescore(t *testing.T) {
	pr := progress.Convert(status.Rescore)
	assert.Equal(t, int32(70), pr)
}

func TestConvert_ResultMake(t *testing.T) {
	pr := progress.Convert(status.ResultMake)
	assert.Equal(t, int32(90), pr)
}
