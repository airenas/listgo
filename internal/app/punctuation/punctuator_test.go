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
var tfWrapMock *mocks.MockTFWrap

func initPTest(t *testing.T) {
	mocks.AttachMockToTest(t)
	dpMock = mocks.NewMockDataProvider()
	tfWrapMock = mocks.NewMockTFWrap()
	pegomock.When(dpMock.GetData()).ThenReturn(defaultData(), nil)
	pegomock.When(dpMock.GetVocab()).ThenReturn(defaultTestVocab(), nil)
	pegomock.When(tfWrapMock.Invoke(pegomock.AnyInt32Slice())).ThenReturn(defaultIntResult(), nil)
}

func initPunctTest(t *testing.T) *PunctuatorImpl {
	initPTest(t)
	p, err := NewPunctuatorImpl(dpMock, tfWrapMock)
	assert.NotNil(t, p)
	assert.Nil(t, err)
	return p
}

func TestInitOK(t *testing.T) {
	initPTest(t)
	p, err := NewPunctuatorImpl(dpMock, tfWrapMock)
	assert.Nil(t, err)
	if err != nil {
		assert.NotNil(t, p, err.Error())
	} else {
		assert.NotNil(t, p)
	}
}
func TestInit_NoUNK_Fails(t *testing.T) {
	initPTest(t)
	d := defaultData()
	d.UnknownWord = "<UUU>"
	pegomock.When(dpMock.GetData()).ThenReturn(d, nil)
	p, err := NewPunctuatorImpl(dpMock, tfWrapMock)
	assert.Nil(t, p)
	assert.NotNil(t, err)
}

func TestInit_NoSE_Fails(t *testing.T) {
	initPTest(t)
	d := defaultData()
	d.SequenceEndWord = "<UUU>"
	pegomock.When(dpMock.GetData()).ThenReturn(d, nil)
	p, err := NewPunctuatorImpl(dpMock, tfWrapMock)
	assert.Nil(t, p)
	assert.NotNil(t, err)
}

