package cmdapp

import (
	"os"
	"path/filepath"

	"github.com/heirko/go-contrib/logrusHelper"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configFile = ""
)

// InitApplication initializes the app by reading config file
func InitApplication(rootCommand *cobra.Command) {
	viper.AutomaticEnv()
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
		panic(errors.Wrap(err, msg))
	}
}
