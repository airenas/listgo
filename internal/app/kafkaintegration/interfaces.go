package kafkaintegration;

import (
	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
)

type kafkaReader interface {
	Get() (*kafkaapi.Msg, error)
	Commit(*kafkaapi.Msg) error
	Close()
}

type kafkaWriter interface {
	Write(msg *kafkaapi.ResponseMsg) error
}

type filer interface {
	Find(kafkaID string) (*kafkaapi.KafkaTrMap, error)
	SetWorking(krIds* kafkaapi.KafkaTrMap) (error)
	Delete(kafkaID string) (error)
}

type db interface {
	GetAudio(kafkaID string) (*kafkaapi.DBEntry, error)
	SaveResult(data* kafkaapi.DBResultEntry) (error)
}

type transcriber interface {
	Upload(audio *kafkaapi.UploadData) (string, error)
	GetStatus(ID string) (*kafkaapi.Status, error)
	GetResult(ID string) (*kafkaapi.Result, error)
}


