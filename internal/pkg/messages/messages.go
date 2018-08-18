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

//NewQueueMsgWithError creates the message with id and error
func NewQueueMsgWithError(id string, errMsg string) *QueueMessage {
	return &QueueMessage{ID: id, Error: errMsg}
}
