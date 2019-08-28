package clean

import (
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
)

type oldIdsProvider interface {
	Get() ([]string, error)
}

type timerServiceData struct {
	runEvery     time.Duration
	cleaner      Cleaner
	idsProvider  oldIdsProvider
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
	cmdapp.Log.Info("Info running cleaning")
	ids, err := data.idsProvider.Get()
	if err != nil {
		cmdapp.Log.Error(err)
	}
	for _, id := range ids {
		err = data.cleaner.Clean(id)
		if err != nil {
			cmdapp.Log.Error(err)
		}
	}
}
