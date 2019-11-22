package messages

import (
	"time"
)

//Tag keeps key/value in message
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

//QueueMessage message going throuht broker
type QueueMessage struct {
	ID         string `json:"id"`
	Recognizer string `json:"recognizer"`
	Tags       []Tag  `json:"tags"`
	Error      string `json:"error"`
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

//NewQueueMessageFromM copies message
func NewQueueMessageFromM(m *QueueMessage) *QueueMessage {
	return &QueueMessage{ID: m.ID, Recognizer: m.Recognizer, Tags: m.Tags}
}

//NewQueueMessage creates the message
func NewQueueMessage(id string, rec string, tags []Tag) *QueueMessage {
	return &QueueMessage{ID: id, Recognizer: rec, Tags: tags}
}

//NewQueueMsgWithError creates the message with id and error
func NewQueueMsgWithError(id string, errMsg string) *QueueMessage {
	return &QueueMessage{ID: id, Error: errMsg}
}

//NewTag creates new tag
func NewTag(key string, value string) Tag {
	return Tag{Key: key, Value: value}
}
