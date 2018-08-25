package progress

import (
	"testing"

	"bitbucket.org/airenas/listgo/internal/pkg/status"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConvert(t *testing.T) {
	Convey("Given a status", t, func() {
		Convey("progress is calculated", func() {
			pr := Convert(status.AudioConvert.Name)
			So(pr, ShouldBeGreaterThan, 0)
		})
		Convey("unknown == 0", func() {
			pr := Convert("olia")
			So(pr, ShouldEqual, 0)
		})
		Convey("COMPLETED == 100", func() {
			pr := Convert(status.Completed.Name)
			So(pr, ShouldEqual, 100)
		})
	})
}
