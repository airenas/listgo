package dispatcher

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/rabbit"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
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

	registrationQueue := cmdapp.Config.GetString("messageServer.registrationQueue")

	err = initRegistrationQueue(msgChannelProvider, registrationQueue)
	data.RegistrationCh, err = rabbit.NewChannel(ch, registrationQueue)
	cmdapp.CheckOrPanic(err, "Can't listen "+registrationQueue+" channel")

	err = StartWorkerService(&data)
	cmdapp.CheckOrPanic(err, "Can't start service")

	<-data.fc.C
	cmdapp.Log.Infof("Exiting service")
}

func validateConfig() error {
	if cmdapp.Config.GetString("messageServer.registrationQueue") == "" {
		return errors.New("No messageServer.registrationQueue configured")
	}
	return nil
}

func initRegistrationQueue(prv *rabbit.ChannelProvider, qName string) error {
	return prv.RunOnChannelWithRetry(func(ch *amqp.Channel) error {
		_, err := rabbit.DeclareQueue(ch, prv.QueueName(qName))
		if err != nil {
			return err
		}
		return nil
	})
}
