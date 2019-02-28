package progress_test

import (
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/progress"
	"bitbucket.org/airenas/listgo/internal/pkg/status"
	"github.com/stretchr/testify/assert"
)

func TestConvert(t *testing.T) {
	pr := progress.Convert(status.AudioConvert.Name)
	assert.True(t, pr > 0)

	pr = progress.Convert("olia")
	assert.Equal(t, pr, int32(0))

	pr = progress.Convert(status.Completed.Name)
	assert.Equal(t, pr, int32(100))
}
