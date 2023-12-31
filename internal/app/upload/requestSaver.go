package upload

import "github.com/airenas/listgo/internal/pkg/persistence"

// RequestSaver saves the request info to db
type RequestSaver interface {
	Save(data *persistence.Request) error
}
