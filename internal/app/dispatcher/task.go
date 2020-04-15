package dispatcher

type task struct {
}

func newTask() *task {
	res := &task{}
	return res
}

func failRequeueTask(t *task) {

}
