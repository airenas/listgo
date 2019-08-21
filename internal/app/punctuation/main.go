package punctuation

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/spf13/cobra"
)

var appName = "LiST Punctuation Restoration Service"

var rootCmd = &cobra.Command{
	Use:   "punctuationService",
	Short: appName,
	Long:  `HTTP server to provide punctuation restoration`,
	Run:   run,
}

func init() {
	cmdapp.InitApplication(rootCmd)
	rootCmd.PersistentFlags().Int32P("port", "", 8000, "Default service port")
	cmdapp.Config.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	cmdapp.Config.SetDefault("port", 8080)
}

//Execute starts the server
func Execute() {
	cmdapp.Execute(rootCmd)
}

func run(cmd *cobra.Command, args []string) {
	cmdapp.Log.Info("Starting " + appName)

	data := &ServiceData{}
	//data.health = healthcheck.NewHandler()
	//data.health.AddLivenessCheck("tensorflow", healthcheck.Async(mongoSessionProvider.Healthy, 10*time.Second))

	provider, err := NewSettingsDataProviderImpl(cmdapp.Config.GetString("modelDir"))
	cmdapp.CheckOrPanic(err, "Cannot init data provider")

	data.punctuator, err = NewPunctuatorImpl(provider)
	cmdapp.CheckOrPanic(err, "Cannot init punctuator")

	data.Port = cmdapp.Config.GetInt("port")

	err = StartWebServer(data)
	cmdapp.CheckOrPanic(err, "")
}
