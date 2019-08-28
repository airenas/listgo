package clean

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
)

type localFile struct {
	StoragePath string
	pattern     string
}

func newLocalFile(storagePath string, pattern string) (*localFile, error) {
	cmdapp.Log.Infof("Init Local File Storage Clean at: %s/%s", storagePath, pattern)
	if storagePath == "" {
		return nil, errors.New("No storage path provided")
	}
	if pattern == "" {
		return nil, errors.New("No pattern provided")
	}
	if !strings.Contains(pattern, "{ID}") {
		return nil, errors.New("Pattern does not contain {ID}")
	}
	f := localFile{StoragePath: storagePath, pattern: pattern}
	return &f, nil
}

func (fs *localFile) Clean(ID string) error {
	p := strings.ReplaceAll(fs.pattern, "{ID}", ID)
	fn := path.Join(fs.StoragePath, p)
	cmdapp.Log.Infof("Removing %s", fn)
	return remove(fn)
}

func remove(fn string) error {
	files, err := filepath.Glob(fn)
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			return err
		}
		cmdapp.Log.Infof("Removed %s", file)
	}
	return nil
}
