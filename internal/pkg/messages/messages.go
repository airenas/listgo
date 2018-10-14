package messages

import (
	"time"
)

//QueueMessage message going throuht broker
type QueueMessage struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

//ResultMessage message going throuht broker with result
type ResultMessage struct {
	QueueMessage
	Result string `json:"result"`
}

//InformMessage message with inform information
type InformMessage struct {
	QueueMessage
	Type string    `json:"type"`
	At   time.Time `json:"at"`
}

//NewQueueMessage creates the message with id
func NewQueueMessage(id string) *QueueMessage {
	return &QueueMessage{ID: id}
}

//NewQueueMsgWithError creates the message with id and error
func NewQueueMsgWithError(id string, errMsg string) *QueueMessage {
	return &QueueMessage{ID: id, Error: errMsg}
}
