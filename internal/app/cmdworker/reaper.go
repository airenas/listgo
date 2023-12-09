package cmdworker

import (
	"sync"

	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/hashicorp/go-reap"
)

func reapChildren(reapLock *sync.RWMutex) {
	cmdapp.Log.Debug("Init children reaper")
	pids := make(reap.PidCh, 1)
	go reap.ReapChildren(pids, nil, nil, reapLock)
	go debugReap(pids)
}

func debugReap(pids reap.PidCh) {
	for {
		pid := <-pids
		cmdapp.Log.Debugf("Reaped child process: %d", pid)
	}
}
