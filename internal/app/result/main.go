package result

import (
	"time"

	"github.com/airenas/listgo/internal/pkg/loader"
	"github.com/airenas/listgo/internal/pkg/metrics"
	"github.com/airenas/listgo/internal/pkg/mongo"

	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "resultService",
	Short: "LiST Transcription Result Service",
	Long:  `HTTP server to provide results for transcription`,
	Run:   run,
}

func init() {
	cmdapp.InitApplication(rootCmd)
	rootCmd.PersistentFlags().Int32P("port", "", 8000, "Default service port")
	cmdapp.Config.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	cmdapp.Config.SetDefault("port", 8080)
	cmdapp.Config.SetDefault("fileStorage.audio", "/data/audio.in/")
}

// Execute starts the server
func Execute() {
	cmdapp.Execute(rootCmd)
}

func run(cmd *cobra.Command, args []string) {
	cmdapp.Log.Info("Starting resultService")
	data := &ServiceData{}
	err := initMetrics(data)
	cmdapp.CheckOrPanic(err, "Can't init metrics")

	data.health = healthcheck.NewHandler()
	mongoSessionProvider, err := mongo.NewSessionProvider()
	cmdapp.CheckOrPanic(err, "Can't init mongo session")

	defer mongoSessionProvider.Close()
	data.health.AddLivenessCheck("mongo", healthcheck.Async(mongoSessionProvider.Healthy, 10*time.Second))

	data.fileNameProvider, err = mongo.NewFileNameProvider(mongoSessionProvider)
	cmdapp.CheckOrPanic(err, "Can't init fileName provider")

	data.audioFileLoader, err = loader.NewLocalFileLoader(cmdapp.Config.GetString("fileStorage.audio"))
	cmdapp.CheckOrPanic(err, "Can't init audioFileLoader provider")

	data.resultFileLoader, err = loader.NewLocalFileLoader(cmdapp.Config.GetString("fileStorage.results"))
	cmdapp.CheckOrPanic(err, "Can't init resultFileLoader provider")
	data.port = cmdapp.Config.GetInt("port")

	err = StartWebServer(data)
	cmdapp.CheckOrPanic(err, "Can't start web server")
}

func initMetrics(data *ServiceData) error {
	namespace := "result_service"
	data.metrics.audioResponseDur = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "audio_request_durations_seconds",
			Help:      "Audio request latency distributions.",
		}, nil)

	err := metrics.Register(data.metrics.audioResponseDur)
	if err != nil {
		return err
	}
	data.metrics.audioResponseSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Name:      "audio_response_size_bytes",
			Help:      "Audio response size in bytes."}, nil)
	err = metrics.Register(data.metrics.audioResponseSize)
	if err != nil {
		return err
	}
	data.metrics.resultResponseDur = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "result_request_durations_seconds",
			Help:      "Result request latency distributions.",
		}, nil)

	err = metrics.Register(data.metrics.resultResponseDur)
	if err != nil {
		return err
	}
	data.metrics.resultResponseSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Name:      "result_response_size_bytes",
			Help:      "Result response size in bytes."}, nil)
	return metrics.Register(data.metrics.resultResponseSize)
}
