package dispatcher

import (
	"github.com/airenas/listgo/internal/pkg/strategy/api"
	"github.com/pkg/errors"
)

type strategyWrapper struct {
	realStrategy api.TaskSelector
}

func newStrategyWrapper(realStrategy api.TaskSelector) (*strategyWrapper, error) {
	res := &strategyWrapper{}
	if realStrategy == nil {
		return nil, errors.New("No task selector provided")
	}
	res.realStrategy = realStrategy
	return res, nil
}

func (sw *strategyWrapper) FindBest(wrks []*worker, tsks map[string]*task, wi int) (*task, error) {
	rws := mapWorkers(wrks)
	rts := mapTasks(tsks)

	rt, err := sw.realStrategy.FindBest(rws, rts, wi)
	if err != nil {
		return nil, errors.Wrap(err, "Can't select best task")
	}
	if rt != nil {
		if rt.RealObject == nil {
			return nil, errors.New("No wrapped task object")
		}
		res := rt.RealObject.(*task)
		if res == nil {
			return nil, errors.New("No wrapped task object")
		}
		return res, nil
	}
	return nil, nil
}

func mapWorkers(wrks []*worker) []*api.Worker {
	res := make([]*api.Worker, len(wrks))
	for i, w := range wrks {
		nw := &api.Worker{}
		nw.EndAt = w.endAt
		nw.TaskType = w.mType
		res[i] = nw
	}
	return res
}

func mapTasks(tsks map[string]*task) []*api.Task {
	res := make([]*api.Task, 0)
	for _, v := range tsks {
		if !v.started && v.failCount < maxTaskFailCount {
			nt := &api.Task{}
			nt.TaskType = v.requiredModelType
			nt.Duration = v.expDuration
			nt.ArrivedAt = v.addedAt
			nt.RealObject = v
			res = append(res, nt)
		}
	}
	return res
}
