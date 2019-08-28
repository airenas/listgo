package clean

import (
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
)

//OldIDsProvider return old ids for clesning service
type OldIDsProvider interface {
	Get() ([]string, error)
}

type timerServiceData struct {
	runEvery     time.Duration
	cleaner      Cleaner
	idsProvider  OldIDsProvider
	qChan        chan struct{}
	workWaitChan chan struct{}
}

func startCleanTimer(data *timerServiceData) error {
	cmdapp.Log.Infof("Starting timer service every %v", data.runEvery)
	go serviceLoop(data)
	return nil
}

func serviceLoop(data *timerServiceData) {
	ticker := time.NewTicker(data.runEvery)
	// run on startup
	doClean(data)
mainloop:
	for {
		select {
		case <-ticker.C:
			doClean(data)
		case <-data.qChan:
			ticker.Stop()
			break mainloop
		}
	}
	cmdapp.Log.Infof("Stopped timer service")
	close(data.workWaitChan)
}

func doClean(data *timerServiceData) {
	cmdapp.Log.Info("Running cleaning")
	ids, err := data.idsProvider.Get()
	if err != nil {
		cmdapp.Log.Error(err)
	}
	cmdapp.Log.Infof("Got %d IDs to clean", len(ids))
	for _, id := range ids {
		err = data.cleaner.Clean(id)
		if err != nil {
			cmdapp.Log.Error(err)
		}
	}
}
