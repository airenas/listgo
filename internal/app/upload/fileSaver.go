package upload

import (
	"io"
)

// FileSaver saves the file with the provided name
type FileSaver interface {
	Save(name string, reader io.Reader) error
}
