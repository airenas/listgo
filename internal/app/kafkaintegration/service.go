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

	if msg.ID == "" {
		cmdapp.Log.Warn("Empty kafka msg ID!")
	} else {
		err = processMsg(data, msg)
		if err != nil {
			cmdapp.Log.Error(err)
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
		return errors.Wrap(err, "Can't get audio from db")
	}

	var upload kafkaapi.UploadData
	upload.ExternalID = msg.ID
	upload.AudioData = audio.Data
	upload.JobType = audio.JobType
	upload.FileName = audio.FileName
	id, err := data.tr.Upload(&upload)
	if err != nil {
		return errors.Wrap(err, "Can't start transcription")
	}

	var idsmap kafkaapi.KafkaTrMap
	idsmap.TrID = id
	idsmap.KafkaID = msg.ID
	err = data.filer.SetWorking(&idsmap)
	if err != nil {
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
		status, err := getStatus(data, ids.TrID)
		if err != nil {
			cmdapp.Log.Error("Can't get status. Give up", err)
			data.fc <- syscall.SIGINT
			return
		}
		cmdapp.Log.Infof("Got status ID: %s, completed: %t, errorCode: %s", ids.KafkaID, status.Completed, status.ErrorCode)
		if status.Completed || status.ErrorCode != "" {
			var result kafkaapi.DBResultEntry
			result.ID = ids.KafkaID
			var msg kafkaapi.ResponseMsg
			msg.ID = ids.KafkaID

			if status.Completed {
				res, err := getResult(data, ids.TrID)
				if err != nil {
					// what do we do now? completed but no result!
					cmdapp.Log.Error("Can't get result\nMarking request as failed!", err)
					result.Status = "failed"
					result.Err.Code = status.ErrorCode
					result.Err.Error = status.Error

					msg.Error.Status = status.ErrorCode
					msg.Error.Msg = status.Error

				} else {
					result.Status = "done"
					result.Transcription.Text = status.Text
					result.Transcription.ResultFileData = res.FileData
				}
			} else {
				result.Status = "failed"
				result.Err.Code = status.ErrorCode
				result.Err.Error = status.Error

				msg.Error.Status = status.ErrorCode
				msg.Error.Msg = status.Error
			}

			err = saveSendResults(data, &result, &msg)
			if err != nil {
				cmdapp.Log.Error("Can't send results. Give up", err)
				data.fc <- syscall.SIGINT
			}

			err = data.filer.Delete(ids.TrID)
			if err != nil {
				cmdapp.Log.Error("Can't mark as finished. Give up", err)
				data.fc <- syscall.SIGINT
			}
			return
		}
	}
}

func getStatus(data *ServiceData, ID string) (*kafkaapi.Status, error) {
	var res *kafkaapi.Status
	op := func() error {
		var err error
		res, err = data.tr.GetStatus(ID)
		return err
	}
	err := backoff.Retry(op, backoff.NewExponentialBackOff())
	return res, err
}

func getResult(data *ServiceData, ID string) (*kafkaapi.Result, error) {
	var res *kafkaapi.Result
	op := func() error {
		var err error
		res, err = data.tr.GetResult(ID)
		return err
	}
	err := backoff.Retry(op, backoff.NewExponentialBackOff())
	return res, err
}

func saveSendResults(data *ServiceData, result *kafkaapi.DBResultEntry, msg *kafkaapi.ResponseMsg) error {
	op := func() error {
		err := data.db.SaveResult(result)
		if err != nil {
			cmdapp.Log.Error(err)
			return err
		}
		err = data.kWriter.Write(msg)
		if err != nil {
			cmdapp.Log.Error(err)
			return err
		}
		return nil
	}

	return backoff.Retry(op, backoff.NewExponentialBackOff())
}
