package cmdapp

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
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
	os.Setenv("MESSAGESERVER_URL", "olia")
	InitApplication(rootCmd)
	a := Config.GetString("messageServer.url")
	assert.Equal(t, a, "olia")
}
