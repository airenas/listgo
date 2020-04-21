package strategy

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"bitbucket.org/airenas/listgo/internal/pkg/strategy/api"
)

//Cost based strategy
type Cost struct {
	modelLoadTime   time.Duration
	rtFactor        float32
	delayCostPerSec float32
}

//NewCost init new Cost task selection strategy
func NewCost() (*Cost, error) {
	return newCost(cmdapp.Config.GetDuration("strategy.modelLoadDuration"),
		float32(cmdapp.Config.GetFloat64("strategy.realTimeFactor")),
		float32(cmdapp.Config.GetFloat64("strategy.delayCostPerSecond")))
}

func newCost(modelLoadTime time.Duration, rtFactor float32, delayCostPerSec float32) (*Cost, error) {
	res := &Cost{}
	if modelLoadTime <= 0 {
		return nil, errors.New("Wrong or no strategy.modelLoadDuration")
	}
	res.modelLoadTime = modelLoadTime
	if rtFactor <= 0.01 || rtFactor > 300 {
		return nil, errors.New("Wrong or no strategy.realTimeFactor")
	}
	res.rtFactor = rtFactor
	if delayCostPerSec <= 0 {
		return nil, errors.New("Wrong or no strategy.delayCostPerSecond")
	}
	res.delayCostPerSec = delayCostPerSec
	return res, nil
}

//FindBest is the main selection method
// select task for worker ws[workerIndex]
func (c *Cost) FindBest(ws []*api.Worker, ts []*api.Task, workerIndex int) (*api.Task, error) {
	ctx := newContext(time.Now())
	ctx.modelLoadTime = c.modelLoadTime
	ctx.delayCostPerSec = c.delayCostPerSec
	ctx.rtFactor = c.rtFactor

	tskg := groupTasks(ts)
	res := findTask(ws, tskg, workerIndex, ctx)
	return res, nil
}

type taskGroups struct {
	data map[string][]*api.Task
	keys []string
}

type context struct {
	now             time.Time
	modelLoadTime   time.Duration
	max             float32
	rtFactor        float32
	delayCostPerSec float32
}

func newContext(t time.Time) *context {
	return &context{now: t, max: 1000.0}
}

func findTask(ws []*api.Worker, tg *taskGroups, wi int, ctx *context) *api.Task {
	m := calcMatrix(ws, tg, ctx)
	bg := getBest(m, tg, wi, ctx)
	if bg != "" {
		res := tg.data[bg][0]
		return res
	}
	return getTaskByV2(ws, tg, wi, ctx)
}

func calcCost(w *api.Worker, t []*api.Task, ctx *context) float32 {
	res := float32(0.0)
	if len(t) == 0 {
		return ctx.max
	}
	if w.TaskType != t[0].TaskType {
		res += float32(ctx.modelLoadTime.Seconds())
	}
	d := ctx.now.Sub(t[0].ArrivedAt)
	if d > 0 {
		res -= float32(d.Seconds()) * ctx.delayCostPerSec
	}
	return res
}

func groupTasks(ts []*api.Task) *taskGroups {
	res := taskGroups{}
	res.data = make(map[string][]*api.Task, 0)
	for _, t := range ts {
		tl, _ := res.data[t.TaskType]
		res.data[t.TaskType] = append(tl, t)
	}
	for _, v := range res.data {
		sort.Slice(v, func(i, j int) bool { return v[i].ArrivedAt.Before(v[j].ArrivedAt) })
	}

	res.keys = make([]string, len(res.data))
	i := 0
	for k := range res.data {
		res.keys[i] = k
		i++
	}
	sort.Slice(res.keys, func(i, j int) bool { return res.keys[i] < res.keys[j] })

	return &res
}

func calcMatrix(ws []*api.Worker, tg *taskGroups, ctx *context) [][]float32 {
	res := make([][]float32, len(ws))
	for i, w := range ws {
		res[i] = make([]float32, len(tg.keys))
		for j, tk := range tg.keys {
			res[i][j] = calcCost(w, tg.data[tk], ctx)
			j++
		}
	}
	return res
}

