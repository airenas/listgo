package inform

import (
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/mongo"
	"bitbucket.org/airenas/listgo/internal/pkg/rabbit"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var appName = "LiST email information Service"

var rootCmd = &cobra.Command{
	Use:   "informService",
	Short: appName,
	Long:  `Service listens for the information events from the queue and informs user`,
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
	cmdapp.CheckOrPanic(err, "")

	data := ServiceData{}

	msgChannelProvider, err := rabbit.NewChannelProvider()
	cmdapp.CheckOrPanic(err, "Can't init rabbit channel")
	defer msgChannelProvider.Close()

	ch, err := msgChannelProvider.Channel()
	cmdapp.CheckOrPanic(err, "Can't open channel")

	err = ch.Qos(1, 0, false)
	cmdapp.CheckOrPanic(err, "Can't set Qos")

	data.TaskName = cmdapp.Config.GetString("worker.taskName")

	data.WorkCh, err = rabbit.NewChannel(ch, data.TaskName)
	cmdapp.CheckOrPanic(err, "Can't listen to "+data.TaskName+" channel")

	data.emailMaker, err = newSimpleEmailMaker(cmdapp.Config)
	cmdapp.CheckOrPanic(err, "Can't init email maker")

	location := cmdapp.Config.GetString("worker.location")
	if location != "" {
		data.location, err = time.LoadLocation(location)
		cmdapp.CheckOrPanic(err, "Can't init location")
	}

	data.emailSender, err = newSimpleEmailSender()
	cmdapp.CheckOrPanic(err, "Can't init email sender")

	mongoSessionProvider, err := mongo.NewSessionProvider()
	cmdapp.CheckOrPanic(err, "Can't init mongo provider")
	defer mongoSessionProvider.Close()

	data.locker, err = mongo.NewLocker(mongoSessionProvider)
	cmdapp.CheckOrPanic(err, "Can't init mongo locker")

	data.emailRetriever, err = mongo.NewEmailRetriever(mongoSessionProvider)
	cmdapp.CheckOrPanic(err, "Can't init mongo email retriever")

	fc, err := StartWorkerService(&data)
	cmdapp.CheckOrPanic(err, "Can't start service")

	<-fc
	cmdapp.Log.Infof("Exiting service")
}

func validateConfig() error {
	taskName := cmdapp.Config.GetString("worker.taskName")
	if taskName == "" {
		return errors.New("No worker.taskName configured")
	}
	return nil
}
