package cmdapp

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "test",
		Long:  `test`,
		Run:   run}
}

func run(cmd *cobra.Command, args []string) {
	Log.Info("Starting uploadService")
}

func TestReadEnvironmentVariable(t *testing.T) {
	os.Setenv("MESSAGESERVER_URL", "olia")
	InitApplication(newRootCmd())

	assert.Equal(t, "olia", Config.GetString("messageServer.url"))
}

func TestReadConfig(t *testing.T) {
	initAppFromTempFile(t, "messageServer:\n     url: olia\n")

	assert.Equal(t, "olia", Config.GetString("messageServer.url"))
}

func TestEnvBeatsConfig(t *testing.T) {
	os.Setenv("MESSAGESERVER_URL", "xxxx")
	initAppFromTempFile(t, "messageServer:\n     url: olia\n")

	assert.Equal(t, "xxxx", Config.GetString("messageServer.url"))
}

func TestDefaultLogger(t *testing.T) {
	initDefaultLevel()
	initAppFromTempFile(t, "")

	assert.Equal(t, "info", Log.GetLevel().String())
}

func TestLoggerInitFromConfig(t *testing.T) {
	initDefaultLevel()
	initAppFromTempFile(t, "logger:\n    level: trace\n")

	assert.Equal(t, "trace", Log.GetLevel().String())
}

func TestLoggerLevelInitFromEnv(t *testing.T) {
	initDefaultLevel()

	os.Setenv("LOGGER_LEVEL", "trace")
	initAppFromTempFile(t, "logger:\n    level: info\n")

	assert.Equal(t, "trace", Log.GetLevel().String())
}

func initAppFromTempFile(t *testing.T, data string) {
	f, err := ioutil.TempFile("", "test.*.yml")
	assert.Nil(t, err)
	f.WriteString(data)
	f.Sync()

	defer os.Remove(f.Name())

	rootCmd := newRootCmd()
	InitApplication(rootCmd)
	configFile = f.Name()
	rootCmd.Execute()
}

func initDefaultLevel(){
	Log.SetLevel(logrus.ErrorLevel)
}
