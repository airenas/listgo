package cmdworker

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/config"
	"bitbucket.org/airenas/listgo/internal/pkg/rabbit"
	"bitbucket.org/airenas/listgo/internal/pkg/tasks"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var appName = "LiST Worker Service"

var rootCmd = &cobra.Command{
	Use:   "cmdWorkerService",
	Short: appName,
	Long:  `Worker service listens for the work event from the queue and invokes configured command to do the work`,
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
	err := validateConfig()
	if err != nil {
		panic(err)
	}
	data := ServiceData{}

	data.RecInfoLoader, err = config.NewFileRecognizerInfoLoader(cmdapp.Config.GetString("recognizerConfig.path"))
	cmdapp.CheckOrPanic(err, "Can't init recognizer info loader config (Did you provide correct setting 'recognizerConfig.path'?)")

	msgChannelProvider, err := rabbit.NewChannelProvider()
	cmdapp.CheckOrPanic(err, "Can't init rabbit channel provider")
	defer msgChannelProvider.Close()

	data.MessageSender = rabbit.NewSender(msgChannelProvider)

	ch, err := msgChannelProvider.Channel()
	cmdapp.CheckOrPanic(err, "Can't open channel")
	err = ch.Qos(1, 0, false)
	cmdapp.CheckOrPanic(err, "Can't set Qos")

	data.TaskName = cmdapp.Config.GetString("worker.taskName")

	data.WorkCh, err = rabbit.NewChannel(ch, msgChannelProvider.QueueName(data.TaskName))
	cmdapp.CheckOrPanic(err, "Can't listen "+data.TaskName+" channel")

	data.Command = cmdapp.Config.GetString("worker.command")
	data.WorkingDir = cmdapp.Config.GetString("worker.workingDir")
	data.ResultFile = cmdapp.Config.GetString("worker.resultFile")
	data.LogFile = cmdapp.Config.GetString("worker.logFile")
	data.ReadFunc = ReadFile

	data.PreloadManager, err = initPreloadManager()
	cmdapp.CheckOrPanic(err, "Can't init preload task manager")
	defer data.PreloadManager.Close()

	fc, err := StartWorkerService(&data)
	cmdapp.CheckOrPanic(err, "Can't start service")

	<-fc
	cmdapp.Log.Infof("Exiting service")
}

func validateConfig() error {
	if cmdapp.Config.GetString("worker.taskName") == "" {
		return errors.New("No worker.taskName configured")
	}
	if cmdapp.Config.GetString("worker.command") == "" {
		return errors.New("No worker.command configured")
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////////
// init prepload task manager
///////////////////////////////////////////////////////////////////////////////////
func initPreloadManager() (PreloadTaskManager, error) {
	kp := cmdapp.Config.GetString("worker.preloadKeyPrefix")
	if kp == "" {
		return &fakePreloadManager{}, nil
	}
	return tasks.NewManager(kp, cmdapp.Config.GetString("worker.workingDir"))
}

type fakePreloadManager struct{}

func (pm *fakePreloadManager) EnsureRunning(map[string]string) error {
	return nil
}

func (pm *fakePreloadManager) Close() error {
	return nil
}

///////////////////////////////////////////////////////////////////////////////////
