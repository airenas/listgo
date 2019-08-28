package clean

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/mongo"
	"github.com/spf13/cobra"
)

var appName = "LiST Data Clean Service"

var rootCmd = &cobra.Command{
	Use:   "cleanService",
	Short: appName,
	Long:  `Service to provide data deletion functionality`,
	Run:   run,
}

func init() {
	cmdapp.InitApplication(rootCmd)
	rootCmd.PersistentFlags().Int32P("port", "", 8000, "Default service port")
	cmdapp.Config.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	cmdapp.Config.SetDefault("port", 8080)
}

//Execute starts the server
func Execute() {
	cmdapp.Execute(rootCmd)
}

func run(cmd *cobra.Command, args []string) {
	cmdapp.Log.Info("Starting " + appName)

	data := &ServiceData{}

	data.Port = cmdapp.Config.GetInt("port")
	mongoSessionProvider, err := mongo.NewSessionProvider()
	cmdapp.CheckOrPanic(err, "Can't init mongo")
	defer mongoSessionProvider.Close()

	data.cleaner, err = newCleanerImpl(mongoSessionProvider, cmdapp.Config.GetString("fileStorage.path"))
	cmdapp.CheckOrPanic(err, "Can't init cleaner")

	tdata := timerServiceData{}
	tdata.runEvery = time.Hour
	tdata.cleaner = data.cleaner
	tdata.idsProvider, err = mongo.NewCleanIDsProvider(mongoSessionProvider, time.Hour*24*7)
	cmdapp.CheckOrPanic(err, "Can't init mongo expired IDs provider")
	tdata.qChan = make(chan struct{})
	tdata.workWaitChan = make(chan struct{})

	go func() {
		err = StartWebServer(data)
		cmdapp.CheckOrPanic(err, "")
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	
	err = startCleanTimer(&tdata)
	cmdapp.CheckOrPanic(err, "")

	<-sigs
	cmdapp.Log.Infof("Stopping")
	// indicate to stop and wait for job complete
	close(tdata.qChan)
	<-tdata.workWaitChan
}
