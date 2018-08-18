package manager

import (
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/mongo"
	"bitbucket.org/airenas/listgo/internal/pkg/rabbit"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
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
	data := ServiceData{}

	mongoSessionProvider, err := mongo.NewSessionProvider()
	if err != nil {
		panic(err)
	}
	defer mongoSessionProvider.Close()

	msgChannelProvider, err := rabbit.NewChannelProvider(cmdapp.Config.GetString("messageServer.broker"))
	if err != nil {
		panic(err)
	}
	defer msgChannelProvider.Close()

	err = initQueues(msgChannelProvider)
	if err != nil {
		panic(errors.Wrap(err, "Can't init queues"))
	}
	data.MessageSender = rabbit.NewSender(msgChannelProvider)

	ch, err := msgChannelProvider.Channel()
	if err != nil {
		panic(errors.Wrap(err, "Can't open channel"))
	}
	err = ch.Qos(1, 0, false)
	if err != nil {
		panic(errors.Wrap(err, "Can't set Qos"))
	}

	data.DecodeCh, err = rabbit.NewChannel(ch, messages.Decode)
	if err != nil {
		panic(errors.Wrap(err, "Can't listen Decode channel"))
	}
	data.AudioConvertCh, err = rabbit.NewChannel(ch, messages.ResultQueueFor(messages.AudioConvert))
	if err != nil {
		panic(errors.Wrap(err, "Can't listen AudioConvertCh channel"))
	}
	data.DiarizationCh, err = rabbit.NewChannel(ch, messages.ResultQueueFor(messages.Diarization))
	if err != nil {
		panic(errors.Wrap(err, "Can't listen DiarizationCh channel"))
	}
	data.TranscriptionCh, err = rabbit.NewChannel(ch, messages.ResultQueueFor(messages.Transcription))
	if err != nil {
		panic(errors.Wrap(err, "Can't listen DiarizationCh channel"))
	}
	data.ResultMakeCh, err = rabbit.NewChannel(ch, messages.ResultQueueFor(messages.ResultMake))
	if err != nil {
		panic(errors.Wrap(err, "Can't listen DiarizationCh channel"))
	}

	data.StatusSaver, err = mongo.NewStatusSaver(mongoSessionProvider)
	if err != nil {
		panic(err)
	}

	err = StartWorkerService(&data)
	if err != nil {
		panic(err)
	}
}

func initQueues(prv *rabbit.ChannelProvider) error {
	cmdapp.Log.Info("Initializing queues")
	return prv.RunOnChannelWithRetry(func(ch *amqp.Channel) error {
		queues := []string{messages.Decode, messages.StartedDecode,
			messages.AudioConvert, messages.ResultQueueFor(messages.AudioConvert),
			messages.Diarization, messages.ResultQueueFor(messages.Diarization),
			messages.Transcription, messages.ResultQueueFor(messages.Transcription),
			messages.ResultMake, messages.ResultQueueFor(messages.ResultMake),
			messages.FinishDecode}
		for _, queue := range queues {
			_, err := rabbit.DeclareQueue(ch, queue)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
