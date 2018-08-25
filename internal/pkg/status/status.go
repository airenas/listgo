package status

//Status represents transcription status
type Status struct {
	Name string
}

//Uploaded - file is in server for transcription
var Uploaded = Status{"UPLOADED"}

//Completed - finished
var Completed = Status{"COMPLETED"}

//AudioConvert in progress
var AudioConvert = Status{"AudioConvert"}

//Diarization in progress
var Diarization = Status{"Diarization"}

//Transcription in progress
var Transcription = Status{"Transcription"}

//ResultMake indicates preparing result in progress
var ResultMake = Status{"ResultMake"}
