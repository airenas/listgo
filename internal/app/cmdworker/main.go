package cmdworker

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/config"
	"bitbucket.org/airenas/listgo/internal/pkg/rabbit"
	"bitbucket.org/airenas/listgo/internal/pkg/tasks"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/streadway/amqp"
)

var appName = "LiST Worker Service"

var rootCmd = &cobra.Command{
	Use:   "cmdWorkerService",
	Short: appName,
	Long:  `Worker service listens for the work event from the queue and invokes configured command to do the work`,
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
	if err != nil {
		panic(err)
	}
	data := ServiceData{}

	data.RecInfoLoader, err = config.NewFileRecognizerInfoLoader(cmdapp.Config.GetString("recognizerConfig.path"))
	cmdapp.CheckOrPanic(err, "Can't init recognizer info loader config (Did you provide correct setting 'recognizerConfig.path'?)")

	msgChannelProvider, err := rabbit.NewChannelProvider()
	cmdapp.CheckOrPanic(err, "Can't init rabbit channel provider")
	defer msgChannelProvider.Close()

	data.MessageSender = rabbit.NewSender(msgChannelProvider)
	queueName := ""
	data.WorkCh, queueName, err = initWorkQueue(msgChannelProvider)
	cmdapp.CheckOrPanic(err, "Can't connect/prepare work queue")
	_ = queueName

	data.Name = cmdapp.Config.GetString("worker.name")
	data.Command = cmdapp.Config.GetString("worker.command")
	data.WorkingDir = cmdapp.Config.GetString("worker.workingDir")
	data.ResultFile = cmdapp.Config.GetString("worker.resultFile")
	data.LogFile = cmdapp.Config.GetString("worker.logFile")
	data.ReadFunc = ReadFile

	data.PreloadManager, err = initPreloadManager()
	cmdapp.CheckOrPanic(err, "Can't init preload task manager")
	defer data.PreloadManager.Close()

	fc, err := StartWorkerService(&data)
	cmdapp.CheckOrPanic(err, "Can't start service")

	<-fc
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

///////////////////////////////////////////////////////////////////////////////////
// init prepload task manager
///////////////////////////////////////////////////////////////////////////////////
func initPreloadManager() (PreloadTaskManager, error) {
	kp := cmdapp.Config.GetString("worker.preloadKeyPrefix")
	if kp == "" {
		return &fakePreloadManager{}, nil
	}
	return tasks.NewManager(kp, cmdapp.Config.GetString("worker.workingDir"))
}

type fakePreloadManager struct{}

func (pm *fakePreloadManager) EnsureRunning(map[string]string) error {
	return nil
}

func (pm *fakePreloadManager) Close() error {
	return nil
}

///////////////////////////////////////////////////////////////////////////////////
func getPrivateQueue(ch *amqp.Channel) (<-chan amqp.Delivery, string, error) {
	q, err := ch.QueueDeclare("", // name
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return nil, "", errors.Wrap(err, "Can't init private queue")
	}
	cd, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	return cd, q.Name, err
}

///////////////////////////////////////////////////////////////////////////
func initWorkQueue(msgChannelProvider *rabbit.ChannelProvider) (<-chan amqp.Delivery, string, error) {
	ch, err := msgChannelProvider.Channel()
	if err != nil {
		return nil, "", errors.Wrap(err, "Can't open channel")
	}
	err = ch.Qos(1, 0, false)
	if err != nil {
		return nil, "", errors.Wrap(err, "Can't set Qos")
	}
	rQueue := cmdapp.Config.GetString("registry.queue")
	if rQueue == "" {
		queue := cmdapp.Config.GetString("worker.queue")
		cmdapp.Log.Infof("Try listen static queue %s", queue)
		if queue == "" {
			return nil, "", errors.Errorf("No worker.queue configured!")
		}
		res, err := rabbit.NewChannel(ch, queue)
		if err != nil {
			return nil, "", errors.Wrap(err, "Can't listen "+queue+" channel")
		}
		return res, queue, nil
	}

	cmdapp.Log.Infof("Creating private worker queue")
	return getPrivateQueue(ch)
}
