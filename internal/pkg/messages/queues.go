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
)

const (
	// InformType_Started type when process started
	InformType_Started string = "Started"
	// InformType_Finished type when process started
	InformType_Finished string = "Finished"
)

const (
	//TopicStatusChange is topic name for status change event
	TopicStatusChange string = "StatusChange"
)

//ResultQueueFor creates result queus name for input queue
func ResultQueueFor(queue string) string {
	return queue + "_Result"
}
