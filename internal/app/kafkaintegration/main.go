package kafkaintegration

import (
	"bitbucket.org/airenas/listgo/internal/pkg/fs"
	"bitbucket.org/airenas/listgo/internal/pkg/kafka"

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
	data.parallelWorkSemaphore, _ = newJobsSemaphore()

	data.kReader, err = kafka.NewReader(data.fc)
	cmdapp.CheckOrPanic(err, "")
	defer data.kReader.Close()

	data.kWriter, err = kafka.NewWriter()
	cmdapp.CheckOrPanic(err, "")

	data.db, err = fs.NewClient()
	cmdapp.CheckOrPanic(err, "")

	err = StartServer(&data)
	cmdapp.CheckOrPanic(err, "")
	cmdapp.Log.Infof("Started")
	<-data.fc
	cmdapp.Log.Infof("Exiting service")
}

func newJobsSemaphore() (chan struct{}, error) {
	jobs := cmdapp.Config.GetInt("jobs")
	if jobs <= 0 {
		jobs = 1
	}
	cmdapp.Log.Infof("Job count = %d", jobs)
	res := make(chan struct{}, jobs)
	res <- struct{}{}
	return res, nil
}
