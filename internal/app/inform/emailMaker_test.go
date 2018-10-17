package inform

import (
	"testing"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/inform"

	"github.com/spf13/viper"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFailsInit(t *testing.T) {
	Convey("Given no url", t, func() {
		m, err := newSimpleEmailMaker(viper.New())
		Convey("Constructor should fail", func() {
			So(err, ShouldNotBeNil)
			So(m, ShouldBeNil)
		})
	})
}

func TestInit_OK(t *testing.T) {
	Convey("Given url", t, func() {
		v := viper.New()
		v.Set("mail.url", "url")
		m, err := newSimpleEmailMaker(v)
		Convey("Constructor should succeed", func() {
			So(err, ShouldBeNil)
			So(m.url, ShouldEqual, "url")
		})
	})
}

func TestEmail(t *testing.T) {
	Convey("Given congig", t, func() {
		v := viper.New()
		v.Set("mail.url", "url")
		v.Set("mail.x.subject", "subject")
		v.Set("mail.x.text", "text")
		m, _ := newSimpleEmailMaker(v)
		data := inform.Data{}
		data.Email = "email"
		data.ID = "id"
		data.MsgType = "x"
		data.MsgTime = time.Now()
		Convey("Mail should be made", func() {
			e, _ := m.Make(&data)
			So(e.Subject, ShouldEqual, "subject")
			So(e.To, ShouldContain, "email")
			So(string(e.Text), ShouldEqual, "text")
		})
		Convey("Should fail no subject", func() {
			v.Set("mail.x.subject", "")
			_, err := m.Make(&data)
			So(err, ShouldNotBeNil)
		})
		Convey("Should fail no text", func() {
			v.Set("mail.x.text", "")
			_, err := m.Make(&data)
			So(err, ShouldNotBeNil)
		})
		Convey("Should change ID", func() {
			v.Set("mail.x.text", "{{ID}}")
			e, _ := m.Make(&data)
			So(string(e.Text), ShouldEqual, "id")
		})
		Convey("Should change URL", func() {
			v.Set("mail.x.text", "{{URL}}")
			e, _ := m.Make(&data)
			So(string(e.Text), ShouldEqual, "url")
		})
		Convey("Should change Date", func() {
			v.Set("mail.x.text", "{{DATE}}")
			e, _ := m.Make(&data)
			So(string(e.Text), ShouldStartWith, data.MsgTime.Format("2006-01-02 15:04:05"))
		})
	})
}
