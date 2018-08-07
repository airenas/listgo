package manager

import (
	"bitbucket.org/airenas/listgo/internal/pkg/mongo"
	"bitbucket.org/airenas/listgo/internal/pkg/msgsender"
	"bitbucket.org/airenas/listgo/internal/pkg/msgworker"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/spf13/cobra"
)

var appName = "LiST Manager Service"

var rootCmd = &cobra.Command{
	Use:   "managerService",
	Short: appName,
	Long:  `Transcription manager service leads audio transcription process`,
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

	msgSender := msgworker.MachineMessageSender{Server: msgServer}
	msgWorker := msgworker.MachineWorker{Server: msgServer}

	mongoSessionProvider, err := mongo.NewSessionProvider()
	if err != nil {
		panic(err)
	}
	defer mongoSessionProvider.Close()

	statusSaver, err := mongo.NewStatusSaver(mongoSessionProvider)
	if err != nil {
		panic(err)
	}
	err = StartWorkerService(&ServiceData{&msgSender, &msgWorker, *statusSaver})
	if err != nil {
		panic(err)
	}
}
