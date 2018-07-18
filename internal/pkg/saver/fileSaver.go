package saver

import (
	"errors"
	"io"
	"log"
	"os"
	"strconv"
)

//WriterCloser keeps Writer interface and close function
type WriterCloser interface {
	io.Writer
	Close() error
}

//OpenFileFunc declares function to open file by name and return Writer
type OpenFileFunc func(fileName string) (WriterCloser, error)

// LocalFileSaver saves file on local disk
type LocalFileSaver struct {
	// StoragePath is the main folder to save into
	StoragePath  string
	OpenFileFunc OpenFileFunc
}

//NewLocalFileSaver creates LocalFileSaver instance
func NewLocalFileSaver(storagePath string) *LocalFileSaver {
	f := LocalFileSaver{StoragePath: storagePath, OpenFileFunc: openFile}
	return &f
}

// Save saves file to disk
func (fs LocalFileSaver) Save(name string, reader io.Reader) error {
	fileName := fs.StoragePath + name
	f, err := fs.OpenFileFunc(fileName)
	if err != nil {
		return errors.New("Can not create file " + fileName + ". " + err.Error())
	}
	defer f.Close()
	savedBytes, err := io.Copy(f, reader)
	if err != nil {
		return errors.New("Can not save file " + fileName + ". " + err.Error())
	}
	log.Println("Saved file " + fileName + ". Size = " + strconv.FormatInt(savedBytes, 10))
	return nil
}

func openFile(fileName string) (WriterCloser, error) {
	return os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
}
