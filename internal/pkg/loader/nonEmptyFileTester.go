package loader

import (
	"bufio"
	"io"
	"os"
	"strings"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
)

type fileOpenFunc func(string) (io.ReadCloser, error)

// NonEmptyFileTester test if file is empty
type NonEmptyFileTester struct {
	Path     string
	fileFunc fileOpenFunc
}

//NewNonEmptyFileTester creates NonEmptyFileTester instance
func NewNonEmptyFileTester(path string) (*NonEmptyFileTester, error) {
	cmdapp.Log.Infof("Init FileTester with pattern: %s", path)
	if path == "" {
		return nil, errors.New("No path provided")
	}
	if !strings.Contains(path, "{ID}") {
		return nil, errors.New("Path pattern does not contain '{ID}'")
	}
	f := NonEmptyFileTester{Path: path, fileFunc: openFileF}
	return &f, nil
}

// Test tests if file is not empty
func (fs *NonEmptyFileTester) Test(id string) (bool, error) {
	realFile := strings.Replace(fs.Path, "{ID}", id, -1)
	cmdapp.Log.Infof("Reading file: %s", realFile)
	f, err := fs.fileFunc(realFile)
	if err != nil {
		return false, errors.Wrap(err, "Can not open file "+realFile)
	}
	defer f.Close()
	rd := bufio.NewReader(f)
	var bytes []byte
	for err == nil {
		bytes, _, err = rd.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, errors.Wrap(err, "Can read file "+realFile)
		}
		if len(strings.TrimSpace(string(bytes))) > 0 {
			return true, nil
		}
	}
	return false, nil

}

func openFileF(name string) (io.ReadCloser, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, errors.Wrap(err, "Can not open file "+name)
	}
	return f, nil
}
