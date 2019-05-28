package cmdapp

import (
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/heirko/go-contrib/logrusHelper"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

var (
	configFile = ""
)

// InitApplication initializes the app by reading config file
func InitApplication(rootCommand *cobra.Command) {
	// make environment variable WEB_URL be found by viper with key web.url
	Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	Config.AutomaticEnv()
	cobra.OnInitialize(initConfig)
	rootCommand.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (default is config.yaml)")
}

func initConfig() {
	failOnNoFail := false
	if configFile != "" {
		// Use config file from the flag.
		Config.SetConfigFile(configFile)
		failOnNoFail = true
	} else {
		// Find home directory.
		ex, err := os.Executable()
		if err != nil {
			Log.Error("Can't get the app directory:", err)
			panic(1)
		}
		Config.AddConfigPath(filepath.Dir(ex))
		Config.SetConfigName("config")
	}

	if err := Config.ReadInConfig(); err != nil {
		Log.Warn("Can't read config:", err)
		if failOnNoFail {
			Log.Error("Exiting the app")
			panic(1)
		}
	}
	initLog()
	Log.Info("Config loaded from: ", Config.ConfigFileUsed())
}

func initLog() {
	initDefaultLogConfig()
	c := logrusHelper.UnmarshalConfiguration(Config.Sub("logger"))
	err := logrusHelper.SetConfig(Log, c)
	if err != nil {
		Log.Error("Can't init log ", err)
	}
}

func initDefaultLogConfig() {
	defaultLogConfig := map[string]interface{}{
		"level":                              "info",
		"formatter.name":                     "text",
		"formatter.options.full_timestamp":   true,
		"formatter.options.timestamp_format": "2006-01-02T15:04:05.000",
	}
	Config.SetDefault("logger", defaultLogConfig)
}

func logPanic() {
	if r := recover(); r != nil {
		Log.Error(r)
		os.Exit(1)
	}
}

//Execute the main command
func Execute(cmd *cobra.Command) {
	defer logPanic()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

//CheckOrPanic panics if err != nil
func CheckOrPanic(err error, msg string) {
	if err != nil {
		if msg == "" {
			panic(err)
		} else {
			panic(errors.Wrap(err, msg))
		}
	}
}

//LogIf logs error if err != nil
func LogIf(err error) {
	if err != nil {
		Log.Error(err)
	}
}

//NewSignalChannel returns new channel that listens for system interupts
func NewSignalChannel() chan os.Signal {
	fc := make(chan os.Signal)
	signal.Notify(fc, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	return fc
}
