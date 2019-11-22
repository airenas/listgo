package upload

import (
	"time"

	"github.com/streadway/amqp"

	"bitbucket.org/airenas/listgo/internal/pkg/config"
	"bitbucket.org/airenas/listgo/internal/pkg/messages"

	"bitbucket.org/airenas/listgo/internal/pkg/mongo"
	"bitbucket.org/airenas/listgo/internal/pkg/rabbit"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/saver"
	"github.com/spf13/cobra"

	"github.com/heptiolabs/healthcheck"
)

var rootCmd = &cobra.Command{
	Use:   "uploadService",
	Short: "LiST Upload Audio File Service",
	Long:  `HTTP server to listen and upload audio files for transcription`,
	Run:   run,
}

func init() {
	cmdapp.InitApplication(rootCmd)
	rootCmd.PersistentFlags().Int32P("port", "", 8000, "Default service port")
	cmdapp.Config.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	cmdapp.Config.SetDefault("port", 8080)
	cmdapp.Config.SetDefault("fileStorage.path", "/data/audio.in/")
}

//Execute starts the server
func Execute() {
	cmdapp.Execute(rootCmd)
}

func run(cmd *cobra.Command, args []string) {
	cmdapp.Log.Info("Starting uploadService")
	var data ServiceData
	var err error
	data.health = healthcheck.NewHandler()
	fs, err := saver.NewLocalFileSaver(cmdapp.Config.GetString("fileStorage.path"))
	cmdapp.CheckOrPanic(err, "Can't init file storage")
	data.FileSaver = fs
	data.health.AddLivenessCheck("fs", fs.HealthyFunc(50))

	data.RecognizerMap, err = config.NewFileRecognizerMap(cmdapp.Config.GetString("recognizerConfig.path"))
	cmdapp.CheckOrPanic(err, "Can't init recognizer config")

	msgChannelProvider, err := rabbit.NewChannelProvider()
	cmdapp.CheckOrPanic(err, "Can't init rabbit channel")
	defer msgChannelProvider.Close()
	data.health.AddLivenessCheck("rabbit", healthcheck.Async(msgChannelProvider.Healthy, 10*time.Second))

	err = initQueues(msgChannelProvider)
	cmdapp.CheckOrPanic(err, "Can't init queues")

	data.MessageSender = rabbit.NewSender(msgChannelProvider)

	mongoSessionProvider, err := mongo.NewSessionProvider()
	cmdapp.CheckOrPanic(err, "Can't init mongo")
	defer mongoSessionProvider.Close()
	data.health.AddLivenessCheck("mongo", healthcheck.Async(mongoSessionProvider.Healthy, 10*time.Second))

	data.StatusSaver, err = mongo.NewStatusSaver(mongoSessionProvider)
	cmdapp.CheckOrPanic(err, "Can't init status saver")

	data.RequestSaver, err = mongo.NewRequestSaver(mongoSessionProvider)
	cmdapp.CheckOrPanic(err, "Can't init request saver")
	data.Port = cmdapp.Config.GetInt("port")

	err = StartWebServer(&data)
	cmdapp.CheckOrPanic(err, "Can't start web server")
}

func initQueues(prv *rabbit.ChannelProvider) error {
	cmdapp.Log.Info("Initializing queues")
	return prv.RunOnChannelWithRetry(func(ch *amqp.Channel) error {
		_, err := rabbit.DeclareQueue(ch, prv.QueueName(messages.Decode))
		return err
	})
}
