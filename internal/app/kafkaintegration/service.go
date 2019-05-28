package kafkaintegration

import (
	"os"
	"syscall"

	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
)

// ServiceData keeps data required for service work
type ServiceData struct {
	fc                    chan os.Signal
	parallelWorkSemaphore chan struct{}
	kReader               kafkaReader
}

//StartServer init the service to listen to kafka messages and pass it to transcrption
func StartServer(data *ServiceData) (<-chan os.Signal, error) {
	go listenKafka(data)
	go listenTranscription(data)

	return data.fc, nil
}

func listenKafka(data *ServiceData) {
	for {
		cmdapp.Log.Info("Waiting for kafka msg")
		msg, err := data.kReader.Get()
		if err != nil {
			cmdapp.Log.Error("Can't read kafka msg", err)
			data.fc <- syscall.SIGINT
			return
		}
		cmdapp.Log.Infof("Got kafka msg %s", msg.ID)
		err = processMsg(data, msg)
		if err != nil {
			err = sendErrorMsg(data, msg, err)
			if err != nil {
				cmdapp.Log.Error("Can't send kafka msg", err)
				data.fc <- syscall.SIGINT
				return
			}
		}
		err = data.kReader.Commit(msg)
		if err != nil {
			cmdapp.Log.Error("Can't commit kafka msg", err)
			data.fc <- syscall.SIGINT
			return
		}
		cmdapp.Log.Info("Waiting for free work slot")
		data.parallelWorkSemaphore <- struct{}{}
		cmdapp.Log.Info("Got access to process kafka messages")
	}
}

func listenTranscription(data *ServiceData) {
	//cmdapp.Log.Infof("Stopped listening queue")
	//data.fc <- syscall.SIGINT
}

func processMsg(data *ServiceData, msg *kafkaapi.Msg) error {
	cmdapp.Log.Infof("Process msg: %s", msg.ID)
	return nil
}

func sendErrorMsg(data *ServiceData, msg *kafkaapi.Msg, err error) error {
	cmdapp.Log.Infof("Sending error kafka msg: %s\n\t%s", msg.ID, err.Error())
	return nil
}
