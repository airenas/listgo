package loader

import (
	"errors"
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"

	"bitbucket.org/airenas/listgo/internal/app/result/api"
	. "github.com/smartystreets/goconvey/convey"
)

func TestLoads(t *testing.T) {
	Convey("Given a loader", t, func() {
		fakeFile := fakeFile("content")
		fileLoader := LocalFileLoader{Path: "/data/",
			OpenFileFunc: func(file string) (api.File, error) {
				return fakeFile, nil
			}}
		Convey("When file is loaded", func() {
			f, err := fileLoader.Load("file")
			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})
			Convey("File is returned", func() {
				So(f, ShouldNotBeNil)
			})
		})
	})
}

func TestFailsOnNoOpen(t *testing.T) {
	Convey("Given a loader", t, func() {
		fileLoader := LocalFileLoader{Path: "",
			OpenFileFunc: func(file string) (api.File, error) {
				return nil, errors.New("olia")
			}}
		Convey("When error is returned on open", func() {
			_, err := fileLoader.Load("file")
			Convey("Then error is returned", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestChecksDirOnInit(t *testing.T) {
	Convey("Given non empty input", t, func() {
		_, err := NewLocalFileLoader("./")
		Convey("Then no error is returned", func() {
			So(err, ShouldBeNil)
		})
	})
	Convey("Given empty input", t, func() {
		_, err := NewLocalFileLoader("")
		Convey("Then error is returned", func() {
			So(err, ShouldNotBeNil)
		})
	})
}

func fakeFile(c string) api.File {
	return mocks.NewMockFile()
}
