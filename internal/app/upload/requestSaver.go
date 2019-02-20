package upload

import "bitbucket.org/airenas/listgo/internal/app/upload/api"

// RequestSaver saves the request info to db
type RequestSaver interface {
	Save(data api.RequestData) error
}
