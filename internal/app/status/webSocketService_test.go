package status

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testdataWS struct {
	ch     chan string
	readCh chan bool
	fc     chan bool
	f      func()
	conn   *wsConnMock
}

func initTestDataWS(t *testing.T) *testdataWS {
	initTest(t)
	res := testdataWS{}
	res.ch = make(chan string)
	res.fc = make(chan bool)
	res.readCh = make(chan bool)
	res.conn = &wsConnMock{valueCh: res.ch, sCh: res.readCh}
	res.f = func() {
		handleConnection(res.conn)
		res.fc <- true
	}
	return &res
}

func TestHandleConnection_ReadFails_Closed(t *testing.T) {
	td := initTestDataWS(t)
	go td.f()

	close(td.ch)
	<-td.fc
	assert.Equal(t, td.conn.closedCount, 1)
}

func TestHandleConnection_ReadOK_Closed(t *testing.T) {
	td := initTestDataWS(t)
	go td.f()

	td.ch <- "id1"
	close(td.ch)
	<-td.fc
	assert.Equal(t, td.conn.closedCount, 1)

	assert.Equal(t, len(idConnectionMap), 0)
	assert.Equal(t, len(connectionIDMap), 0)
}

func TestHandleConnection_SeveralReadOK_Closed(t *testing.T) {
	td := initTestDataWS(t)
	go td.f()

	td.ch <- "ids"
	td.ch <- "ids2"
	td.ch <- "ids1"
	close(td.ch)
	<-td.fc

	assert.Equal(t, td.conn.closedCount, 1)

	assert.Equal(t, len(idConnectionMap), 0)
	assert.Equal(t, len(connectionIDMap), 0)
}

func TestHandleConnection_SeveralReadOK_Waiting(t *testing.T) {
	td := initTestDataWS(t)
	go td.f()

	td.ch <- "id3"
	<-td.readCh
	<-td.readCh // wait for next read
	c, ok := getConnections("id3")

	assert.True(t, ok)
	assert.True(t, c[td.conn])

	assert.Equal(t, len(idConnectionMap), 1)
	assert.Equal(t, len(connectionIDMap), 1)

	close(td.ch)
	<-td.fc
}

func TestHandleConnection_SeveralConnections(t *testing.T) {
	td := initTestDataWS(t)
	go td.f()

	td.ch <- "id4"

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
	assert.True(t, ok)
	assert.True(t, c[td.conn])
	assert.True(t, c[conn1])

	assert.Equal(t, len(idConnectionMap), 1)
	fmt.Println(connectionIDMap)
	assert.Equal(t, len(connectionIDMap), 2)

	assert.Equal(t, td.conn.closedCount, 0)
	assert.Equal(t, conn1.closedCount, 0)

	close(ch1)
	close(td.ch)
	<-td.fc
	<-fc1

	_, ok = getConnections("id4")
	assert.False(t, ok)
	assert.Equal(t, len(idConnectionMap), 0)
	fmt.Println(connectionIDMap)
	assert.Equal(t, len(connectionIDMap), 0)

	assert.Equal(t, td.conn.closedCount, 1)
	assert.Equal(t, conn1.closedCount, 1)
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

func (f *wsConnMock) WriteJSON(v interface{}) error {
	return nil
}
