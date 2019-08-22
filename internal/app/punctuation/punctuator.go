package punctuation

import (
	"bufio"
	"io"
	"strings"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
)

//Data keeps punctuation settings
type Data struct {
	Info                  string
	PunctuationVovabulary []string `yaml:"puctuationVocabulary,flow"`
	SentenceEnd           []string `yaml:"sentenceEnd,flow"`
	Timesteps             int      `yaml:"timesteps"`
}

//DataProvider provides data to initializer
type DataProvider interface {
	GetVocab() (io.ReadCloser, error)
	GetData() (*Data, error)
}

//TFWrap makes real call to tensorflow service
type TFWrap interface {
	Invoke([]int32) ([]int32, error)
}

//PunctuatorImpl implements punctuator service
type PunctuatorImpl struct {
	vocab        map[string]int32
	puncVocab    map[int32]string
	sentenceEnds map[int32]bool
	timesteps    int
	tfWrap       TFWrap
}

//NewPunctuatorImpl creates instance
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
	p.puncVocab = initPunctuations(data.PunctuationVovabulary)
	cmdapp.Log.Infof("Punctuation vocab size: %d", len(p.puncVocab))
	p.sentenceEnds = initSentenceEnds(data.SentenceEnd, p.puncVocab)
	p.tfWrap = tfWrap
	return &p, nil
}

//Process is main Punctuator method
func (p *PunctuatorImpl) Process(text string) (string, error) {
	arr := p.convertToArray(text)
	num := p.convertToNum(arr)
	punct, err := p.punctuate(num)
	if err != nil {
		return "", errors.Wrap(err, "Cannot punctuate")
	}
	result := p.prepareText(arr, punct)
	return result, nil
}

func readVocab(d DataProvider) (map[string]int32, error) {
	file, err := d.GetVocab()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make(map[string]int32)
	var i int32
	i = 0
	scanner := bufio.NewScanner(file)
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

func (p *PunctuatorImpl) convertToArray(strs string) []string {
	arr := strings.Split(strs, " ")
	result := make([]string, 0)
	for _, s := range arr {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

func (p *PunctuatorImpl) convertToNum(strs []string) []int32 {
	result := make([]int32, 0)
	for _, s := range strs {
		k, f := p.vocab[s]
		if !f {
			k = p.vocab["<UNK>"]
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
		numsP[p.timesteps-1] = p.vocab["</S>"]
		res, err := p.tfWrap.Invoke(numsP)
		if err != nil {
			return nil, errors.Wrap(err, "Cannot invoke tensorflow service")
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
			lEnd = cpos
		}
		cpos++
	}
	if lEnd == pos || cpos == to {
		lEnd = pos + p.timesteps
	}
	return lEnd
}

func (p *PunctuatorImpl) prepareText(strs []string, res []int32) string {
	to := len(strs)
	result := ""
	for i, v := range res {
		if i < to {
			result = result + strs[i]
			p, _ := p.puncVocab[v]
			if p != "" {
				result = result + p
			}
			result = result + " "
		}
	}
	return result
}
