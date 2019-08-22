package punctuation

import (
	"time"

	"bitbucket.org/airenas/listgo/internal/app/punctuation/tf"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/heptiolabs/healthcheck"
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

	provider, err := NewSettingsDataProviderImpl(cmdapp.Config.GetString("modelDir"))
	cmdapp.CheckOrPanic(err, "Cannot init data provider")

	tfWrapper, err := tf.NewWrapper(cmdapp.Config.GetString("tf.url"), cmdapp.Config.GetString("tf.name"),
		cmdapp.Config.GetInt("tf.version"))
	cmdapp.CheckOrPanic(err, "Cannot init tensorflow wrapper")

	data.health = healthcheck.NewHandler()
	data.health.AddLivenessCheck("tensorflow", healthcheck.Async(tfWrapper.Healthy, 10*time.Second))

	data.punctuator, err = NewPunctuatorImpl(provider, tfWrapper)
	cmdapp.CheckOrPanic(err, "Cannot init punctuator")

	data.Port = cmdapp.Config.GetInt("port")

	err = StartWebServer(data)
	cmdapp.CheckOrPanic(err, "")
}
