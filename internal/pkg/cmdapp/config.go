package cmdapp

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//Config is a viper based application config
var Config = viper.New()

//Log is applications logger
var Log = logrus.New()
