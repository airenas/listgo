package clean

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
)

type localFile struct {
	StoragePath string
	pattern     string
}

func newLocalFile(storagePath string, pattern string) (*localFile, error) {
	cmdapp.Log.Infof("Init Local File Storage Clean at: %s/%s", storagePath, pattern)
	if pattern == "" {
		return nil, errors.New("No pattern provided")
	}
	if !strings.Contains(pattern, "{ID}") {
		return nil, errors.New("Pattern does not contain {ID}")
	}
	sP := ""
	if !strings.HasPrefix(pattern, "/") {
		if storagePath == "" {
			return nil, errors.New("No storage path provided")
		}
		sP = storagePath
	}
	f := localFile{StoragePath: sP, pattern: pattern}
	return &f, nil
}

func (fs *localFile) Clean(ID string) error {
	fp := fs.getPath(ID)
	cmdapp.Log.Infof("Removing %s", fp)
	return remove(fp)
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

func (fs *localFile) getPath(ID string) string {
	res := strings.ReplaceAll(fs.pattern, "{ID}", ID)
	if fs.StoragePath != "" {
		res = path.Join(fs.StoragePath, res)
	}
	return res
}
