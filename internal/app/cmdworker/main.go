package cmdworker

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/rabbit"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var appName = "LiST Worker Service"

var rootCmd = &cobra.Command{
	Use:   "cmdWorkerService",
	Short: appName,
	Long:  `Worker service listens for the work event from the queue and invokes configured command to do the work`,
	Run:   run,
}

func init() {
	cmdapp.InitApplication(rootCmd)
}

//Execute starts the server
func Execute() {
	cmdapp.Execute(rootCmd)
}

func run(cmd *cobra.Command, args []string) {
	cmdapp.Log.Info("Starting " + appName)
	err := validateConfig()
	if err != nil {
		panic(err)
	}
	data := ServiceData{}

	msgChannelProvider, err := rabbit.NewChannelProvider()
	if err != nil {
		panic(err)
	}
	defer msgChannelProvider.Close()

	data.MessageSender = rabbit.NewSender(msgChannelProvider)

	ch, err := msgChannelProvider.Channel()
	if err != nil {
		panic(errors.Wrap(err, "Can't open channel"))
	}
	err = ch.Qos(1, 0, false)
	if err != nil {
		panic(errors.Wrap(err, "Can't set Qos"))
	}

	data.TaskName = cmdapp.Config.GetString("worker.taskName")

	data.WorkCh, err = rabbit.NewChannel(ch, msgChannelProvider.QueueName(data.TaskName))
	if err != nil {
		panic(errors.Wrap(err, "Can't listen "+data.TaskName+" channel"))
	}

	data.Command = cmdapp.Config.GetString("worker.command")
	data.WorkingDir = cmdapp.Config.GetString("worker.workingDir")
	data.ResultFile = cmdapp.Config.GetString("worker.resultFile")
	data.ReadFunc = ReadFile

	fc, err := StartWorkerService(&data)
	if err != nil {
		panic(err)
	}
	<-fc
	cmdapp.Log.Infof("Exiting service")
}

func validateConfig() error {
	if cmdapp.Config.GetString("worker.taskName") == "" {
		return errors.New("No worker.taskName configured")
	}
	if cmdapp.Config.GetString("worker.command") == "" {
		return errors.New("No worker.command configured")
	}
	return nil
}
