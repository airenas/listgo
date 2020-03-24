package kafkaintegration

import (
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"

	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	errc "bitbucket.org/airenas/listgo/internal/pkg/err"
	"bitbucket.org/airenas/listgo/internal/pkg/utils"
)

// ServiceData keeps data required for service work
type ServiceData struct {
	fc          *utils.MultiCloseChannel
	kReader     KafkaReader
	kWriter     KafkaWriter
	filer       Filer
	db          DB
	tr          Transcriber
	bp          backoffProvider
	statusSleep time.Duration
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

//processMsg tries to process message, returns error if no commit is needed
func processMsg(data *ServiceData, msg *kafkaapi.Msg) error {
	cmdapp.Log.Infof("Process msg: %s", msg.ID)
	ids, err := data.filer.Find(msg.ID)
	if err != nil {
		return errors.Wrap(err, "Can't check ids map")
	}
	if ids == nil {
		audio, err := getAudio(data, msg.ID)
		if err != nil {
			return sendKafkaMsg(data, msg.ID, errors.Wrap(err, "Can't get audio from db").Error())
		}

		upReq := kafkaapi.UploadData{ExternalID: msg.ID, AudioData: audio.Data, JobType: audio.JobType,
			FileName: audio.FileName, NumberOfSpeakers: audio.NumberOfSpeakers,
			RecordQuality: audio.RecordQuality}
		id, err := upload(data, &upReq)
		if err != nil {
			cmdapp.Log.Error(err)
			return saveSendResults(data, &kafkaapi.DBResultEntry{ID: msg.ID, Status: kafkaapi.DBStatusFailed,
				Err: kafkaapi.DBTranscriptionError{Code: errc.DefaultCode,
					Error: errors.Wrap(err, "Can't start transcription").Error()}})
		}

		ids = &kafkaapi.KafkaTrMap{TrID: id, KafkaID: msg.ID}
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
		time.Sleep(data.statusSleep)
		status, err := getStatus(data, ids.TrID)
		if err != nil {
			return saveSendResults(data, &kafkaapi.DBResultEntry{ID: ids.KafkaID, Status: kafkaapi.DBStatusFailed,
				Err: kafkaapi.DBTranscriptionError{Code: errc.DefaultCode,
					Error: errors.Wrap(err, "Can't get status").Error()}})
		}
		cmdapp.Log.Infof("Got status ID: %s, completed: %t, errorCode: %s", ids.KafkaID, status.Completed, status.ErrorCode)
		if status.Completed || status.ErrorCode != "" {
			var result kafkaapi.DBResultEntry
			result.ID = ids.KafkaID

			if status.Completed {
				res, err := getResult(data, ids.TrID)
				if err != nil {
					// what do we do now? completed but no result!
					err = errors.Wrap(err, "Can't get result\nMarking request as failed!")
					cmdapp.Log.Error(err)
					result.Status = kafkaapi.DBStatusFailed
					result.Err.Code = errc.DefaultCode
					result.Err.Error = err.Error()
				} else {
					result.Status = kafkaapi.DBStatusDone
					result.Transcription.Text = status.Text
					result.Transcription.ResultFileData = res.FileData
				}
			} else {
				result.Status = kafkaapi.DBStatusFailed
				result.Err.Code = status.ErrorCode
				result.Err.Error = status.Error
			}

			err = saveSendResults(data, &result)
			if err != nil {
				return errors.Wrap(err, "Can't send results. Give up")
			}

			err = data.filer.Delete(ids.KafkaID)
			if err != nil {
				return errors.Wrap(err, "Can't mark as finished. Give up")
			}

			err = data.tr.Delete(ids.TrID)
			if err != nil {
				cmdapp.Log.Warning(errors.Wrap(err, "Can't invoke data cleaner"))
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

func saveSendResults(data *ServiceData, result *kafkaapi.DBResultEntry) error {
	op := func() error {
		err := data.db.SaveResult(result)
		if err != nil {
			return err
		}
		return nil
	}
	err := backoff.Retry(op, data.bp.Get())
	if err != nil {
		return sendKafkaMsg(data, result.ID,
			errors.Wrap(err, "Can't post transcription result to file storage").Error())
	}
	return sendKafkaMsg(data, result.ID, "")
}

func sendKafkaMsg(data *ServiceData, id string, dmsg string) error {
	msg := &kafkaapi.ResponseMsg{ID: id}
	if dmsg != "" {
		msg.Error.Code = errc.DefaultCode
		msg.Error.DebugMessage = dmsg
	}
	op := func() error {
		err := data.kWriter.Write(msg)
		if err != nil {
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

func upload(data *ServiceData, upReq *kafkaapi.UploadData) (string, error) {
	var res string
	op := func() error {
		var err error
		res, err = data.tr.Upload(upReq)
		if errors.Is(err, utils.ErrWrongHTTPCall) {
			return backoff.Permanent(err)
		}
		return err
	}
	err := backoff.Retry(op, data.bp.Get())
	return res, err
}
