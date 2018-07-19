package msgsender

import (
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/spf13/viper"
)

//NewServerConfig reads the mesage broker config from viper
func NewServerConfig() *config.Config {
	return &config.Config{
		Broker:        viper.GetString("messageServer.broker"),
		DefaultQueue:  viper.GetString("messageServer.defaultQueue"),
		ResultBackend: viper.GetString("messageServer.resultBackend")}
}
