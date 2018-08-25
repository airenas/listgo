package mocks

import (
	"testing"

	"github.com/petergtz/pegomock"
	"github.com/smartystreets/goconvey/convey"
)

//go:generate pegomock generate --package=mocks --output=saver.go -m bitbucket.org/airenas/listgo/internal/pkg/status Saver

//go:generate pegomock generate --package=mocks --output=acknowledger.go -m github.com/streadway/amqp Acknowledger

//AttachMockToConvey register pegomock verification to be passed to Convey
func AttachMockToConvey(t *testing.T) {
	pegomock.RegisterMockFailHandler(handleByConvey(t))
}

func handleByConvey(t *testing.T) pegomock.FailHandler {
	return func(message string, callerSkip ...int) {
		convey.So(message, convey.ShouldBeEmpty)
	}
}
