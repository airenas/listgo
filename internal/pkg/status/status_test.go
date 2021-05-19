package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMin(t *testing.T) {
	assert.Equal(t, AudioConvert, Min(AudioConvert, AudioConvert))
	assert.Equal(t, AudioConvert, Min(AudioConvert, Diarization))
	assert.Equal(t, AudioConvert, Min(AudioConvert, Transcription))
	assert.Equal(t, Transcription, Min(Completed, Transcription))
}

func TestFrom(t *testing.T) {
	assert.Equal(t, JoinResults, From("JoinResults"))
	assert.Equal(t, AudioConvert, From("AudioConvert"))
}

func TestName(t *testing.T) {
	assert.Equal(t, "JoinResults", Name(JoinResults))
}
