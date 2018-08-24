package messages

// Publisher publish a transcription id to some topic
type Publisher interface {
	Publish(id string, topic string) error
}
