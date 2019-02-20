package api

import (
	"io"
	"os"
)

// File interface
type File interface {
	io.ReadCloser
	io.Seeker
	Stat() (os.FileInfo, error)
}
