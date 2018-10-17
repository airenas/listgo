package mocks

import (
	"testing"

	"github.com/petergtz/pegomock"
	"github.com/smartystreets/goconvey/convey"
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

//AttachMockToConvey register pegomock verification to be passed to Convey
func AttachMockToConvey(t *testing.T) {
	pegomock.RegisterMockFailHandler(handleByConvey(t))
}

func handleByConvey(t *testing.T) pegomock.FailHandler {
	return func(message string, callerSkip ...int) {
		convey.So(message, convey.ShouldBeEmpty)
	}
}
