package status

//Status represents transcription status
type Status int

const (
	// Uploaded value
	Uploaded Status = iota + 1
	// AudioConvert value
	AudioConvert
	// Diarization value
	Diarization
	// Transcription value
	Transcription
	// Rescore status
	Rescore
	// ResultMake status
	ResultMake
	// JoinResults status
	JoinResults
	// Completed status
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

// Name return status as string
func Name(st Status) string {
	return statusName[st]
}

// From converts string to Status
func From(st string) Status {
	return nameStatus[st]
}

// Min selects min status of the two
func Min(st1, st2 Status) Status {
	if st1 < st2 {
		return st1
	}
	return st2
}
