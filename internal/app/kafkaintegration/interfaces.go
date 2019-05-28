package kafkaintegration;

import (
	"bitbucket.org/airenas/listgo/internal/app/kafkaintegration/kafkaapi"
)

type kafkaReader interface {
	Get() (*kafkaapi.Msg, error)
	Commit(*kafkaapi.Msg) error
	Close()
}

