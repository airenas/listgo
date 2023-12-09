package cmdworker

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
)

// RunCommand executes system comman end return error if any
func RunCommand(command string, workingDir string, id string, envs []string, outWriter io.Writer) error {
	logger := log.New(outWriter, "cmd: ", log.LstdFlags)
	realCommand := strings.Replace(command, "{ID}", id, -1)
	cmdapp.Log.Infof("Running command: %s", realCommand)
	logger.Printf("===== Running command: %s", realCommand)
	cmdapp.Log.Debugf("Working Dir: %s", workingDir)
	cmdArr := strings.Split(realCommand, " ")
	if len(cmdArr) < 2 {
		return errors.New("Wrong command. No parameter " + realCommand)
	}

	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	cmd.Dir = workingDir
	cmd.Env = os.Environ()
	for _, env := range envs {
		cmdapp.Log.Debug("Append env: " + env)
		cmd.Env = append(cmd.Env, env)
	}

	var outputBuffer bytes.Buffer
	outCopyWriter := io.MultiWriter(&outputBuffer, outWriter)
	cmd.Stdout = outCopyWriter
	cmd.Stderr = outCopyWriter

	err := cmd.Run()
	if err != nil {
		logger.Printf("===== ERROR ============")
		return errors.Wrap(err, "Output: "+string(outputBuffer.Bytes()))
	}
	logger.Printf("===== Finished ============")
	return nil
}
