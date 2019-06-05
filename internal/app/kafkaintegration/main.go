package kafkaintegration

import (
	"bitbucket.org/airenas/listgo/internal/pkg/file"
	"bitbucket.org/airenas/listgo/internal/pkg/fs"
	"bitbucket.org/airenas/listgo/internal/pkg/kafka"
	transcriberapi "bitbucket.org/airenas/listgo/internal/pkg/transcriber"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/spf13/cobra"
)

var appName = "LiST Kafka Integration Service"

var rootCmd = &cobra.Command{
	Use:   "kafkaIntegrationService",
	Short: appName,
	Long:  `Service to handle integration with kafka messages`,
	Run:   run,
}

func init() {
	cmdapp.InitApplication(rootCmd)
}

//Execute starts the server
func Execute() {
	cmdapp.Execute(rootCmd)
}

func run(cmd *cobra.Command, args []string) {
	cmdapp.Log.Info("Starting " + appName)
	var err error

	data := ServiceData{}
	data.fc = cmdapp.NewSignalChannel()

	data.kReader, err = kafka.NewReader(data.fc)
	cmdapp.CheckOrPanic(err, "")
	defer data.kReader.Close()

	data.kWriter, err = kafka.NewWriter()
	cmdapp.CheckOrPanic(err, "")

	data.db, err = fs.NewClient()
	cmdapp.CheckOrPanic(err, "")

	data.tr, err = transcriberapi.NewClient()
	cmdapp.CheckOrPanic(err, "")

	data.filer, err = file.NewFiler()
	cmdapp.CheckOrPanic(err, "")

	err = StartServer(&data)
	cmdapp.CheckOrPanic(err, "")
	cmdapp.Log.Infof("Started")
	<-data.fc
	close(data.fc)
	cmdapp.Log.Infof("Exiting service")
}
