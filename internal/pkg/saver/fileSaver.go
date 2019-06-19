package saver

import (
	"io"
	"os"
	"strconv"
	"syscall"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
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
	cmdapp.Log.Infof("Saved file %s. Size = %s b", fileName, strconv.FormatInt(savedBytes, 10))
	return nil
}

func openFile(fileName string) (WriterCloser, error) {
	return os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
}

//HealthyFunc returns func for health check
func (fs *LocalFileSaver) HealthyFunc(sizeInMb uint64) func() error {
	return func() error {
		var info syscall.Statfs_t
		err := syscall.Statfs(fs.StoragePath, &info)
		if err != nil {
			return errors.Errorf("Can't get info for dir: %s", fs.StoragePath)
		}

		mb := info.Bavail * uint64(info.Bsize) / 1024 / 1024

		if mb < sizeInMb {
			return errors.Errorf("Disk space is %d mb, threshold: %d mb, location: %s", mb, sizeInMb, fs.StoragePath)
		}
		return nil
	}
}
