package messages

const (
	// Decode queue
	Decode string = "Decode"
	// StartedDecode queue
	StartedDecode string = "StartedDecode"
	// AudioConvert queue
	AudioConvert string = "AudioConvert"
	// Diarization queue
	Diarization string = "Diarization"
	// Transcription queue
	Transcription string = "Transcription"
	// ResultMake queue
	ResultMake string = "ResultMake"
	// FinishDecode queue
	FinishDecode string = "FinishDecode"
)

const (
	//TopicStatusChange is topic name for status change event
	TopicStatusChange string = "StatusChange"
)

//ResultQueueFor creates result queus name for input queue
func ResultQueueFor(queue string) string {
	return queue + "_Result"
}