func getTaskByV2(ws []*api.Worker, tg *taskGroups, wi int, ctx *context) *api.Task {
	var res *api.Task
	b := ctx.max
	for _, tk := range tg.keys {
		arr := make([]float32, len(ws))
		for i, w := range ws {
			arr[i] = float32(w.EndAt.Sub(ctx.now).Seconds())
			if arr[i] < 0 {
				arr[i] = 0
			}
			if tk != w.TaskType {
				arr[i] += float32(ctx.modelLoadTime.Seconds())
			}
		}
		for _, t := range tg.data[tk] {
			if isLowest(arr, wi) {
				if arr[wi] < b {
					b = arr[wi]
					res = t
					break
				}
			}
			addToLowest(&arr, t, ctx)
		}
	}
	return res
}

func print(mtrx [][]float32) {
	fmt.Fprintf(os.Stdout, "-----------------------\n")
	for _, r := range mtrx {
		for _, c := range r {
			fmt.Fprintf(os.Stdout, "%.1f\t", c)
		}
		fmt.Fprintf(os.Stdout, "\n")
	}
}

func printW(ws []*api.Worker, ctx *context) {
	fmt.Fprintf(os.Stdout, "----Workers--------------\n")
	for i, w := range ws {
		fmt.Fprintf(os.Stdout, "%d - tt: %s, end: %d\n", i, w.TaskType, toSec(w.EndAt, ctx))
	}
}

func printT(tg *taskGroups, ctx *context) {
	fmt.Fprintf(os.Stdout, "----Tasks--------------\n")
	for _, k := range tg.keys {
		fmt.Fprintf(os.Stdout, "%s\t", k)
	}
	fmt.Fprintf(os.Stdout, "\n")
	b := true
	for i := 0; b; i++ {
		b = false
		for _, k := range tg.keys {
			sl := tg.data[k]
			if len(sl) > i {
				fmt.Fprintf(os.Stdout, "%d-%.0f", toSec(sl[i].ArrivedAt, ctx), sl[i].Duration.Seconds())
				b = true
			}
			fmt.Fprintf(os.Stdout, "\t")
		}
		fmt.Fprintf(os.Stdout, "\n")
	}
}

func sortedIndexes(r []float32) []int {
	l := len(r)
	res := make([]int, len(r))
	for i := 1; i < l; i++ {
		res[i] = i
	}
	sort.Slice(res, func(i, j int) bool { return r[res[i]] < r[res[j]] })
	return res
}

func toSec(t time.Time, ctx *context) int {
	d := ctx.now.Sub(t)
	return int(d.Seconds())
}

func addToLowest(arr *[]float32, t *api.Task, ctx *context) {
	lv := ctx.max
	bi := 0
	for i, v := range *arr {
		if lv > v {
			bi = i
			lv = v
		}
	}
	d := ctx.now.Sub(t.ArrivedAt)
	(*arr)[bi] += float32(t.Duration.Seconds())
	if d > 0 {
		(*arr)[bi] -= float32(d.Seconds()) * ctx.delayCostPerSec
	}
}

func getBest(mtrx [][]float32, tg *taskGroups, wi int, ctx *context) string {
	sri := sortedIndexes(mtrx[wi])
	for _, ri := range sri {
		if mtrx[wi][ri] >= ctx.max {
			break
		}
		if isBestColumnValue(mtrx, wi, ri) {
			return tg.keys[ri]
		}
	}
	return ""
}

func isBestColumnValue(mtrx [][]float32, wi int, ri int) bool {
	bv := mtrx[wi][ri]
	lw := len(mtrx)
	for i := 0; i < lw; i++ {
		if mtrx[i][ri] < bv {
			return false
		}
	}
	return true
}

func isLowest(arr []float32, wi int) bool {
	for _, a := range arr {
		if a < arr[wi] {
			return false
		}
	}
	return true
}
