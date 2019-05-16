package status

import (
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/mongo"
	"bitbucket.org/airenas/listgo/internal/pkg/rabbit"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
)

var appName = "LiST Status Provider Service"

var rootCmd = &cobra.Command{
	Use:   "statusProviderService",
	Short: appName,
	Long:  `HTTP server to provide transcription status`,
	Run:   run,
}

func init() {
	cmdapp.InitApplication(rootCmd)
	rootCmd.PersistentFlags().Int32P("port", "", 8000, "Default service port")
	cmdapp.Config.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	cmdapp.Config.SetDefault("port", 8080)
}

//Execute starts the server
func Execute() {
	cmdapp.Execute(rootCmd)
}

func run(cmd *cobra.Command, args []string) {
	cmdapp.Log.Info("Starting " + appName)
	mongoSessionProvider, err := mongo.NewSessionProvider()
	cmdapp.CheckOrPanic(err, "")
	defer mongoSessionProvider.Close()

	data := &ServiceData{}
	data.StatusProvider, err = mongo.NewStatusProvider(mongoSessionProvider)
	cmdapp.CheckOrPanic(err, "")

	msgChannelProvider, err := rabbit.NewChannelProvider()
	cmdapp.CheckOrPanic(err, "")
	defer msgChannelProvider.Close()

	data.EventChannelFunc = func() (<-chan amqp.Delivery, error) {
		return initEventChannel(msgChannelProvider)
	}

	data.Port = cmdapp.Config.GetInt("port")

	err = StartWebServer(data)
	cmdapp.CheckOrPanic(err, "")
}

func initEventChannel(provider *rabbit.ChannelProvider) (<-chan amqp.Delivery, error) {
	cmdapp.Log.Info("Init event channel")
	ch, err := provider.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "Can't open channel")
	}

	var q amqp.Queue
	provider.RunOnChannelWithRetry(func(*amqp.Channel) error {
		q, err = ch.QueueDeclare(
			"",    // name
			false, // durable
			true,  // delete when usused
			true,  // exclusive
			false, // no-wait
			nil,   // arguments
		)
		return err
	})
	if err != nil {
		return nil, errors.Wrap(err, "Can't init queue")
	}
	err = ch.QueueBind(q.Name, "", messages.TopicStatusChange, false, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Can't bing to topic queue")
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, errors.Wrap(err, "Can't open channel")
	}
	cmdapp.Log.Info("Channel opened succesfully")
	return msgs, nil
}
