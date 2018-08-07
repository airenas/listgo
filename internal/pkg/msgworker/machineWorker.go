package msgworker

import (
	"bitbucket.org/airenas/listgo/internal/pkg/cmdapp"
	"github.com/RichardKnop/machinery/v1"
)

//MachineWorker listens for messages and performs work
type MachineWorker struct {
	Server *machinery.Server
}

//Listen listen for event queue
func (w *MachineWorker) Listen(workerName string) error {
	worker := w.Server.NewWorker(workerName, 0)
	cmdapp.Log.Info("Starting consume queue")
	return worker.Launch()
}

//RegisterTask register function to process message from queue
func (w *MachineWorker) RegisterTask(task string, taskFunction interface{}) error {
	return w.Server.RegisterTask(task, taskFunction)
}
