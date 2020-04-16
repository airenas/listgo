package dispatcher

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/rabbit"
	"bitbucket.org/airenas/listgo/internal/pkg/strategy"
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
	data.tsks = newTasks()

	msgChannelProvider, err := rabbit.NewChannelProvider()
	cmdapp.CheckOrPanic(err, "Can't init rabbit channel provider")
	defer msgChannelProvider.Close()

	rbSender := rabbit.NewSender(msgChannelProvider)
	data.replySender = rbSender

	ch, err := msgChannelProvider.Channel()
	cmdapp.CheckOrPanic(err, "Can't open channel")
	err = ch.Qos(1, 0, false)
	cmdapp.CheckOrPanic(err, "Can't set Qos")

	registrationQueue := cmdapp.Config.GetString("messageServer.registrationQueue")

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
	strg, err := strategy.NewCost(data.modelLoadDuration)
	cmdapp.CheckOrPanic(err, "Can't init strategy")
	data.selectionStrategy, err = newStrategyWrapper(strg)
	cmdapp.CheckOrPanic(err, "Can't init strategy wrapper")

	err = StartWorkerService(&data)
	cmdapp.CheckOrPanic(err, "Can't start service")

	<-data.fc.C
	cmdapp.Log.Infof("Exiting service")
}

///////////////////////////////////////////////////////////////////////////
func validateConfig() error {
	if cmdapp.Config.GetString("messageServer.registrationQueue") == "" {
		return errors.New("No messageServer.registrationQueue configured")
	}
	if cmdapp.Config.GetString("dispatcher.workQueue") == "" {
		return errors.New("No dispatcher.workQueue configured")
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////
func initRegistrationQueue(prv *rabbit.ChannelProvider, qName string) error {
	return prv.RunOnChannelWithRetry(func(ch *amqp.Channel) error {
		_, err := rabbit.DeclareQueue(ch, prv.QueueName(qName))
		if err != nil {
			return err
		}
		return nil
	})
}

///////////////////////////////////////////////////////////////////////////
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

///////////////////////////////////////////////////////////////////////////
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
