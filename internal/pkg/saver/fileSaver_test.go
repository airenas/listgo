package saver

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSaves(t *testing.T) {
	Convey("Given an input with body", t, func() {
		fakeFile := fakeWriterCloser{bytes.NewBufferString(""), "", false}
		fileSaver := LocalFileSaver{StoragePath: "/data/",
			OpenFileFunc: func(file string) (WriterCloser, error) {
				fakeFile.Name = file
				return &fakeFile, nil
			}}
		Convey("When file is saved", func() {
			err := fileSaver.Save("file", strings.NewReader("body"))
			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the file content should be body", func() {
				So(fakeFile.String(), ShouldEqual, "body")
			})
			Convey("Then the file name should be /data/file", func() {
				So(fakeFile.Name, ShouldEqual, "/data/file")
			})
			Convey("Then the file should be closed", func() {
				So(fakeFile.Closed, ShouldBeTrue)
			})
		})
	})
}

func TestFailsOnNoOpen(t *testing.T) {
	Convey("Given an input", t, func() {
		fakeFile := fakeWriterCloser{bytes.NewBufferString(""), "", false}
		fileSaver := LocalFileSaver{StoragePath: "",
			OpenFileFunc: func(file string) (WriterCloser, error) {
				return &fakeFile, errors.New("olia")
			}}
		Convey("When error is returned on open", func() {
			err := fileSaver.Save("file", strings.NewReader("body"))
			Convey("Then error is returned", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestChecksDirOnInit(t *testing.T) {
	Convey("Given non empty input", t, func() {
		_, err := NewLocalFileSaver("./")
		Convey("Then no error is returned", func() {
			So(err, ShouldBeNil)
		})
	})
	Convey("Given empty input", t, func() {
		_, err := NewLocalFileSaver("")
		Convey("Then error is returned", func() {
			So(err, ShouldNotBeNil)
		})
	})
}

type fakeWriterCloser struct {
	*bytes.Buffer
	Name   string
	Closed bool
}

func (t *fakeWriterCloser) Close() error {
	t.Closed = true
	return nil
}
