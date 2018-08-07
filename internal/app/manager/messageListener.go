package manager

// MessageListener listens for messages
type MessageListener interface {
	Listen(workerName string) error
	RegisterTask(name string, taskFunc interface{}) error
}
