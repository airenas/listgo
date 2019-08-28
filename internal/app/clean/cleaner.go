package clean

import (
	"errors"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
)

type cleanerImpl struct {
	jobs []Cleaner
}

func newCleanerImpl(fileStorage string) (*cleanerImpl, error) {
	c := cleanerImpl{}
	c.jobs = make([]Cleaner, 0)
	lf, err := newLocalFile(fileStorage, "audio.in/{ID}.*")
	if err != nil {
		return nil, err
	}
	c.jobs = append(c.jobs, lf)
	lf, err = newLocalFile(fileStorage, "decoded/audio/segmented/{ID}")
	if err != nil {
		return nil, err
	}
	c.jobs = append(c.jobs, lf)
	return &c, nil
}

func (c *cleanerImpl) Clean(ID string) error {
	failed := 0
	for _, job := range c.jobs {
		err := job.Clean(ID)
		if err != nil {
			cmdapp.Log.Error(err)
			failed++
		}
	}
	if failed == len(c.jobs) {
		return errors.New("All delete tasks failed")
	}
	return nil
}
