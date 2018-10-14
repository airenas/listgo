package messages

//Message base message interface for sending to queue
type Message interface {
}

// Sender sends a messages to message broker
type Sender interface {
	Send(message Message, queue string, replyQueue string) error
}
