package saver

import (
	"errors"
	"io"
	"os"
	"strconv"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
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
func NewLocalFileSaver(storagePath string) (*LocalFileSaver, error) {
	cmdapp.Log.Infof("Init Local File Storage at: %s", storagePath)
	if storagePath == "" {
		return nil, errors.New("No storage path provided")
	}
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		cmdapp.Log.Infof("Trying to create storage directory at: %s", storagePath)
		err = os.MkdirAll(storagePath, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	f := LocalFileSaver{StoragePath: storagePath, OpenFileFunc: openFile}
	return &f, nil
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
	cmdapp.Log.Infof("Saved file %s. Size = %d b", fileName, strconv.FormatInt(savedBytes, 10))
	return nil
}

func openFile(fileName string) (WriterCloser, error) {
	return os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
}
