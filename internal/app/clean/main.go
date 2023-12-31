package clean

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/airenas/listgo/internal/pkg/metrics"
	"github.com/airenas/listgo/internal/pkg/mongo"
	"github.com/heptiolabs/healthcheck"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
)

var appName = "LiST Data Clean Service"

var rootCmd = &cobra.Command{
	Use:   "cleanService",
	Short: appName,
	Long:  `Service to provide data deletion functionality`,
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

	data := &ServiceData{}
	err := initMetrics(data)
	cmdapp.CheckOrPanic(err, "Can't init metrics")

	data.Port = cmdapp.Config.GetInt("port")
	mongoSessionProvider, err := mongo.NewSessionProvider()
	cmdapp.CheckOrPanic(err, "Can't init mongo")
	defer mongoSessionProvider.Close()

	cm, err := newCleanMetric()
	cmdapp.CheckOrPanic(err, "Can't init clean metrics")
	cln, err := newCleanerImpl(mongoSessionProvider, cmdapp.Config.GetString("fileStorage.path"),
		cmdapp.Config.GetString("fileStorage.patterns"), cm)
	cmdapp.CheckOrPanic(err, "Can't init cleaner")
	data.cleaner = cln

	data.health = healthcheck.NewHandler()
	data.health.AddLivenessCheck("mongo", healthcheck.Async(mongoSessionProvider.Healthy, 10*time.Second))
	data.health.AddLivenessCheck("fs", cln.HealthyFunc())

	tdata := timerServiceData{}
	tdata.runEvery = time.Hour
	tdata.cleaner = data.cleaner
	expireDuraton := cmdapp.Config.GetDuration("expireDuration")
	if expireDuraton < time.Minute {
		cmdapp.CheckOrPanic(errors.Errorf("Wrong expire duration %v", expireDuraton), "Can't init mongo expired IDs provider")
	}
	cmdapp.Log.Infof("Expire duration %v", expireDuraton)

	tdata.idsProvider, err = mongo.NewCleanIDsProvider(mongoSessionProvider, expireDuraton)
	cmdapp.CheckOrPanic(err, "Can't init mongo expired IDs provider")
	tdata.qChan = make(chan struct{})
	tdata.workWaitChan = make(chan struct{})

	go func() {
		err = StartWebServer(data)
		cmdapp.CheckOrPanic(err, "")
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	err = startCleanTimer(&tdata)
	cmdapp.CheckOrPanic(err, "")

	<-sigs
	cmdapp.Log.Infof("Stopping")
	// indicate to stop and wait for job complete
	close(tdata.qChan)
	<-tdata.workWaitChan
}

const namespace = "clean_service"

func initMetrics(data *ServiceData) error {

	data.metrics.responseDur = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "request_durations_seconds",
			Help:      "Request latency distributions.",
		}, nil)

	return metrics.Register(data.metrics.responseDur)
}

func newCleanMetric() (prometheus.Counter, error) {
	res := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "clean_counter",
			Help:      "Count of clean invoked.",
		})

	err := metrics.Register(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
