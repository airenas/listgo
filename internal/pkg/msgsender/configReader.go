package msgsender

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/RichardKnop/machinery/v1/config"
)

//NewServerConfig reads the mesage broker config from viper
func NewServerConfig() *config.Config {
	return &config.Config{
		Broker:        cmdapp.Config.GetString("messageServer.broker"),
		DefaultQueue:  cmdapp.Config.GetString("messageServer.defaultQueue"),
		ResultBackend: cmdapp.Config.GetString("messageServer.resultBackend")}
}
