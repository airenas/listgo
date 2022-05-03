package messages

const (
	// Decode queue
	Decode string = "Decode"
	// Inform queue
	Inform string = "Inform"
	// AudioConvert queue
	AudioConvert string = "AudioConvert"
	// Diarization queue
	Diarization string = "Diarization"
	// Transcription queue
	Transcription string = "Transcription"
	// Rescore queue
	Rescore string = "Rescore"
	// ResultMake queue
	ResultMake string = "ResultMake"

	// DecodeMultiple queue
	DecodeMultiple string = "DecodeMultiple"
	// JoinAudio queue
	JoinAudio string = "JoinAudio"
	// JoinResults queue
	JoinResults string = "JoinResults"
	// OneCompleted queue
	OneCompleted string = "OneCompleted"
	// OneStatus queue
	OneStatus string = "OneStatus"
)

const (
	// InformTypeStarted type when process started
	InformTypeStarted string = "Started"
	// InformTypeFinished type when process finished
	InformTypeFinished string = "Finished"
	// InformTypeFailed type when process failed
	InformTypeFailed string = "Failed"
)

const (
	//TopicStatusChange is topic name for status change event
	TopicStatusChange string = "StatusChange"
)

//ResultQueueFor creates result queus name for input queue
func ResultQueueFor(queue string) string {
	return queue + "_Result"
}
