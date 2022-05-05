package clean

import (
	"testing"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
)

var idsProviderMock *mocks.MockOldIDsProvider

func initTimerTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	cleanerMock = mocks.NewMockCleaner()
	idsProviderMock = mocks.NewMockOldIDsProvider()
	pegomock.When(cleanerMock.Clean(pegomock.AnyString())).ThenReturn(nil)
	pegomock.When(idsProviderMock.Get()).ThenReturn([]string{}, nil)
}

func TestInvokesOnStartup(t *testing.T) {
	initTimerTest(t)
	d := newtData()

	startCleanTimer(d)

	go close(d.qChan)
	<-d.workWaitChan
	idsProviderMock.VerifyWasCalled(pegomock.Once()).Get()
}

func TestInvokesOnTimer(t *testing.T) {
	initTimerTest(t)
	d := newtData()
	d.runEvery = time.Millisecond * 5

	startCleanTimer(d)

	time.Sleep(30 * time.Millisecond)
	go close(d.qChan)
	<-d.workWaitChan
	idsProviderMock.VerifyWasCalled(pegomock.AtLeast(3)).Get()
}

func TestInvokesCleaner(t *testing.T) {
	initTimerTest(t)
	d := newtData()
	pegomock.When(idsProviderMock.Get()).ThenReturn([]string{"1", "2"}, nil)

	startCleanTimer(d)

	go close(d.qChan)
	<-d.workWaitChan
	cleanerMock.VerifyWasCalled(pegomock.Twice()).Clean(pegomock.AnyString())
}

func TestInvokesCleanerWithFailure(t *testing.T) {
	initTimerTest(t)
	d := newtData()
	pegomock.When(idsProviderMock.Get()).ThenReturn([]string{"1", "2"}, nil)
	pegomock.When(cleanerMock.Clean(pegomock.AnyString())).ThenReturn(errors.New("error"))

	startCleanTimer(d)

	go close(d.qChan)
	<-d.workWaitChan
	cleanerMock.VerifyWasCalled(pegomock.Twice()).Clean(pegomock.AnyString())
}

func TestContinuesOnProviderError(t *testing.T) {
	initTimerTest(t)
	d := newtData()
	pegomock.When(idsProviderMock.Get()).ThenReturn(nil, errors.New("error"))
	d.runEvery = time.Millisecond * 10

	startCleanTimer(d)

	time.Sleep(55 * time.Millisecond)
	go close(d.qChan)
	<-d.workWaitChan
	idsProviderMock.VerifyWasCalled(pegomock.AtLeast(5)).Get()
}

func newtData() *timerServiceData {
	data := timerServiceData{}
	data.workWaitChan = make(chan struct{})
	data.qChan = make(chan struct{})
	data.runEvery = time.Minute
	data.cleaner = cleanerMock
	data.idsProvider = idsProviderMock
	return &data
}
