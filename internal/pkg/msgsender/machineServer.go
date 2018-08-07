package msgsender

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"

	"github.com/RichardKnop/machinery/v1"
)

//NewMachineryServer initializes machinery server
func NewMachineryServer() (*machinery.Server, error) {
	config := NewServerConfig()
	cmdapp.Log.Infof("Initializing the machinery server at: %s", config.Broker)
	return machinery.NewServer(config)
}
