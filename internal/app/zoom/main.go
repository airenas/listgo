package zoom

import (
	"bitbucket.org/airenas/listgo/internal/app/manager"
	"bitbucket.org/airenas/listgo/internal/pkg/audio"
	"bitbucket.org/airenas/listgo/internal/pkg/loader"
	"bitbucket.org/airenas/listgo/internal/pkg/messages"
	"bitbucket.org/airenas/listgo/internal/pkg/mongo"
	"bitbucket.org/airenas/listgo/internal/pkg/rabbit"
	"bitbucket.org/airenas/listgo/internal/pkg/saver"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
)

var appName = "LiST Zoom Manager Service"

var rootCmd = &cobra.Command{
	Use:   "zoomService",
	Short: appName,
	Long:  `Zoom manager service leads multi file audio transcription process`,
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
	cmdapp.CheckOrPanic(err, "can't init mongo provider")
	defer mongoSessionProvider.Close()

	msgChannelProvider, err := rabbit.NewChannelProvider()
	cmdapp.CheckOrPanic(err, "can't init rabbit provider")
	defer msgChannelProvider.Close()

	err = initQueues(msgChannelProvider)
	cmdapp.CheckOrPanic(err, "can't init queues")

	data.MessageSender = rabbit.NewSender(msgChannelProvider)
	if cmdapp.Config.GetBool("sendInformMessages") {
		data.InformMessageSender = data.MessageSender
	} else {
		data.InformMessageSender = manager.NewFakeMessageSender()
	}

	data.Publisher = rabbit.NewPublisher(msgChannelProvider)

	ch, err := msgChannelProvider.Channel()
	cmdapp.CheckOrPanic(err, "can't open channel")

	err = ch.Qos(1, 0, false)
	cmdapp.CheckOrPanic(err, "can't set Qos")

	data.DecodeMultiCh = makeQChannel(ch, msgChannelProvider.QueueName(messages.DecodeMultiple))
	data.OneStatusCh = makeQChannel(ch, msgChannelProvider.QueueName(messages.OneStatus))
	data.OneCompletedCh = makeQChannel(ch, msgChannelProvider.QueueName(messages.OneCompleted))

	data.StatusSaver, err = mongo.NewStatusSaver(mongoSessionProvider)
	cmdapp.CheckOrPanic(err, "can't init status saver")
	data.ResultSaver, err = mongo.NewResultSaver(mongoSessionProvider)
	cmdapp.CheckOrPanic(err, "can't init result saver")
	data.FilesGetter, err = loader.NewLocalFileList(cmdapp.Config.GetString("audio.path"))
	cmdapp.CheckOrPanic(err, "can't init files loader")
	data.Loader, err = loader.NewLocalFileLoader(cmdapp.Config.GetString("audio.path"))
	cmdapp.CheckOrPanic(err, "can't init file loader")
	data.AudioLen, err = audio.NewDurationClient(cmdapp.Config.GetString("audio.durationUrl"))
	cmdapp.CheckOrPanic(err, "can't init file loader")
	data.FileSaver, err = saver.NewLocalFileSaver(cmdapp.Config.GetString("audio.path"))
	cmdapp.CheckOrPanic(err, "can't init file storage")
	data.RequestSaver, err = mongo.NewRequestSaver(mongoSessionProvider)
	cmdapp.CheckOrPanic(err, "can't init request saver")

	err = StartWorkerService(&data)
	cmdapp.CheckOrPanic(err, "can't start worker service")

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
		queues := []string{messages.DecodeMultiple,
			messages.JoinAudio, messages.ResultQueueFor(messages.JoinAudio),
			messages.JoinResults, messages.ResultQueueFor(messages.JoinResults),
			messages.OneCompleted, messages.OneStatus}
		for _, queue := range queues {
			_, err := rabbit.DeclareQueue(ch, prv.QueueName(queue))
			if err != nil {
				return err
			}
		}
		return nil
	})
}
