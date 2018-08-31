package status

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHandleConnection(t *testing.T) {
	Convey("Given a mock connection", t, func() {
		ch := make(chan string)
		readCh := make(chan bool)
		fc := make(chan bool)
		conn := &wsConnMock{valueCh: ch, sCh: readCh}
		go func() {
			handleConnection(conn)
			fc <- true
		}()
		Convey("When read fails", func() {
			close(ch)
			<-fc
			Convey("Then the connection is closed", func() {
				So(conn.closedCount, ShouldEqual, 1)
			})
		})

		Convey("When read succeeds", func() {
			ch <- "id1"
			close(ch)
			<-fc
			Convey("Then the connection is closed", func() {
				So(conn.closedCount, ShouldEqual, 1)
			})
			Convey("Maps are empty", func() {
				So(len(idConnectionMap), ShouldEqual, 0)
				So(len(connectionIDMap), ShouldEqual, 0)
			})
		})
		Convey("When read succeeds several times", func() {
			ch <- "ids"
			ch <- "ids2"
			ch <- "ids1"
			close(ch)
			<-fc
			Convey("Then the connection is closed", func() {
				So(conn.closedCount, ShouldEqual, 1)
			})
			Convey("Maps are empty", func() {
				So(len(idConnectionMap), ShouldEqual, 0)
				So(len(connectionIDMap), ShouldEqual, 0)
			})
		})
		Convey("When read succeeds 2", func() {
			ch <- "id3"
			<-readCh
			<-readCh // wait for next read
			c, ok := getConnections("id3")
			Convey("Then return conn by id3", func() {
				So(ok, ShouldBeTrue)
				So(c[conn], ShouldBeTrue)
			})
			Convey("Then maps are not empty", func() {
				So(len(idConnectionMap), ShouldEqual, 1)
				So(len(connectionIDMap), ShouldEqual, 1)
			})
			close(ch)
		})
		Convey("When Connection with same id arrives", func() {
			ch <- "id4"
			ch1 := make(chan string)
			readCh1 := make(chan bool)
			fc1 := make(chan bool)
			conn1 := &wsConnMock{valueCh: ch1, sCh: readCh1}
			go func() {
				handleConnection(conn1)
				fc1 <- true
			}()
			ch1 <- "id4"
			<-readCh1
			<-readCh1 // wait for next read
			c, ok := getConnections("id4")
			So(ok, ShouldBeTrue)

			Convey("Then return conn by id", func() {
				So(c[conn], ShouldBeTrue)
			})
			Convey("Then return conn1 by id", func() {
				So(c[conn1], ShouldBeTrue)
			})
			Convey("Then id map contains 1 value", func() {
				So(len(idConnectionMap), ShouldEqual, 1)
			})
			Convey("Then connection map contains two value", func() {
				fmt.Println(connectionIDMap)
				So(len(connectionIDMap), ShouldEqual, 2)
			})
			Convey("Then the connection is not closed", func() {
				So(conn.closedCount, ShouldEqual, 0)
			})
			Convey("Then the new connection is not closed", func() {
				So(conn.closedCount, ShouldEqual, 0)
			})
			close(ch1)
			close(ch)
			<-fc
			<-fc1
			Convey("Then connections are closed", func() {
				_, ok := getConnections("id4")

				Convey("No conn by id", func() {
					So(ok, ShouldBeFalse)
				})
				Convey("Then id map should be empty", func() {
					So(len(idConnectionMap), ShouldEqual, 0)
				})
				Convey("Then connection map should be empty", func() {
					So(len(connectionIDMap), ShouldEqual, 0)
				})
				Convey("Then the connection is closed", func() {
					So(conn.closedCount, ShouldEqual, 1)
				})
				Convey("Then the new connection is closed", func() {
					So(conn.closedCount, ShouldEqual, 1)
				})
			})
		})

	})
}

type wsConnMock struct {
	sCh         chan<- bool   // start
	valueCh     <-chan string // value
	closedCount int
}

func (f *wsConnMock) ReadMessage() (messageType int, p []byte, err error) {
	go func() { f.sCh <- true }()
	s, ok := <-f.valueCh
	if ok {
		return 1, []byte(s), nil
	}
	return 1, nil, errors.New("closed")
}

func (f *wsConnMock) Close() error {
	f.closedCount++
	return nil
}
