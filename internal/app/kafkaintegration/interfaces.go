package kafkaintegration;

import (
	"github.com/cenkalti/backoff"
	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
)

//KafkaReader provides messages from Kafka
type KafkaReader interface {
	Get() (*kafkaapi.Msg, error)
	Commit(*kafkaapi.Msg) error
	Close()
}

//KafkaWriter writes msgs to kafka
type KafkaWriter interface {
	Write(msg *kafkaapi.ResponseMsg) error
}

//Filer helps persist working IDs
type Filer interface {
	Find(kafkaID string) (*kafkaapi.KafkaTrMap, error)
	SetWorking(krIds* kafkaapi.KafkaTrMap) (error)
	Delete(kafkaID string) (error)
}

//DB loads writes data to AFT Storage
type DB interface {
	GetAudio(kafkaID string) (*kafkaapi.DBEntry, error)
	SaveResult(data* kafkaapi.DBResultEntry) (error)
}

//Transcriber comunicates with transcription service
type Transcriber interface {
	Upload(audio *kafkaapi.UploadData) (string, error)
	GetStatus(ID string) (*kafkaapi.Status, error)
	GetResult(ID string) (*kafkaapi.Result, error)
	Delete(ID string) (error)
}

type backoffProvider interface {
	Get() backoff.BackOff
}


