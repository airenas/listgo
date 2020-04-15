package dispatcher

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/rabbit"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var appName = "LiST Dispatcher Service"

var rootCmd = &cobra.Command{
	Use:   "dispatcherService",
	Short: appName,
	Long:  `Dispatcher service listens for the work event from the queue and dispatches work to other queues`,
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
	cmdapp.CheckOrPanic(err, "Configuration error")

	data := ServiceData{}
	data.fc = utils.NewMultiCloseChannel()
	data.wrkrs = newWorkers()

	msgChannelProvider, err := rabbit.NewChannelProvider()
	cmdapp.CheckOrPanic(err, "Can't init rabbit channel provider")
	defer msgChannelProvider.Close()

	data.MessageSender = rabbit.NewSender(msgChannelProvider)

	ch, err := msgChannelProvider.Channel()
	cmdapp.CheckOrPanic(err, "Can't open channel")
	err = ch.Qos(1, 0, false)
	cmdapp.CheckOrPanic(err, "Can't set Qos")

	managerQueue := cmdapp.Config.GetString("worker.managerQueue")

	data.ManagerCh, err = rabbit.NewChannel(ch, managerQueue)
	cmdapp.CheckOrPanic(err, "Can't listen "+managerQueue+" channel")

	err = StartWorkerService(&data)
	cmdapp.CheckOrPanic(err, "Can't start service")

	<-data.fc.C
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
