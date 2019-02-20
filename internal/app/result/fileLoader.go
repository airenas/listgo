package result

import (
	"bitbucket.org/airenas/listgo/internal/app/result/api"
)

// FileLoader loads file by the name
type FileLoader interface {
	Load(name string) (api.File, error)
}
