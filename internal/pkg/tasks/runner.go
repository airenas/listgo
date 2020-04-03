package tasks

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
)

//Runner executes external process and manages it
type Runner struct {
	errWriter  io.Writer
	outWriter  io.Writer
	workingDir string

	lock *sync.Mutex

	cmd        *exec.Cmd
	startWait  *sync.WaitGroup
	running    bool  // status
	processErr error // exit error
}

//NewRunner inits new runner instance
func NewRunner(workingDir string) (*Runner, error) {
	r := &Runner{}
	r.workingDir = workingDir
	r.lock = &sync.Mutex{}
	r.startWait = &sync.WaitGroup{}
	return r, nil
}

// Close terminates runnig preocess if any
func (r *Runner) Close() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.cmd != nil && r.running && r.cmd.Process != nil {
		cmdapp.Log.Infof("Closing pid: %d", r.cmd.Process.Pid)
		return r.cmd.Process.Signal(syscall.SIGTERM)
	}
	return nil
}

// Run starts the preocess
func (r *Runner) Run(cmdStr string, envs []string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	cmdapp.Log.Infof("Running command: %s", cmdStr)
	cmdapp.Log.Infof("Working Dir: %s", r.workingDir)
	cmdArr := strings.Split(cmdStr, " ")
	if len(cmdArr) < 1 {
		return errors.Errorf("Wrong or no command '%s'.", cmdStr)
	}

	r.cmd = exec.Command(cmdArr[0], cmdArr[1:]...)
	r.cmd.Dir = r.workingDir
	r.cmd.Env = os.Environ()
	for _, env := range envs {
		cmdapp.Log.Debug("Append env: " + env)
		r.cmd.Env = append(r.cmd.Env, env)
	}
	r.cmd.Stdout = r.outWriter
	r.cmd.Stderr = r.errWriter

	r.startWait.Add(1)
	go func() {
		r.runCmd(r.cmd)
	}()
	r.startWait.Wait()
	return r.processErr
}

// Running returns the running status
func (r *Runner) Running() bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.startWait.Wait()
	return r.running
}

func (r *Runner) runCmd(cmd *exec.Cmd) {
	r.running = false
	r.processErr = nil
	r.cmd = cmd
	err := cmd.Start()
	if err != nil {
		r.running = false
		cmdapp.Log.Error(err)
		r.processErr = err
		r.startWait.Done()
		return
	}
	r.running = true
	r.startWait.Done()

	err = cmd.Wait()
	r.running = false
	
	if err != nil {
		cmdapp.Log.Error(err)
		r.processErr = err
	}
}
