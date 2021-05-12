package status

//Status represents transcription status
type Status int

const (
	//Uploaded value
	Uploaded Status = iota + 1
	//AudioConvert value
	AudioConvert
	//Diarization value
	Diarization
	//Transcription
	Transcription
	Rescore
	ResultMake
	JoinResults
	Completed
)

var (
	statusName = map[Status]string{Uploaded: "UPLOADED", Completed: "COMPLETED",
		AudioConvert: "AudioConvert", Diarization: "Diarization",
		Transcription: "Transcription", Rescore: "Rescore",
		ResultMake: "ResultMake", JoinResults: "JoinResults"}
	nameStatus = map[string]Status{"UPLOADED": Uploaded, "COMPLETED": Completed,
		"AudioConvert": AudioConvert, "Diarization": Diarization,
		"Transcription": Transcription, "Rescore": Rescore,
		"ResultMake": ResultMake, "JoinResults": JoinResults}
)

func Name(st Status) string {
	return statusName[st]
}

func From(st string) Status {
	return nameStatus[st]
}

func Min(st1, st2 Status) Status {
	if st1 < st2 {
		return st1
	}
	return st2
}
