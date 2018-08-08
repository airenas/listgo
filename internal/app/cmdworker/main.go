package cmdworker

import (
	"errors"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/msgsender"
	"bitbucket.org/airenas/listgo/internal/pkg/msgworker"

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
	msgServer, err := msgsender.NewMachineryServer()
	if err != nil {
		panic(err)
	}

	msgWorker := msgworker.MachineWorker{Server: msgServer}
	err = validateConfig()
	if err != nil {
		panic(err)
	}

	err = StartWorkerService(&ServiceData{&msgWorker, cmdapp.Config.GetString("worker.taskName"),
		cmdapp.Config.GetString("worker.command"), cmdapp.Config.GetString("worker.workingDir")})
	if err != nil {
		panic(err)
	}
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
