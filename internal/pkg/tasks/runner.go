package tasks

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
)

// Runner executes external process and manages it
type Runner struct {
	errWriter  io.Writer
	outWriter  io.Writer
	workingDir string

	lock *sync.Mutex

	cmd        *exec.Cmd
	startWait  *sync.WaitGroup
	startErr   error // start error
	exitErr    error // exit error
	finishChan chan struct{}
}

// NewRunner inits new runner instance
func NewRunner(workingDir string) (*Runner, error) {
	r := &Runner{}
	r.workingDir = workingDir
	r.lock = &sync.Mutex{}
	r.startWait = &sync.WaitGroup{}
	r.startWait.Add(1)
	r.finishChan = make(chan struct{})
	return r, nil
}

// Close terminates runnig preocess if any
func (r *Runner) Close() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.cmd != nil && r.Running() && r.cmd.Process != nil {
		cmdapp.Log.Infof("Closing pid: %d", r.cmd.Process.Pid)
		_ = syscall.Kill(-r.cmd.Process.Pid, syscall.SIGTERM) // send terminate to group with -pid
		cmdapp.Log.Infof("Sent term signal to pid: %d", r.cmd.Process.Pid)
		err := r.waitForFinish()
		if err == nil {
			return nil
		}
		cmdapp.Log.Infof("Not responded. Killing pid: %d", r.cmd.Process.Pid)
		_ = syscall.Kill(-r.cmd.Process.Pid, syscall.SIGKILL) // send terminate to group with -pid
		cmdapp.Log.Infof("Sent kill signal to pid: %d", r.cmd.Process.Pid)
		return r.waitForFinish()
	}
	return nil
}

func (r *Runner) waitForFinish() error {
	select {
	case <-r.finishChan:
		cmdapp.Log.Info("Process exited")
		return nil
	case <-time.After(10 * time.Second):
		return errors.Errorf("timeout waiting to finish")
	}
}

// Run starts the preocess
func (r *Runner) Run(cmdStr string, envs []string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	cmdapp.Log.Infof("Running command: %s", cmdStr)
	cmdapp.Log.Infof("Working Dir: %s", r.workingDir)
	cmdArr := stringToArgs(cmdStr)
	if len(cmdArr) < 1 {
		return errors.Errorf("Wrong or no command '%s'.", cmdStr)
	}

	r.cmd = exec.Command(cmdArr[0], cmdArr[1:]...)
	r.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	r.cmd.Dir = r.workingDir
	r.cmd.Env = os.Environ()
	for _, env := range envs {
		cmdapp.Log.Debug("Append env: " + env)
		r.cmd.Env = append(r.cmd.Env, env)
	}
	r.cmd.Stdout = r.outWriter
	r.cmd.Stderr = r.errWriter

	go func() {
		r.runCmd(r.cmd)
	}()
	r.startWait.Wait()
	return r.startErr
}

// Running returns the running status
func (r *Runner) Running() bool {
	r.startWait.Wait()
	select {
	case <-r.finishChan:
		return false
	default:
		return true
	}
}

func (r *Runner) runCmd(cmd *exec.Cmd) {
	r.exitErr = nil
	r.startErr = nil
	r.cmd = cmd
	err := cmd.Start()
	if err != nil {
		cmdapp.Log.Error(err)
		r.startErr = err
		close(r.finishChan)
		r.startWait.Done()
		return
	}
	r.startWait.Done()

	err = cmd.Wait()

	close(r.finishChan)
	if err != nil {
		cmdapp.Log.Error(err)
		r.exitErr = err
	}
}

func stringToArgs(str string) []string {
	res := make([]string, 0)
	w := ""
	for len(str) > 0 {
		w, str = getWord(str)
		if w != "" {
			res = append(res, w)
		}
	}
	return res
}

func getWord(str string) (string, string) {
	str = strings.TrimSpace(str)
	if str == "" {
		return "", ""
	}
	c := ' '
	if str[0] == '"' || str[0] == '\'' {
		c = rune(str[0])
		str = str[1:]
	}
	pos := -1
	for i, s := range str {
		if s == c && !(i > 0 && str[i-1] == '\\') {
			pos = i
			break
		}
	}
	if pos == -1 {
		return str, ""
	}
	w := str[:pos]
	str = strings.TrimSpace(str[pos+1:])
	return w, str
}
