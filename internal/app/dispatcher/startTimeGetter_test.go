package dispatcher

import (
	"testing"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"github.com/stretchr/testify/assert"
)

func TestGetStartTime(t *testing.T) {
	g := newTimeGetter()
	st, err := g.get(newTestTags(messages.NewTag(messages.TagTimestamp, "1000")), time.Now())
	assert.Nil(t, err)
	assert.Equal(t, time.Unix(1000, 0), st)
}

func TestGetStartTime_SeveralTags(t *testing.T) {
	g := newTimeGetter()
	st, err := g.get(newTestTags(messages.NewTag(messages.TagTimestamp, "1000"),
		messages.NewTag("olia", "olia")), time.Now())
	assert.Nil(t, err)
	assert.Equal(t, time.Unix(1000, 0), st)
}

func TestGetStartTime_Default(t *testing.T) {
	g := newTimeGetter()
	now := time.Now()
	st, err := g.get(newTestTags(), now)
	assert.Nil(t, err)
	assert.Equal(t, now, st)
}

func TestGetStartTime_OnErrorValue(t *testing.T) {
	g := newTimeGetter()
	now := time.Now()
	_, err := g.get(newTestTags(messages.NewTag(messages.TagTimestamp, "olia")), now)
	assert.NotNil(t, err)
}

func newTestTags(tags ...messages.Tag) []messages.Tag {
	return tags
}
