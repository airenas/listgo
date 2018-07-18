package cmdapp

import (
	"log"
	"os"
	"path/filepath"

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
		viper.SetConfigFile(configFile)
		failOnNoFail = true
	} else {
		// Find home directory.
		ex, err := os.Executable()
		if err != nil {
			log.Fatalln("Can't get the app directory:", err)
			panic(1)
		}
		viper.AddConfigPath(filepath.Dir(ex))
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Println("Can't read config:", err)
		if failOnNoFail {
			log.Fatalln("Exiting the app")
			panic(1)
		}
	} else {
		log.Println("Config loaded from: ", viper.ConfigFileUsed())
	}
}
