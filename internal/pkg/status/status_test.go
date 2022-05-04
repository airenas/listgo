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
	assert.Equal(t, SplitChannels, From("SplitChannels"))
}

func TestName(t *testing.T) {
	assert.Equal(t, "JoinResults", Name(JoinResults))
	assert.Equal(t, "SplitChannels", Name(SplitChannels))
}
