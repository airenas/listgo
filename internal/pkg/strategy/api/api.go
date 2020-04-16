package api

import (
	"time"
)

//Worker object wrapper
type Worker struct {
	TaskType string
	EndAt    time.Time
}

//Task object wrapper
type Task struct {
	TaskType  string
	ArrivedAt time.Time
	Duration  time.Duration

	RealObject interface{}
}

//TaskSelector provides the best task for worker ws[workerIndex]
type TaskSelector interface {
	FindBest(ws []*Worker, ts []*Task, workerIndex int) (*Task, error)
}
