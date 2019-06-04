package kafkaintegration

import (
	"os"
	"syscall"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"

	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
)

// ServiceData keeps data required for service work
type ServiceData struct {
	fc                    chan os.Signal
	parallelWorkSemaphore chan struct{}
	kReader               kafkaReader
	kWriter               kafkaWriter
	filer                 filer
	db                    db
	tr                    transcriber
}

//StartServer init the service to listen to kafka messages and pass it to transcrption
func StartServer(data *ServiceData) error {
	err := validateData(data)
	if err != nil {
		return err
	}
	err = initPreviousWork(data)
	if err != nil {
		return errors.Wrap(err, "Can't init previous work")
	}
	go listenKafka(data)
	return nil
}

func validateData(data *ServiceData) error {
	if data.kReader == nil {
		return errors.New("No Kafka reader")
	}
	if data.kWriter == nil {
		return errors.New("No Kafka writer")
	}
	if data.db == nil {
		return errors.New("No DB set")
	}
	if data.tr == nil {
		return errors.New("No Transcriber set")
	}
	if data.filer == nil {
		return errors.New("No File helper set")
	}
	return nil
}

func initPreviousWork(data *ServiceData) error {
	wlist, err := data.filer.GetPending()
	if err != nil {
		return errors.Wrap(err, "Can't get working list")
	}
	for _, we := range wlist {
		waitForWorkAccess(data)
		go listenTranscription(data, we)
	}
	return nil
}

func listenKafka(data *ServiceData) {
	for {
		waitForWorkAccess(data)
		readProcessKafkaMsg(data)
	}
}

func waitForWorkAccess(data *ServiceData) {
	cmdapp.Log.Info("Waiting for free work slot")
	data.parallelWorkSemaphore <- struct{}{}
	cmdapp.Log.Info("Got access to work")
}

func readProcessKafkaMsg(data *ServiceData) {
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
		cmdapp.Log.Error(err)
		data.fc <- syscall.SIGINT
		return
	}
	err = data.kReader.Commit(msg)
	if err != nil {
		cmdapp.Log.Error("Can't commit kafka msg", err)
		data.fc <- syscall.SIGINT
		return
	}
}

func processMsg(data *ServiceData, msg *kafkaapi.Msg) error {
	op := func() error {
		return processMsgInt(data, msg)
	}

	return backoff.Retry(op, backoff.NewExponentialBackOff())
}

func processMsgInt(data *ServiceData, msg *kafkaapi.Msg) error {
	cmdapp.Log.Infof("Process msg: %s", msg.ID)
	ids, err := data.filer.FindWorking(msg.ID)
	if err != nil {
		return errors.Wrap(err, "Can't check ids map")
	}
	if ids != nil {
		go listenTranscription(data, ids)
		return nil
	}
	audio, err := data.db.GetAudio(msg.ID)
	if err != nil {
		//todo repeat
		return errors.Wrap(err, "Can't get audio from db")
	}
	var upload kafkaapi.UploadData
	upload.ExternalID = msg.ID
	upload.AudioData = audio.Data
	upload.JobType = audio.JobType
	upload.FileName = audio.FileName
	id, err := data.tr.Upload(&upload)
	if err != nil {
		//todo repeat
		return errors.Wrap(err, "Can't start transcription")
	}
	var idsmap kafkaapi.KafkaTrMap
	idsmap.TrID = id
	idsmap.KafkaID = msg.ID
	err = data.filer.SetWorking(&idsmap)
	if err != nil {
		//todo fail
		return errors.Wrap(err, "Can't mark as working")
	}
	go listenTranscription(data, &idsmap)
	return nil
}

func listenTranscription(data *ServiceData, ids *kafkaapi.KafkaTrMap) {
	cmdapp.Log.Infof("Waiting for transcription to complete, ID: %s", ids.KafkaID)
	defer func() { <-data.parallelWorkSemaphore }()
	for {
		time.Sleep(3 * time.Second)
		status, err := data.tr.GetStatus(ids.TrID)
		if err != nil {
			//todo repeat
			cmdapp.Log.Error(err)
			break
		}
		cmdapp.Log.Infof("Got status ID: %s, completed: %t, errorCode: %s", ids.KafkaID, status.Completed, status.ErrorCode)
		if status.ErrorCode != "" {
			//todo send error
			var result kafkaapi.DBResultEntry
			result.ID = ids.KafkaID
			result.Status = "failed"
			result.Err.Code = status.ErrorCode
			result.Err.Error = status.Error
			err = data.db.SaveResult(&result)
			if err != nil {
				//todo repeat
				cmdapp.Log.Error(err)
				break
			}
			var msg kafkaapi.ResponseMsg
			msg.ID = ids.KafkaID
			msg.Error.Status = status.ErrorCode
			msg.Error.Msg = status.Error
			err = data.kWriter.Write(&msg)
			if err != nil {
				//exit service
				cmdapp.Log.Error(err)
				break
			}
			err = data.filer.Delete(ids.TrID)
			if err != nil {
				//exit service?
				cmdapp.Log.Error(err)
				break
			}
			break
		}
		if status.Completed {
			res, err := data.tr.GetResult(ids.TrID)
			if err != nil {
				//todo repeat
				cmdapp.Log.Error(err)
				break
			}
			var result kafkaapi.DBResultEntry
			result.ID = ids.KafkaID
			result.Status = "done"
			result.Transcription.Text = status.Text
			result.Transcription.ResultFileData = res.FileData
			err = data.db.SaveResult(&result)
			if err != nil {
				//todo repeat
				cmdapp.Log.Error(err)
				break
			}
			var msg kafkaapi.ResponseMsg
			msg.ID = ids.KafkaID
			err = data.kWriter.Write(&msg)
			if err != nil {
				//exit service
				cmdapp.Log.Error(err)
				break
			}
			err = data.filer.Delete(ids.TrID)
			if err != nil {
				//exit service?
				cmdapp.Log.Error(err)
				break
			}
			break
		}
	}
}

func sendErrorMsg(data *ServiceData, msg *kafkaapi.ResponseMsg) error {
	cmdapp.Log.Infof("Sending error kafka msg: %s\n\t%s", msg.ID, msg.Error.Msg)
	err := data.kWriter.Write(msg)
	if err != nil {
		return errors.Wrap(err, "Can't send kafka msg")
	}
	return nil
}
