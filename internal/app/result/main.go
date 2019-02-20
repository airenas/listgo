package result

import (
	"bitbucket.org/airenas/listgo/internal/pkg/loader"
	"bitbucket.org/airenas/listgo/internal/pkg/mongo"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
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
	mongoSessionProvider, err := mongo.NewSessionProvider()
	if err != nil {
		panic(err)
	}
	defer mongoSessionProvider.Close()

	fileNameProvider, err := mongo.NewFileNameProvider(mongoSessionProvider)
	cmdapp.CheckOrPanic(err, "Can't init fileName provider")

	fileLoader, err := loader.NewLocalFileLoader(cmdapp.Config.GetString("fileStorage.audio"))
	cmdapp.CheckOrPanic(err, "Can't init fileLoader provider")

	err = StartWebServer(&ServiceData{fileLoader, fileNameProvider, cmdapp.Config.GetInt("port")})
	cmdapp.CheckOrPanic(err, "Can't start web server")
}
