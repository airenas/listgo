package messages

//QueueMessage message going throuht broker
type QueueMessage struct {
	ID     string `json:"id"`
	Error  string `json:"error"`
	Result string `json:"result"`
}

//NewQueueMessage creates the message with id
func NewQueueMessage(id string) *QueueMessage {
	return &QueueMessage{ID: id}
}
