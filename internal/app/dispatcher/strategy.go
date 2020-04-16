package dispatcher

import (
	"bitbucket.org/airenas/listgo/internal/pkg/strategy/api"
	"github.com/pkg/errors"
)

type strategyWrapper struct {
	realStrategy api.TaskSelector
}

func newStrategyWrapper(realStrategy api.TaskSelector) (*strategyWrapper, error) {
	res := &strategyWrapper{realStrategy: realStrategy}
	return res, nil
}

func (sw *strategyWrapper) findBest(wrks []*worker, tsks *tasks, wi int) (*task, error) {
	rws := mapWorkers(wrks)
	rts := mapTasks(tsks)

	rt, err := sw.realStrategy.FindBest(rws, rts, wi)
	if err != nil {
		return nil, errors.Wrap(err, "Can't select best task")
	}
	if rt != nil {
		return rt.RealObject.(*task), nil
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

func mapTasks(tsks *tasks) []*api.Task {
	res := make([]*api.Task, 0)
	for _, v := range tsks.tsks {
		if !v.started {
			nt := &api.Task{}
			nt.TaskType = v.requiredModelType
			nt.Duration = v.expectedDuration
			nt.ArrivedAt = v.addedAt
			res = append(res, nt)
		}
	}
	return res
}
