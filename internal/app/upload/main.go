package upload

import (
	"os"

	"github.com/streadway/amqp"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"

	"bitbucket.org/airenas/listgo/internal/pkg/mongo"
	"bitbucket.org/airenas/listgo/internal/pkg/rabbit"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/saver"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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

func logPanic() {
	if r := recover(); r != nil {
		cmdapp.Log.Error(r)
		os.Exit(1)
	}
}

//Execute starts the server
func Execute() {
	defer logPanic()
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func run(cmd *cobra.Command, args []string) {
	cmdapp.Log.Info("Starting uploadService")
	fileSaver, err := saver.NewLocalFileSaver(cmdapp.Config.GetString("fileStorage.path"))
	if err != nil {
		panic(err)
	}
	msgChannelProvider, err := rabbit.NewChannelProvider(cmdapp.Config.GetString("messageServer.broker"))
	if err != nil {
		panic(err)
	}
	defer msgChannelProvider.Close()

	err = initQueues(msgChannelProvider)
	if err != nil {
		panic(errors.Wrap(err, "Can't init queues"))
	}
	msgSender := rabbit.NewSender(msgChannelProvider)

	mongoSessionProvider, err := mongo.NewSessionProvider()
	if err != nil {
		panic(err)
	}
	defer mongoSessionProvider.Close()

	statusSaver, err := mongo.NewStatusSaver(mongoSessionProvider)
	if err != nil {
		panic(err)
	}
	err = StartWebServer(&ServiceData{fileSaver, msgSender, statusSaver, cmdapp.Config.GetInt("port")})
	if err != nil {
		panic(err)
	}
}

func initQueues(prv *rabbit.ChannelProvider) error {
	cmdapp.Log.Info("Initializing queues")
	return prv.RunOnChannelWithRetry(func(ch *amqp.Channel) error {
		_, err := rabbit.DeclareQueue(ch, messages.Decode)
		return err
	})
}
