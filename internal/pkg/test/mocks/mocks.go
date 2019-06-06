package mocks

import (
	"testing"

	"github.com/petergtz/pegomock"
)

//go:generate pegomock generate --package=mocks --output=saver.go -m bitbucket.org/airenas/listgo/internal/pkg/status Saver

//go:generate pegomock generate --package=mocks --output=acknowledger.go -m github.com/streadway/amqp Acknowledger

//go:generate pegomock generate --package=mocks --output=resultSaver.go -m bitbucket.org/airenas/listgo/internal/app/manager ResultSaver

//go:generate pegomock generate --package=mocks --output=publisher.go -m bitbucket.org/airenas/listgo/internal/pkg/messages Publisher

//go:generate pegomock generate --package=mocks --output=messageSender.go -m bitbucket.org/airenas/listgo/internal/pkg/messages Sender

//go:generate pegomock generate --package=mocks --output=wsConn.go -m bitbucket.org/airenas/listgo/internal/app/status WsConn

//go:generate pegomock generate --package=mocks --output=statusProvider.go -m bitbucket.org/airenas/listgo/internal/app/status Provider

//go:generate pegomock generate --package=mocks --output=requestSaver.go -m bitbucket.org/airenas/listgo/internal/app/upload RequestSaver

//go:generate pegomock generate --package=mocks --output=emailMaker.go -m bitbucket.org/airenas/listgo/internal/app/inform EmailMaker

//go:generate pegomock generate --package=mocks --output=emailRetriever.go -m bitbucket.org/airenas/listgo/internal/app/inform EmailRetriever

//go:generate pegomock generate --package=mocks --output=locker.go -m bitbucket.org/airenas/listgo/internal/app/inform Locker

//go:generate pegomock generate --package=mocks --output=file.go -m bitbucket.org/airenas/listgo/internal/app/result/api File

//go:generate pegomock generate --package=mocks --output=fileLoader.go -m bitbucket.org/airenas/listgo/internal/app/result FileLoader

//go:generate pegomock generate --package=mocks --output=fileNameProvider.go -m bitbucket.org/airenas/listgo/internal/app/result FileNameProvider

//go:generate pegomock generate --package=mocks --output=kReader.go -m bitbucket.org/airenas/listgo/internal/app/kafkaintegration KafkaReader

//go:generate pegomock generate --package=mocks --output=kWriter.go -m bitbucket.org/airenas/listgo/internal/app/kafkaintegration KafkaWriter

//go:generate pegomock generate --package=mocks --output=db.go -m bitbucket.org/airenas/listgo/internal/app/kafkaintegration DB

//go:generate pegomock generate --package=mocks --output=filer.go -m bitbucket.org/airenas/listgo/internal/app/kafkaintegration Filer

//go:generate pegomock generate --package=mocks --output=transcriber.go -m bitbucket.org/airenas/listgo/internal/app/kafkaintegration Transcriber

//AttachMockToTest register pegomock verification to be passed to testing engine
func AttachMockToTest(t *testing.T) {
	pegomock.RegisterMockFailHandler(handleByTest(t))
}

func handleByTest(t *testing.T) pegomock.FailHandler {
	return func(message string, callerSkip ...int) {
		if message != "" {
			t.Error(message)
		}
	}
}
