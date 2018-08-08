package cmdworker

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRun_NoParameter_Fail(t *testing.T) {
	Convey("Given a command", t, func() {
		cmd := "ls"
		Convey("When the command is executed", func() {
			err := RunCommand(cmd, "/", "id")
			Convey("Then the result should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestRun_WrongParameter_Fail(t *testing.T) {
	Convey("Given a command", t, func() {
		cmd := "ls -{olia}"
		Convey("When the command is executed", func() {
			err := RunCommand(cmd, "/", "id")
			Convey("Then the result should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
func TestRun(t *testing.T) {
	Convey("Given a command", t, func() {
		cmd := "ls -la"
		Convey("When the command is executed", func() {
			err := RunCommand(cmd, "/", "id")
			Convey("Then the result should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestRun_ID_Changed(t *testing.T) {
	Convey("Given a command with {ID} tag", t, func() {
		cmd := "ls -{ID}"
		Convey("When the command is executed", func() {
			err := RunCommand(cmd, "/", "la")
			Convey("Then the result should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
