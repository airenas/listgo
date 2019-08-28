package clean

import (
	"errors"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/mongo"
)

type cleanerImpl struct {
	jobs []Cleaner
}

func newCleanerImpl(mng *mongo.SessionProvider, fileStorage string) (*cleanerImpl, error) {
	c := cleanerImpl{}
	c.jobs = make([]Cleaner, 0)

	fcs, err := newFileCleaners(fileStorage,
		"audio.in/{ID}.*",
		"audio.prepared/{ID}.*",
		"decoded/audio/segmented/{ID}",
		"decoded/diarization/{ID}",
		"decoded/trans/{ID}",
		"results/{ID}")
	if err != nil {
		return nil, err
	}
	for _, fc := range fcs {
		c.jobs = append(c.jobs, fc)
	}

	mcs, err := mongo.NewCleanRecords(mng)
	if err != nil {
		return nil, err
	}
	for _, mc := range mcs {
		c.jobs = append(c.jobs, mc)
	}
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

func newFileCleaners(fs string, patterns ...string) ([]*localFile, error) {
	result := make([]*localFile, 0)
	for _, p := range patterns {
		fc, err := newLocalFile(fs, p)
		if err != nil {
			return nil, err
		}
		result = append(result, fc)
	}
	return result, nil
}
