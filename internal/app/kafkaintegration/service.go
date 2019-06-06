package kafkaintegration

import (
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"

	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"
)

// ServiceData keeps data required for service work
type ServiceData struct {
	fc      *utils.MultiCloseChannel
	kReader KafkaReader
	kWriter KafkaWriter
	filer   Filer
	db      DB
	tr      Transcriber
	bp      backoffProvider
}

//StartServer init the service to listen to kafka messages and pass it to transcrption
func StartServer(data *ServiceData) error {
	err := validateData(data)
	if err != nil {
		return err
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
	if data.bp == nil {
		return errors.New("No BackOff provider set")
	}
	return nil
}

func listenKafka(data *ServiceData) {
	var err error
	for err == nil {
		err := readProcessKafkaMsg(data)
		if err != nil {
			cmdapp.Log.Error(err)
			data.fc.Close()
			return
		}
	}
}

func readProcessKafkaMsg(data *ServiceData) error {
	cmdapp.Log.Info("Waiting for kafka msg")
	msg, err := data.kReader.Get()
	if err != nil {
		return errors.Wrap(err, "Can't read kafka msg")
	}
	cmdapp.Log.Infof("Got kafka msg %s", msg.ID)

	if msg.ID == "" {
		cmdapp.Log.Warn("Empty kafka msg ID!")
	} else {
		err = processMsg(data, msg)
		if err != nil {
			return err
		}
	}
	err = data.kReader.Commit(msg)
	if err != nil {
		return errors.Wrap(err, "Can't commit kafka msg")
	}
	return nil
}

func processMsg(data *ServiceData, msg *kafkaapi.Msg) error {
	cmdapp.Log.Infof("Process msg: %s", msg.ID)
	ids, err := data.filer.Find(msg.ID)
	if err != nil {
		return errors.Wrap(err, "Can't check ids map")
	}
	if ids == nil {
		audio, err := getAudio(data, msg.ID)
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

		ids = &kafkaapi.KafkaTrMap{}
		ids.TrID = id
		ids.KafkaID = msg.ID
		err = data.filer.SetWorking(ids)
		if err != nil {
			return errors.Wrap(err, "Can't mark as working")
		}
	}
	return listenTranscription(data, ids)
}

func listenTranscription(data *ServiceData, ids *kafkaapi.KafkaTrMap) error {
	cmdapp.Log.Infof("Waiting for transcription to complete, ID: %s", ids.KafkaID)
	for {
		time.Sleep(3 * time.Second)
		status, err := getStatus(data, ids.TrID)
		if err != nil {
			return errors.Wrap(err, "Can't get status. Give up")
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
					result.Status = kafkaapi.DBStatusFailed
					result.Err.Code = status.ErrorCode
					result.Err.Error = status.Error

					msg.Error.Status = status.ErrorCode
					msg.Error.Msg = status.Error

				} else {
					result.Status = kafkaapi.DBStatusDone
					result.Transcription.Text = status.Text
					result.Transcription.ResultFileData = res.FileData
				}
			} else {
				result.Status = kafkaapi.DBStatusFailed
				result.Err.Code = status.ErrorCode
				result.Err.Error = status.Error

				msg.Error.Status = status.ErrorCode
				msg.Error.Msg = status.Error
			}

			err = saveSendResults(data, &result, &msg)
			if err != nil {
				return errors.Wrap(err, "Can't send results. Give up")
			}

			err = data.filer.Delete(ids.KafkaID)
			if err != nil {
				return errors.Wrap(err, "Can't mark as finished. Give up")
			}
			return nil
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
	err := backoff.Retry(op, data.bp.Get())
	return res, err
}

func getResult(data *ServiceData, ID string) (*kafkaapi.Result, error) {
	var res *kafkaapi.Result
	op := func() error {
		var err error
		res, err = data.tr.GetResult(ID)
		return err
	}
	err := backoff.Retry(op, data.bp.Get())
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
	return backoff.Retry(op, data.bp.Get())
}

func getAudio(data *ServiceData, ID string) (*kafkaapi.DBEntry, error) {
	var res *kafkaapi.DBEntry
	op := func() error {
		var err error
		res, err = data.db.GetAudio(ID)
		return err
	}
	err := backoff.Retry(op, data.bp.Get())
	return res, err
}
