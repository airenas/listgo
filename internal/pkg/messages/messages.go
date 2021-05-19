package messages

import (
	"time"
)

//Tag keeps key/value in message
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

const (
	//TagTimestamp transcription add time in time.Unix()
	TagTimestamp = "timestamp"
	//TagNumberOfSpeakers number of speakers arameter
	TagNumberOfSpeakers = "number_of_speakers"
	//TagSkipNumJoin skips joining of numbers in final transcription
	TagSkipNumJoin = "skip_num_join"
	//TagParentID id of parent transcription
	TagParentID = "parent_id"
	//TagResultQueue //TagStatusQueue is queue name for result message
	TagResultQueue = "result_queue"
	//TagStatusQueue is queue name for status messages
	TagStatusQueue = "status_queue"
	//TagChildIDS contains IDs of childrens transcriptions
	TagChildIDS = "ch_ids"
	//TagChildIDSFileNames contains IDs=FileName mappings of childrens transcriptions
	TagChildIDSFileNames = "ch_ids_fn"
)

//QueueMessage message going throuht broker
type QueueMessage struct {
	ID         string `json:"id"`
	Recognizer string `json:"recognizer"`
	Tags       []Tag  `json:"tags,omitempty"`
	Error      string `json:"error,omitempty"`
}

//ResultMessage message going throuht broker with result
type ResultMessage struct {
	QueueMessage
	Result string `json:"result,omitempty"`
}

//InformMessage message with inform information
type InformMessage struct {
	QueueMessage
	Type string    `json:"type"`
	At   time.Time `json:"at"`
}

//RegistrationMessage message for registering worker
type RegistrationMessage struct {
	Queue     string `json:"queue"`
	Timestamp int64  `json:"timestamp"` //time.Unix in seconds
	Working   bool   `json:"working"`
	Type      string `json:"type"` // see RgrTypeXxx consts
}

const (
	//RgrTypeRegister indicates registration
	RgrTypeRegister = "Register"
	//RgrTypeExit stops worker
	RgrTypeExit = "Exit"
	//RgrTypeBeat alive beat
	RgrTypeBeat = "Beat"
)

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

//GetTag retrieves tag value from tag list
func GetTag(tags []Tag, key string) (string, bool) {
	for _, t := range tags {
		if t.Key == key {
			return t.Value, true
		}
	}
	return "", false
}
