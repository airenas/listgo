package mocks

//go:generate pegomock generate --package=mocks --output=saver.go -m bitbucket.org/airenas/listgo/internal/pkg/status Saver

//go:generate pegomock generate --package=mocks --output=acknowledger.go -m github.com/streadway/amqp Acknowledger
