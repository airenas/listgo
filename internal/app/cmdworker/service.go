package cmdworker

import (
	"os/exec"
	"strings"

	"bitbucket.org/airenas/listgo/internal/app/manager"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
)

// ServiceData keeps data required for service work
type ServiceData struct {
	MessageListener manager.MessageListener
	TaskName        string
	Command         string
	WorkingDir      string
}

//StartWorkerService starts the event queue listener service tp listen for Decode events
func StartWorkerService(data *ServiceData) error {
	cmdapp.Log.Infof("Starting listen for messages")
	if data.TaskName == "" {
		return errors.New("No Task Name")
	}
	if data.Command == "" {
		return errors.New("No command")
	}

	err := data.MessageListener.RegisterTask(data.TaskName, func(id string) error {
		return work(data, id)
	})
	if err != nil {
		return errors.Wrap(err, "Can't register task")
	}

	err = data.MessageListener.Listen(data.TaskName + "_worker")

	if err != nil {
		return errors.Wrap(err, "Can't listen for message queue")
	}
	return nil
}

//work is main method to process of the worker
func work(data *ServiceData, id string) error {
	cmdapp.Log.Infof("Got task %s for ID: %s", data.TaskName, id)
	err := RunCommand(data.Command, data.WorkingDir, id)
	if err != nil {
		cmdapp.Log.Error(err)
		return err
	}
	return nil
}

//RunCommand executes system comman end return error if any
func RunCommand(command string, workingDir string, id string) error {
	realCommand := strings.Replace(command, "{ID}", id, -1)
	cmdapp.Log.Infof("Running command: %s", realCommand)
	cmdapp.Log.Infof("Working Dir: %s", workingDir)
	cmdArr := strings.Split(realCommand, " ")
	if len(cmdArr) < 2 {
		return errors.New("Wrong command. No parameter " + realCommand)
	}

	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	cmd.Dir = workingDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		errR := errors.Wrap(err, "Output: "+string(output))
		cmdapp.Log.Error(errR.Error())
		return errR
	}
	return nil
}
