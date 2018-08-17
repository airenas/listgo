package messages

// Sender sends a messages to message broker
type Sender interface {
	Send(message *QueueMessage, queue string, replyQueue string) error
}
