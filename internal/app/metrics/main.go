package metrics

import (
	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/airenas/listgo/internal/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
)

var appName = "LiST Metrics Collector"

var rootCmd = &cobra.Command{
	Use:   "metricsCollector",
	Short: appName,
	Long:  `HTTP server to collect and expose metrics`,
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

	data, err := newServiceData()
	cmdapp.CheckOrPanic(err, "Can' int metrics")

	data.Port = cmdapp.Config.GetInt("port")

	err = StartWebServer(data)
	cmdapp.CheckOrPanic(err, "")
}

func initMetrics(data *ServiceData) error {
	namespace := "metrics_collector"
	data.tasksMetrics = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "task_duration_seconds",
			Help:      "Tasks duration metrics",
			Buckets:   prometheus.ExponentialBuckets(0.5, 2, 15),
		}, []string{"worker", "task", "model"})
	err := metrics.Register(data.tasksMetrics)
	if err != nil {
		return err
	}
	data.tasksStarted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "task_start",
			Help:      "Tasks start counter",
		}, []string{"worker", "task", "model"})
	err = metrics.Register(data.tasksStarted)
	if err != nil {
		return err
	}
	data.tasksEnded = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "task_end",
			Help:      "Tasks end counter",
		}, []string{"worker", "task", "model"})
	err = metrics.Register(data.tasksEnded)
	if err != nil {
		return err
	}

	data.metricDur = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "request_durations_seconds",
			Help:      "Request latency distributions.",
		}, nil)

	return metrics.Register(data.metricDur)
}
