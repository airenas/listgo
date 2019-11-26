package upload

import "bitbucket.org/airenas/listgo/internal/app/upload/api"

// RecognizerMap provides the recognizer ID by key
type RecognizerMap interface {
	Get(key string) (string, error)
}

// RecognizerProvider provides available recognizers list
type RecognizerProvider interface {
	GetAll() ([]*api.Recognizer, error)
}
