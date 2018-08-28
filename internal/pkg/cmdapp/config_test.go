package cmdapp

import (
	"os"
	"testing"

	"github.com/spf13/cobra"

	. "github.com/smartystreets/goconvey/convey"
)

var rootCmd = &cobra.Command{
	Use:   "test",
	Short: "test",
	Long:  `test`,
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	Log.Info("Starting uploadService")
}

func TestReadEnvironmentVariable(t *testing.T) {
	Convey("Given an environment variable and app init", t, func() {
		os.Setenv("MESSAGESERVER_URL", "olia")
		InitApplication(rootCmd)
		Convey("viper reads it", func() {
			a := Config.GetString("messageServer.url")
			So(a, ShouldEqual, "olia")
		})
	})
}
