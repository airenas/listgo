package dispatcher

import (
	"bufio"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/pkg/errors"
)

type durationLoader struct {
	pathPattern string
}

func newDurationLoader(pathPattern string) (*durationLoader, error) {
	if pathPattern == "" {
		return nil, errors.New("No duration path pattern set")
	}
	return &durationLoader{pathPattern: pathPattern}, nil
}

func (g *durationLoader) Get(id string) (time.Duration, error) {
	defDur := time.Second * 60
	file := strings.Replace(g.pathPattern, "{ID}", id, -1)
	cmdapp.Log.Infof("Loading file: %s", file)
	fData, err := ioutil.ReadFile(file)
	if err != nil {
		return defDur, errors.Wrap(err, "Can't load: "+file)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(fData)))
	res := time.Second * 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			//8689651a-0f62-4d3b-b12a-a87c917a9525 1 2412 226 M S U S0
			strs := strings.Split(line, " ")
			if len(strs) > 3 {
				d := toDuration(strs[2]) + toDuration(strs[3])
				if d > res {
					res = d
				}
			}
		}
	}
	return res, nil
}

func toDuration(s string) time.Duration {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0 * time.Second
	}
	return time.Duration(i) * 10 * time.Second
}
