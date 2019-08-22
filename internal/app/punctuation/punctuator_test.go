package punctuation

import (
	"io"
	"strings"
	"testing"

	"bitbucket.org/airenas/listgo/internal/app/punctuation/api"
	"bitbucket.org/airenas/listgo/internal/pkg/test/mocks"
	"github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
)

var dpMock *mocks.MockDataProvider

func initPTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	dpMock = mocks.NewMockDataProvider()
	pegomock.When(dpMock.GetData()).ThenReturn(defaultData(), nil)
	pegomock.When(dpMock.GetVocab()).ThenReturn(defaultVocab(), nil)
}

func initPunctTest(t *testing.T) *PunctuatorImpl {
	initPTest(t)
	p, err := NewPunctuatorImpl(dpMock, nil)
	assert.NotNil(t, p)
	assert.Nil(t, err)
	return p
}

func TestInitOK(t *testing.T) {
	initPTest(t)
	p, err := NewPunctuatorImpl(dpMock, nil)
	assert.Nil(t, err)
	if err != nil {
		assert.NotNil(t, p, err.Error())
	} else {
		assert.NotNil(t, p)
	}
}
func TestInit_NuUNK_Fails(t *testing.T) {
	initPTest(t)
	d := defaultData()
	d.UnknownWord = "<UUU>"
	pegomock.When(dpMock.GetData()).ThenReturn(d, nil)
	p, err := NewPunctuatorImpl(dpMock, nil)
	assert.Nil(t, p)
	assert.NotNil(t, err)
}

func TestInit_NoSE_Fails(t *testing.T) {
	initPTest(t)
	d := defaultData()
	d.SequenceEndWord = "<UUU>"
	pegomock.When(dpMock.GetData()).ThenReturn(d, nil)
	p, err := NewPunctuatorImpl(dpMock, nil)
	assert.Nil(t, p)
	assert.NotNil(t, err)
}

func TestReadVocab(t *testing.T) {
	p := initPunctTest(t)
	_, f := p.vocab["a"]
	assert.True(t, f)
}

func TestReadPunctVocab(t *testing.T) {
	p := initPunctTest(t)
	_, f := p.puncVocab[1]
	assert.True(t, f)
}

func TestReadSentenceEnd(t *testing.T) {
	p := initPunctTest(t)
	_, f := p.sentenceEnds[2]
	assert.True(t, f)
}

func newTestVocab(v string) io.Reader {
	return strings.NewReader(v)
}

func defaultData() *api.Data {
	r := api.Data{}
	r.UnknownWord = "<UNK>"
	r.SequenceEndWord = "</S>"
	r.PunctuationVovabulary = []string{" ", ",", "."}
	r.SentenceEnd = []string{"."}
	return &r
}

func defaultVocab() io.Reader {
	return newTestVocab("a\n<UNK>\n</S>")
}
