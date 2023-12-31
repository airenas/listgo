package status

import (
	"time"

	"github.com/airenas/listgo/internal/pkg/messages"
	"github.com/airenas/listgo/internal/pkg/metrics"
	"github.com/airenas/listgo/internal/pkg/mongo"
	"github.com/airenas/listgo/internal/pkg/rabbit"

	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/heptiolabs/healthcheck"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
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

// Execute starts the server
func Execute() {
	cmdapp.Execute(rootCmd)
}

func run(cmd *cobra.Command, args []string) {
	cmdapp.Log.Info("Starting " + appName)
	mongoSessionProvider, err := mongo.NewSessionProvider()
	cmdapp.CheckOrPanic(err, "")
	defer mongoSessionProvider.Close()

	data := &ServiceData{}
	err = initMetrics(data)
	cmdapp.CheckOrPanic(err, "Can't init metrics")

	data.health = healthcheck.NewHandler()
	data.StatusProvider, err = mongo.NewStatusProvider(mongoSessionProvider)
	cmdapp.CheckOrPanic(err, "")
	data.health.AddLivenessCheck("mongo", healthcheck.Async(mongoSessionProvider.Healthy, 10*time.Second))

	msgChannelProvider, err := rabbit.NewChannelProvider()
	cmdapp.CheckOrPanic(err, "")
	defer msgChannelProvider.Close()
	data.health.AddLivenessCheck("rabbit", healthcheck.Async(msgChannelProvider.Healthy, 10*time.Second))

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
	err = ch.QueueBind(q.Name, "", provider.QueueName(messages.TopicStatusChange), false, nil)
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

func initMetrics(data *ServiceData) error {
	namespace := "status_service"
	data.metrics.responseDur = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "request_durations_seconds",
			Help:      "Request latency distributions.",
		}, nil)

	err := metrics.Register(data.metrics.responseDur)
	if err != nil {
		return err
	}
	data.metrics.responseSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Name:      "request_response_size_bytes",
			Help:      "Response size in bytes."}, nil)
	return metrics.Register(data.metrics.responseSize)
}
