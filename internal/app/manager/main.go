package manager

import (
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/mongo"
	"bitbucket.org/airenas/listgo/internal/pkg/rabbit"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
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
	cmdapp.CheckOrPanic(err, "Can't init mongo provider")

	defer mongoSessionProvider.Close()

	msgChannelProvider, err := rabbit.NewChannelProvider(cmdapp.Config.GetString("messageServer.broker"))
	cmdapp.CheckOrPanic(err, "Can't init rabbit provider")
	defer msgChannelProvider.Close()

	err = initQueues(msgChannelProvider)
	cmdapp.CheckOrPanic(err, "Can't init queues")

	data.MessageSender = rabbit.NewSender(msgChannelProvider)

	ch, err := msgChannelProvider.Channel()
	cmdapp.CheckOrPanic(err, "Can't open channel")

	err = ch.Qos(1, 0, false)
	cmdapp.CheckOrPanic(err, "Can't set Qos")

	data.DecodeCh = makeQChannel(ch, messages.Decode)
	data.AudioConvertCh = makeQChannel(ch, messages.ResultQueueFor(messages.AudioConvert))
	data.DiarizationCh = makeQChannel(ch, messages.ResultQueueFor(messages.Diarization))
	data.TranscriptionCh = makeQChannel(ch, messages.ResultQueueFor(messages.Transcription))
	data.ResultMakeCh = makeQChannel(ch, messages.ResultQueueFor(messages.ResultMake))

	data.StatusSaver, err = mongo.NewStatusSaver(mongoSessionProvider)
	cmdapp.CheckOrPanic(err, "Can't init status saver")

	fc, err := StartWorkerService(&data)
	cmdapp.CheckOrPanic(err, "Can't start worker service")

	<-fc
	cmdapp.Log.Infof("Exiting service")
}

func makeQChannel(ch *amqp.Channel, qname string) <-chan amqp.Delivery {
	result, err := rabbit.NewChannel(ch, qname)
	cmdapp.CheckOrPanic(err, "Can't listen "+qname+" channel")
	return result
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
