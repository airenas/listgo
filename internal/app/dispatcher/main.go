package dispatcher

import (
	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/airenas/listgo/internal/pkg/config"
	"github.com/airenas/listgo/internal/pkg/rabbit"
	"github.com/airenas/listgo/internal/pkg/strategy"
	"github.com/airenas/listgo/internal/pkg/utils"

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

// Execute starts the server
func Execute() {
	cmdapp.Execute(rootCmd)
}

func run(cmd *cobra.Command, args []string) {
	cmdapp.Log.Info("Starting " + appName)
	err := validateConfig()
	cmdapp.CheckOrPanic(err, "Configuration error")

	data := ServiceData{}
	data.fc = utils.NewSignalChannel()
	data.wrkrs = newWorkers()
	data.tsks = newTasks()
	// make same lock
	data.tsks.lock = data.wrkrs.lock

	msgChannelProvider, err := rabbit.NewChannelProvider()
	cmdapp.CheckOrPanic(err, "Can't init rabbit channel provider")
	defer msgChannelProvider.Close()

	rbSender := rabbit.NewSender(msgChannelProvider)
	data.replySender = rbSender

	ch, err := msgChannelProvider.Channel()
	cmdapp.CheckOrPanic(err, "Can't open channel")
	err = ch.Qos(1, 0, false)
	cmdapp.CheckOrPanic(err, "Can't set Qos")

	registrationQueue := cmdapp.Config.GetString("dispatcher.registrationQueue")

	err = initRegistrationQueue(msgChannelProvider, registrationQueue)
	data.RegistrationCh, err = rabbit.NewChannel(ch, registrationQueue)
	cmdapp.CheckOrPanic(err, "Can't listen "+registrationQueue+" channel")

	respQName := ""
	data.ResponseCh, respQName, err = initResponseQueue(ch)
	cmdapp.CheckOrPanic(err, "Can't init response queue")
	data.workSender, err = newMsgWithCorrSender(rbSender, respQName)
	cmdapp.CheckOrPanic(err, "Can't init work queue sender")

	// work queue initialization
	msgWorkChannelProvider, err := rabbit.NewChannelProvider()
	cmdapp.CheckOrPanic(err, "Can't init rabbit work channel provider")
	defer msgWorkChannelProvider.Close()

	data.WorkCh, err = initWorkQueue(msgWorkChannelProvider)
	cmdapp.CheckOrPanic(err, "Can't listen channel")
	//end work queue
	data.modelLoadDuration = cmdapp.Config.GetDuration("strategy.modelLoadDuration")
	data.rtFactor = cmdapp.Config.GetFloat64("strategy.realTimeFactor")
	cmdapp.Log.Infof("Dispatch params: modelLoadTime=%v, rt=%f", data.modelLoadDuration, data.rtFactor)
	strg, err := strategy.NewCost()
	cmdapp.CheckOrPanic(err, "Can't init strategy")
	data.selectionStrategy, err = newStrategyWrapper(strg)
	cmdapp.CheckOrPanic(err, "Can't init strategy wrapper")

	recProvider, err := config.NewFileRecognizerInfoLoader(cmdapp.Config.GetString("recognizerConfig.path"))
	cmdapp.CheckOrPanic(err, "Can't init recognizer config (Did you provide correct setting 'recognizerConfig.path'?)")
	data.modelTypeGetter, err = newTypeGetter(recProvider, cmdapp.Config.GetString("recognizerConfig.key"))
	cmdapp.CheckOrPanic(err, "Can't init model type getter. recognizerConfig.key config missing?")
	data.durationGetter, err = newDurationLoader(cmdapp.Config.GetString("duration.pathPattern"))
	cmdapp.CheckOrPanic(err, "Can't init duration loader. duration.pathPattern config missing?")
	data.startTimeGetter = newTimeGetter()

	err = StartWorkerService(&data)
	cmdapp.CheckOrPanic(err, "Can't start service")

	<-data.fc.C
	cmdapp.Log.Infof("Bye")
}

// /////////////////////////////////////////////////////////////////////////
func validateConfig() error {
	if cmdapp.Config.GetString("dispatcher.registrationQueue") == "" {
		return errors.New("No dispatcher.registrationQueue configured")
	}
	if cmdapp.Config.GetString("dispatcher.workQueue") == "" {
		return errors.New("No dispatcher.workQueue configured")
	}
	return nil
}

// /////////////////////////////////////////////////////////////////////////
func initRegistrationQueue(prv *rabbit.ChannelProvider, qName string) error {
	return prv.RunOnChannelWithRetry(func(ch *amqp.Channel) error {
		_, err := rabbit.DeclareQueue(ch, prv.QueueName(qName))
		if err != nil {
			return err
		}
		return nil
	})
}

// /////////////////////////////////////////////////////////////////////////
func initResponseQueue(ch *amqp.Channel) (<-chan amqp.Delivery, string, error) {
	q, err := ch.QueueDeclare("", // name
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		return nil, "", errors.Wrap(err, "Can't init queue")
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

// /////////////////////////////////////////////////////////////////////////
func initWorkQueue(chPrv *rabbit.ChannelProvider) (<-chan amqp.Delivery, error) {
	workCh, err := chPrv.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "Can't open work channel")
	}
	preload := cmdapp.Config.GetInt("dispatcher.preload")
	cmdapp.Log.Infof("Msg preload count %d", preload)
	err = workCh.Qos(preload, 0, false)
	if err != nil {
		return nil, errors.Wrap(err, "Can't set Qos for work channel")
	}
	workQueue := cmdapp.Config.GetString("dispatcher.workQueue")
	wCh, err := rabbit.NewChannel(workCh, workQueue)
	if err != nil {
		return nil, errors.Wrap(err, "Can't listen "+workQueue+" channel")
	}
	return wCh, nil
}
