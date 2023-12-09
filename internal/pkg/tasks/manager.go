package tasks

import (
	"fmt"
	"strings"
	"sync"

	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
)

// ProcessRunner executes external process and manages it
type ProcessRunner interface {
	Run(cmd string, env []string) error
	Close() error
	Running() bool
}

// Manager manages tasks by key
type Manager struct {
	keyPrefix string

	lock       *sync.Mutex
	currentKey string
	runner     ProcessRunner
}

// NewManager creates LocalFileSaver instance
func NewManager(prefix string, workingDir string) (*Manager, error) {
	r, err := NewRunner(workingDir)
	if err != nil {
		return nil, err
	}
	logWriter := cmdapp.Log.Writer()
	r.outWriter = logWriter
	r.errWriter = logWriter
	return newManager(prefix, r)
}

// NewManager creates LocalFileSaver instance
func newManager(prefix string, r ProcessRunner) (*Manager, error) {
	if prefix == "" {
		return nil, errors.New("No prefix for task manager provided")
	}
	if r == nil {
		return nil, errors.New("No runner for task manager provided")
	}
	m := Manager{keyPrefix: prefix, lock: &sync.Mutex{}, runner: r}
	return &m, nil
}

// EnsureRunning loads data by key and ensure it is running
func (m *Manager) EnsureRunning(in map[string]string) error {
	key, f := in[m.keyPrefix+"_key"]
	if !f {
		return errors.Errorf("No preload key '%s' found for the task", m.keyPrefix+"_key")
	}
	m.lock.Lock()
	defer m.lock.Unlock()

	if key == "" {
		err := m.runner.Close()
		if err != nil {
			cmdapp.Log.Warn(err)
		}
		m.currentKey = ""
		return nil
	}

	if key != m.currentKey {
		if m.currentKey != "" {
			cmdapp.Log.Infof("Closing preloader task for key %s...", m.currentKey)
		}
		err := m.runner.Close()
		if err != nil {
			return errors.Wrap(err, "Can't close the running task")
		}
		return m.start(key, in)
	}

	if !m.runner.Running() {
		cmdapp.Log.Info("Preload task is not running. Trying to start...")
		return m.start(key, in)
	}
	cmdapp.Log.Info("Preload task is running.")
	return nil
}

func (m *Manager) start(key string, in map[string]string) error {
	cmd, f := in[m.keyPrefix+"_cmd"]
	if !f || cmd == "" {
		return errors.Errorf("No preload command '%s' found for the task", m.keyPrefix+"_cmd")
	}
	cmdapp.Log.Info("Starting preload task " + cmd)
	err := m.runner.Run(cmd, makeEnv(in))
	if err != nil {
		return errors.Wrap(err, "Can't start the preload task")
	}
	m.currentKey = key
	if !m.runner.Running() {
		return errors.New("Preload task is not running or terminated")
	}
	return nil
}

// Close terminates runnig preocess if any
func (m *Manager) Close() error {
	return m.runner.Close()
}

func makeEnv(in map[string]string) []string {
	var res []string
	for k, v := range in {
		res = append(res, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
	}
	return res
}
