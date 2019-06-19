package result

import (
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/loader"
	"bitbucket.org/airenas/listgo/internal/pkg/mongo"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/heptiolabs/healthcheck"
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

//Execute starts the server
func Execute() {
	cmdapp.Execute(rootCmd)
}

func run(cmd *cobra.Command, args []string) {
	cmdapp.Log.Info("Starting resultService")
	data := ServiceData{}
	data.health = healthcheck.NewHandler()
	mongoSessionProvider, err := mongo.NewSessionProvider()
	if err != nil {
		panic(err)
	}
	defer mongoSessionProvider.Close()
	data.health.AddLivenessCheck("mongo", healthcheck.Async(mongoSessionProvider.Healthy, 10*time.Second))

	data.fileNameProvider, err = mongo.NewFileNameProvider(mongoSessionProvider)
	cmdapp.CheckOrPanic(err, "Can't init fileName provider")

	data.audioFileLoader, err = loader.NewLocalFileLoader(cmdapp.Config.GetString("fileStorage.audio"))
	cmdapp.CheckOrPanic(err, "Can't init audioFileLoader provider")

	data.resultFileLoader, err = loader.NewLocalFileLoader(cmdapp.Config.GetString("fileStorage.results"))
	cmdapp.CheckOrPanic(err, "Can't init resultFileLoader provider")
	data.port = cmdapp.Config.GetInt("port")

	err = StartWebServer(&data)
	cmdapp.CheckOrPanic(err, "Can't start web server")
}
