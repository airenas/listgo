package dispatcher

import (
	"testing"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var loaderMock *mocks.MockLoader

func initTestDuration(t *testing.T) {
	mocks.AttachMockToTest(t)
	loaderMock = mocks.NewMockLoader()
}

func TestInitDurationLoader(t *testing.T) {
	l, err := newDurationLoader("/aaa/{ID}/aa")
	assert.Nil(t, err)
	assert.NotNil(t, l)
}

func TestInitDuration_NoPath(t *testing.T) {
	_, err := newDurationLoader("")
	assert.NotNil(t, err)
}

func TestInitDuration_NoID(t *testing.T) {
	_, err := newDurationLoader("/olia/")
	assert.NotNil(t, err)
}

func TestGetDuration(t *testing.T) {
	initTestDuration(t)
	pegomock.When(loaderMock.Read(pegomock.AnyString())).ThenReturn([]byte("a a 100 200 a"), nil)
	l, _ := newDurationLoaderInt("/aaa/{ID}/aa", loaderMock)
	assert.NotNil(t, l)
	r, err := l.Get("key")
	assert.Nil(t, err)
	assert.Equal(t, 3*time.Second, r)
	calledFile := loaderMock.VerifyWasCalled(pegomock.Once()).Read(pegomock.AnyString()).GetCapturedArguments()
	assert.Equal(t, "/aaa/key/aa", calledFile)
}

func TestGetDuration_OnErrorDefault(t *testing.T) {
	initTestDuration(t)
	pegomock.When(loaderMock.Read(pegomock.AnyString())).ThenReturn(nil, errors.New("err"))
	l, _ := newDurationLoaderInt("/aaa/{ID}/aa", loaderMock)
	assert.NotNil(t, l)
	r, err := l.Get("key")
	assert.NotNil(t, err)
	assert.Equal(t, defDuration, r)
}

func TestGetDuration_SelectMax(t *testing.T) {
	initTestDuration(t)
	pegomock.When(loaderMock.Read(pegomock.AnyString())).
		ThenReturn([]byte("a a 100 200 a\na a 1000 200 a\a a 1000 100 a"), nil)
	l, _ := newDurationLoaderInt("/aaa/{ID}/aa", loaderMock)
	assert.NotNil(t, l)
	r, err := l.Get("key")
	assert.Nil(t, err)
	assert.Equal(t, 12*time.Second, r)
}
