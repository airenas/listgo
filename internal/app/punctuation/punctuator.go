package punctuation

import (
	"bufio"
	"io"
	"strings"
	"unicode"

	"github.com/airenas/listgo/internal/app/punctuation/api"
	"github.com/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
)

// DataProvider provides data to initializer
type DataProvider interface {
	GetVocab() (io.Reader, error)
	GetData() (*api.Data, error)
}

// TFWrap makes real call to tensorflow service
type TFWrap interface {
	Invoke([]int32) ([]int32, error)
}

// PunctuatorImpl implements punctuator service
type PunctuatorImpl struct {
	vocab        map[string]int32
	puncVocab    map[int32]string
	sentenceEnds map[int32]bool
	timesteps    int
	tfWrap       TFWrap
	unkID        int32
	seID         int32
	numID        int32
}

// NewPunctuatorImpl creates instance
func NewPunctuatorImpl(d DataProvider, tfWrap TFWrap) (*PunctuatorImpl, error) {
	p := PunctuatorImpl{}
	var err error
	p.vocab, err = readVocab(d)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot init vocabulary")
	}
	cmdapp.Log.Infof("Vocab size: %d", len(p.vocab))

	data, err := d.GetData()
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get data")
	}
	p.timesteps = data.Timesteps
	if p.timesteps < 3 {
		return nil, errors.Errorf("Wrong timesteps, Timesteps = %d", p.timesteps)
	}
	p.puncVocab = initPunctuations(data.PunctuationVocabulary)
	cmdapp.Log.Infof("Punctuation vocab size: %d", len(p.puncVocab))
	p.sentenceEnds = initSentenceEnds(data.SentenceEnd, p.puncVocab)
	p.tfWrap = tfWrap
	if p.tfWrap == nil {
		return nil, errors.New("No TF wrapper set")
	}
	var f bool
	p.unkID, f = p.vocab[data.UnknownWord]
	if !f {
		return nil, errors.Errorf("Cannot find <UNK> in vocabulary, UNK = %s", data.UnknownWord)
	}
	p.seID, f = p.vocab[data.SequenceEndWord]
	if !f {
		return nil, errors.Errorf("Cannot find sequence end word in vocabulary, SE = %s", data.SequenceEndWord)
	}
	p.numID, f = p.vocab[data.NumdWord]
	if !f {
		return nil, errors.Errorf("Cannot find num word in vocabulary, NUM = %s", data.NumdWord)
	}
	return &p, nil
}

// Process is main Punctuator method
func (p *PunctuatorImpl) Process(text []string) (*api.PResult, error) {
	result := &api.PResult{}
	result.Original = text
	result.WordIDs = p.convertToNum(text)
	var err error
	result.PunctIDs, err = p.punctuate(result.WordIDs)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot punctuate")
	}
	result.Punctuated, err = p.preparePunctuated(result.Original, result.PunctIDs)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot prepare result")
	}
	result.PunctuatedText = p.fillText(result.Punctuated)
	return result, nil
}

func readVocab(d DataProvider) (map[string]int32, error) {
	vdata, err := d.GetVocab()
	if err != nil {
		return nil, err
	}
	result := make(map[string]int32)
	var i int32
	i = 0
	scanner := bufio.NewScanner(vdata)
	for scanner.Scan() {
		s := scanner.Text()
		result[s] = i
		i++
	}
	return result, nil
}

func initPunctuations(str []string) map[int32]string {
	result := make(map[int32]string)
	var i int32
	i = 0
	for _, s := range str {
		result[i] = s
		i++
	}
	return result
}

func initSentenceEnds(str []string, pvr map[int32]string) map[int32]bool {
	pv := make(map[string]int32)
	for k, v := range pvr {
		pv[v] = k
	}
	result := make(map[int32]bool)
	for _, s := range str {
		i, f := pv[s]
		if f {
			result[i] = true
		} else {
			cmdapp.Log.Warnf("Unknown sentence end string: %s", s)
		}
	}
	return result
}

func (p *PunctuatorImpl) convertToNum(strs []string) []int32 {
	result := make([]int32, 0)
	for _, s := range strs {
		k, f := p.vocab[s]
		if !f {
			if isNum(s) {
				k = p.numID
			} else {
				k = p.unkID
			}
		}
		result = append(result, k)
	}
	return result
}

func (p *PunctuatorImpl) punctuate(nums []int32) ([]int32, error) {
	l := len(nums)
	result := make([]int32, l)
	numsP := make([]int32, p.timesteps)
	for ci := 0; ci < l; {
		p.copyArr(numsP, nums, ci)
		numsP[p.timesteps-1] = p.seID
		res, err := p.tfWrap.Invoke(numsP)
		if err != nil {
			return nil, errors.Wrap(err, "Cannot invoke tensorflow service")
		}
		if len(res) < (p.timesteps - 1) {
			return nil, errors.Errorf("Wrong result returned. Len = %d", len(res))
		}
		ci = p.fillResult(result, res[0:p.timesteps-1], ci, l)
	}
	return result, nil
}

func (p *PunctuatorImpl) copyArr(nums []int32, from []int32, pos int) {
	l := len(from)
	i := 0
	to := p.timesteps - 1
	for ; pos+i < l && i < to; i++ {
		nums[i] = from[pos+i]
	}
	for i1 := 0; i < to; i++ { // add missing from the start
		nums[i] = from[i1%l]
		i1++
	}
}

func (p *PunctuatorImpl) fillResult(result []int32, res []int32, pos int, to int) int {
	lEnd := pos
	cpos := pos
	l := len(res)
	for i := 0; i < l && cpos < to; i++ {
		result[cpos] = res[i]
		_, f := p.sentenceEnds[res[i]]
		if f {
			lEnd = cpos + 1
		}
		cpos++
	}
	if lEnd == pos || cpos == to {
		lEnd = pos + p.timesteps - 1
	}
	return lEnd
}

func (p *PunctuatorImpl) preparePunctuated(strs []string, res []int32) ([]string, error) {
	to := len(strs)
	if to != len(res) {
		return nil, errors.Errorf("Result array is of wrong size. Expected %d, was %d", to, len(res))
	}
	result := make([]string, to)
	uc := true
	for i, v := range res {
		if i < to {
			w := strs[i]
			if uc {
				w = toTitle(w)
			}
			ps, _ := p.puncVocab[v]
			if ps == "-" {
				w = w + " "
			}
			w = w + ps
			result[i] = strings.TrimSpace(w)
			_, uc = p.sentenceEnds[v]
		}
	}
	return result, nil
}

func (p *PunctuatorImpl) fillText(strs []string) string {
	res := ""
	sep := ""
	for _, w := range strs {
		res = res + sep + w
		sep = " "
	}
	return res
}

func toTitle(data string) string {
	if len(data) == 0 {
		return data
	}
	r := []rune(data)
	r[0] = unicode.ToTitle(r[0])
	return string(r)
}

func isNum(word string) bool {
	res := false
	for _, c := range []rune(word) {
		if unicode.IsDigit(c) {
			res = true
			continue
		}
		if c == '.' || c == ',' || c == '/' || c == ':' || c == '-' {
			continue
		}
		return false
	}
	return res
}