func TestInit_NoNUM_Fails(t *testing.T) {
	initPTest(t)
	d := defaultData()
	d.NumdWord = "NNN"
	pegomock.When(dpMock.GetData()).ThenReturn(d, nil)
	p, err := NewPunctuatorImpl(dpMock, tfWrapMock)
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

func TestIsNum(t *testing.T) {
	assert.True(t, isNum("10"))
	assert.True(t, isNum("10.10"))
	assert.True(t, isNum("10.10,20"))
	assert.True(t, isNum("/10.10,20:-10"))
}

func TestIsNotNum(t *testing.T) {
	assert.False(t, isNum("a10"))
	assert.False(t, isNum("XI10.10"))
	assert.False(t, isNum("olia"))
	assert.False(t, isNum("<NUM>"))
	assert.False(t, isNum("ltšą"))
}

func TestProcess_OK(t *testing.T) {
	p := initPunctTest(t)
	pegomock.When(tfWrapMock.Invoke(pegomock.AnyInt32Slice())).ThenReturn([]int32{0, 0, 0, 0, 0}, nil)
	r, err := p.Process(strings.Split("a a", " "))
	assert.Nil(t, err)
	assert.Equal(t, []string{"A", "a"}, r.Punctuated)
}

func TestProcess_ReturnsOriginal(t *testing.T) {
	p := initPunctTest(t)
	pegomock.When(tfWrapMock.Invoke(pegomock.AnyInt32Slice())).ThenReturn([]int32{0, 0, 0, 0, 0}, nil)
	r, err := p.Process(strings.Split("a a", " "))
	assert.Nil(t, err)
	assert.Equal(t, []string{"a", "a"}, r.Original)
}

func TestProcess_ReturnsPuntuatedText(t *testing.T) {
	p := initPunctTest(t)
	pegomock.When(tfWrapMock.Invoke(pegomock.AnyInt32Slice())).ThenReturn([]int32{0, 0, 0, 0, 0}, nil)
	r, err := p.Process(strings.Split("a a", " "))
	assert.Nil(t, err)
	assert.Equal(t, "A a", r.PunctuatedText)
}

func TestProcess_FirstWord_Uppercase(t *testing.T) {
	p := initPunctTest(t)
	pegomock.When(tfWrapMock.Invoke(pegomock.AnyInt32Slice())).ThenReturn([]int32{0, 0, 0, 0, 0}, nil)
	r, err := p.Process(strings.Split("aaaa a", " "))
	assert.Nil(t, err)
	assert.Equal(t, []string{"Aaaa", "a"}, r.Punctuated)
	r, _ = p.Process([]string{"aaaa a", "a"})
	assert.Equal(t, []string{"Aaaa a", "a"}, r.Punctuated)
}

func TestProcess_AddPunctuation(t *testing.T) {
	p := initPunctTest(t)
	pegomock.When(tfWrapMock.Invoke(pegomock.AnyInt32Slice())).ThenReturn([]int32{0, 1, 2, 0, 0}, nil)
	r, err := p.Process(strings.Split("aaaa a b b", " "))
	assert.Nil(t, err)
	assert.Equal(t, []string{"Aaaa", "a,", "b.", "B"}, r.Punctuated)
}

func TestProcess_AddDash(t *testing.T) {
	p := initPunctTest(t)
	pegomock.When(tfWrapMock.Invoke(pegomock.AnyInt32Slice())).ThenReturn([]int32{0, 1, 3, 0, 0}, nil)
	r, err := p.Process(strings.Split("aaaa a b b", " "))
	assert.Nil(t, err)
	assert.Equal(t, []string{"Aaaa", "a,", "b -", "b"}, r.Punctuated)
}

func TestProcess_Split(t *testing.T) {
	p := initPunctTest(t)
	pegomock.When(tfWrapMock.Invoke(pegomock.AnyInt32Slice())).ThenReturn([]int32{0, 0, 2, 0}, nil)
	r, err := p.Process(strings.Split("a b a b a b a", " "))
	assert.Nil(t, err)
	assert.Equal(t, []string{"A", "b", "a.", "B", "a", "b.", "A"}, r.Punctuated)
	tfWrapMock.VerifyWasCalled(pegomock.Times(2)).Invoke(pegomock.AnyInt32Slice())
}

func TestProcess_Split3(t *testing.T) {
	p := initPunctTest(t)
	pegomock.When(tfWrapMock.Invoke(pegomock.AnyInt32Slice())).ThenReturn([]int32{0, 0, 2, 0}, nil)
	r, err := p.Process(strings.Split("a b a b a b a b", " "))
	assert.Nil(t, err)
	assert.Equal(t, []string{"A", "b", "a.", "B", "a", "b.", "A", "b"}, r.Punctuated)
	tfWrapMock.VerifyWasCalled(pegomock.Times(3)).Invoke(pegomock.AnyInt32Slice())
}

func TestProcess_Split2(t *testing.T) {
	p := initPunctTest(t)
	pegomock.When(tfWrapMock.Invoke(pegomock.AnyInt32Slice())).ThenReturn([]int32{0, 0, 1, 0}, nil)
	r, err := p.Process(strings.Split("a b a b a b a b", " "))
	assert.Nil(t, err)
	assert.Equal(t, []string{"A", "b", "a,", "b", "a", "b", "a,", "b"}, r.Punctuated)
	tfWrapMock.VerifyWasCalled(pegomock.Times(2)).Invoke(pegomock.AnyInt32Slice())
}

func TestProcess_SplitLast(t *testing.T) {
	p := initPunctTest(t)
	pegomock.When(tfWrapMock.Invoke(pegomock.AnyInt32Slice())).ThenReturn([]int32{0, 0, 0, 2}, nil)
	r, err := p.Process(strings.Split("a b a b a b a b", " "))
	assert.Nil(t, err)
	assert.Equal(t, []string{"A", "b", "a", "b.", "A", "b", "a", "b."}, r.Punctuated)
	tfWrapMock.VerifyWasCalled(pegomock.Times(2)).Invoke(pegomock.AnyInt32Slice())
}

func TestProcess_ReturnWordIDs(t *testing.T) {
	p := initPunctTest(t)
	pegomock.When(tfWrapMock.Invoke(pegomock.AnyInt32Slice())).ThenReturn([]int32{0, 0, 0, 2}, nil)
	r, err := p.Process(strings.Split("a b a b xxx", " "))
	assert.Nil(t, err)
	assert.Equal(t, []int32{0, 1, 0, 1, 2}, r.WordIDs)
}

func TestProcess_ReturnPunctIDs(t *testing.T) {
	p := initPunctTest(t)
	pegomock.When(tfWrapMock.Invoke(pegomock.AnyInt32Slice())).ThenReturn([]int32{0, 1, 2, 0}, nil)
	r, err := p.Process(strings.Split("a b a", " "))
	assert.Nil(t, err)
	assert.Equal(t, []int32{0, 1, 2}, r.PunctIDs)
}

func TestProcess_Nums(t *testing.T) {
	p := initPunctTest(t)
	pegomock.When(tfWrapMock.Invoke(pegomock.AnyInt32Slice())).ThenReturn([]int32{0, 0, 0, 0}, nil)
	r, err := p.Process(strings.Split("a 10.125 10 xxx", " "))
	assert.Nil(t, err)
	assert.Equal(t, []int32{0, 4, 4, 2}, r.WordIDs)
	ca := tfWrapMock.VerifyWasCalled(pegomock.Times(1)).Invoke(pegomock.AnyInt32Slice()).GetCapturedArguments()
	assert.Equal(t, []int32{0, 4, 4, 2, 3}, ca)
}

func newTestVocab(v string) io.Reader {
	return strings.NewReader(v)
}

func defaultData() *api.Data {
	r := api.Data{}
	r.UnknownWord = "<UNK>"
	r.SequenceEndWord = "</S>"
	r.NumdWord = "<NUM>"
	r.PunctuationVocabulary = []string{" ", ",", ".", "-"}
	r.SentenceEnd = []string{"."}
	r.Timesteps = 5
	return &r
}

func defaultTestVocab() io.Reader {
	return newTestVocab(`a
b
<UNK>
</S>
<NUM>`)
}

func defaultIntResult() []int32 {
	return []int32{0, 0, 0, 0, 0}
}
