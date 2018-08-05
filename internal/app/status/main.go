package status

import (
	"os"

	"bitbucket.org/airenas/listgo/internal/pkg/mongo"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/spf13/cobra"
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

func logPanic() {
	if r := recover(); r != nil {
		cmdapp.Log.Error(r)
		os.Exit(1)
	}
}

//Execute starts the server
func Execute() {
	defer logPanic()
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func run(cmd *cobra.Command, args []string) {
	cmdapp.Log.Info("Starting " + appName)
	mongoSessionProvider, err := mongo.NewSessionProvider()
	if err != nil {
		panic(err)
	}
	defer mongoSessionProvider.Close()

	statusProvider, err := mongo.NewStatusProvider(mongoSessionProvider)
	if err != nil {
		panic(err)
	}
	err = StartWebServer(&ServiceData{*statusProvider, cmdapp.Config.GetInt("port")})
	if err != nil {
		panic(err)
	}
}
