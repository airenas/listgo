package dispatcher

import (
	"testing"

	"github.com/airenas/listgo/internal/pkg/recognizer"
	"github.com/airenas/listgo/internal/pkg/test/mocks"
	"github.com/petergtz/pegomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var recInfoLoaderMock *mocks.MockRecInfoLoader

func initTestTypeGetter(t *testing.T) {
	mocks.AttachMockToTest(t)
	recInfoLoaderMock = mocks.NewMockRecInfoLoader()
	pegomock.When(recInfoLoaderMock.Get(pegomock.EqString("rkey"))).ThenReturn(newTestRec(), nil)
}

func TestInit(t *testing.T) {
	initTestTypeGetter(t)
	g, err := newTypeGetter(recInfoLoaderMock, "key")
	assert.Nil(t, err)
	assert.NotNil(t, g)
}

func TestInit_NoKey(t *testing.T) {
	initTestTypeGetter(t)
	_, err := newTypeGetter(recInfoLoaderMock, "")
	assert.NotNil(t, err)
}

func TestInit_NoLoader(t *testing.T) {
	initTestTypeGetter(t)
	_, err := newTypeGetter(nil, "key")
	assert.NotNil(t, err)
}

func TestGetSetting_ErrorByReconizer(t *testing.T) {
	initTestTypeGetter(t)
	pegomock.When(recInfoLoaderMock.Get(pegomock.EqString("rkey"))).ThenReturn(nil, errors.New("No rec"))
	g, _ := newTypeGetter(recInfoLoaderMock, "key1")
	assert.NotNil(t, g)
	_, err := g.Get("rkey")
	assert.NotNil(t, err)
}

func TestGetSetting_NoKey(t *testing.T) {
	initTestTypeGetter(t)
	g, _ := newTypeGetter(recInfoLoaderMock, "key1")
	assert.NotNil(t, g)
	_, err := g.Get("rkey")
	assert.NotNil(t, err)
}

func TestGetSetting_Success(t *testing.T) {
	initTestTypeGetter(t)
	g, _ := newTypeGetter(recInfoLoaderMock, "key")
	assert.NotNil(t, g)
	r, err := g.Get("rkey")
	assert.Nil(t, err)
	assert.Equal(t, "vOlia", r)
}

func newTestRec() *recognizer.Info {
	return &recognizer.Info{Settings: map[string]string{"key": "vOlia"}}
}
