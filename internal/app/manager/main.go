package manager

import (
	"bitbucket.org/airenas/listgo/internal/pkg/loader"
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/mongo"
	"bitbucket.org/airenas/listgo/internal/pkg/rabbit"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"

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
	data.fc = utils.NewSignalChannel()

	mongoSessionProvider, err := mongo.NewSessionProvider()
	cmdapp.CheckOrPanic(err, "Can't init mongo provider")
	defer mongoSessionProvider.Close()

	msgChannelProvider, err := rabbit.NewChannelProvider()
	cmdapp.CheckOrPanic(err, "Can't init rabbit provider")
	defer msgChannelProvider.Close()

	err = initQueues(msgChannelProvider)
	cmdapp.CheckOrPanic(err, "Can't init queues")
	err = initEventExchange(msgChannelProvider)
	cmdapp.CheckOrPanic(err, "Can't init event exchange")

	data.MessageSender = rabbit.NewSender(msgChannelProvider)
	if cmdapp.Config.GetBool("sendInformMessages") {
		data.InformMessageSender = data.MessageSender
	} else {
		data.InformMessageSender = newFakeMessageSender()
	}

	data.Publisher = rabbit.NewPublisher(msgChannelProvider)

	ch, err := msgChannelProvider.Channel()
	cmdapp.CheckOrPanic(err, "Can't open channel")

	err = ch.Qos(1, 0, false)
	cmdapp.CheckOrPanic(err, "Can't set Qos")

	data.DecodeCh = makeQChannel(ch, msgChannelProvider.QueueName(messages.Decode))
	data.AudioConvertCh = makeQChannel(ch, msgChannelProvider.QueueName(messages.ResultQueueFor(messages.AudioConvert)))
	data.DiarizationCh = makeQChannel(ch, msgChannelProvider.QueueName(messages.ResultQueueFor(messages.Diarization)))
	data.TranscriptionCh = makeQChannel(ch, msgChannelProvider.QueueName(messages.ResultQueueFor(messages.Transcription)))
	data.RescoreCh = makeQChannel(ch, msgChannelProvider.QueueName(messages.ResultQueueFor(messages.Rescore)))
	data.ResultMakeCh = makeQChannel(ch, msgChannelProvider.QueueName(messages.ResultQueueFor(messages.ResultMake)))

	data.StatusSaver, err = mongo.NewStatusSaver(mongoSessionProvider)
	cmdapp.CheckOrPanic(err, "Can't init status saver")
	data.ResultSaver, err = mongo.NewResultSaver(mongoSessionProvider)
	cmdapp.CheckOrPanic(err, "Can't init result saver")
	data.speechIndicator, err = loader.NewNonEmptyFileTester(cmdapp.Config.GetString("speechIndicator.pathPattern"))
	cmdapp.CheckOrPanic(err, "Can't init result saver")

	err = StartWorkerService(&data)
	cmdapp.CheckOrPanic(err, "Can't start worker service")

	<-data.fc.C
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
		queues := []string{messages.Decode, messages.Inform,
			messages.AudioConvert, messages.ResultQueueFor(messages.AudioConvert),
			messages.Diarization, messages.ResultQueueFor(messages.Diarization),
			messages.Transcription, messages.ResultQueueFor(messages.Transcription),
			messages.Rescore, messages.ResultQueueFor(messages.Rescore),
			messages.ResultMake, messages.ResultQueueFor(messages.ResultMake)}
		for _, queue := range queues {
			_, err := rabbit.DeclareQueue(ch, prv.QueueName(queue))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func initEventExchange(prv *rabbit.ChannelProvider) error {
	cmdapp.Log.Info("Initializing exchanges")
	return prv.RunOnChannelWithRetry(func(ch *amqp.Channel) error {
		return rabbit.DeclareExchange(ch, prv.QueueName(messages.TopicStatusChange))
	})
}
