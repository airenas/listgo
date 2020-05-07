package clean

import (
	"strings"
	"syscall"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/mongo"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type cleanerImpl struct {
	jobs        []Cleaner
	fileStorage string

	counter prometheus.Counter
}

func newCleanerImpl(mng *mongo.SessionProvider, fileStorage string, patterns string, counter prometheus.Counter) (*cleanerImpl, error) {
	c := cleanerImpl{}
	c.jobs = make([]Cleaner, 0)
	c.fileStorage = fileStorage

	fcs, err := newFileCleaners(fileStorage, patterns)
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
	if counter == nil {
		return nil, errors.New("No metrics counter")
	}
	c.counter = counter
	return &c, nil
}

func (c *cleanerImpl) Clean(ID string) error {
	c.counter.Inc()
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

func newFileCleaners(fs string, patterns string) ([]*localFile, error) {
	ps := strings.Split(patterns, "\n")
	result := make([]*localFile, 0)
	for _, p := range ps {
		p = strings.TrimSpace(p)
		if p != "" {
			fc, err := newLocalFile(fs, p)
			if err != nil {
				return nil, err
			}
			result = append(result, fc)
		}
	}
	return result, nil
}

//HealthyFunc returns func for health check
func (c *cleanerImpl) HealthyFunc() func() error {
	return func() error {
		var info syscall.Statfs_t
		err := syscall.Statfs(c.fileStorage, &info)
		if err != nil {
			return errors.Errorf("Can't get info for dir: %s", c.fileStorage)
		}
		return nil
	}
}
