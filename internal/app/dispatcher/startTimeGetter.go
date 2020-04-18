package dispatcher

import (
	"strconv"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"github.com/pkg/errors"
)

type timeGetter struct {
}

func newTimeGetter() *timeGetter {
	return &timeGetter{}
}

func (g *timeGetter) Get(tags []messages.Tag) (time.Time, error) {
	return g.get(tags, time.Now())
}

func (g *timeGetter) get(tags []messages.Tag, def time.Time) (time.Time, error) {
	for _, t := range tags {
		if t.Key == messages.TagTimestamp {
			return toTime(t.Value, def)
		}
	}
	return def, nil
}

func toTime(s string, def time.Time) (time.Time, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return def, errors.Wrapf(err, "Can't parse %s", s)
	}
	return time.Unix(int64(i), 0), err
}
