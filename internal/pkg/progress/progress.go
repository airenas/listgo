package progress

import (
	"bitbucket.org/airenas/listgo/internal/pkg/status"
)

var statusProgressMap = make(map[status.Status]int32)

func init() {
	statusProgressMap[status.Uploaded] = 5
	statusProgressMap[status.SplitChannels] = 6
	statusProgressMap[status.AudioConvert] = 7
	statusProgressMap[status.Diarization] = 35
	statusProgressMap[status.Transcription] = 50
	statusProgressMap[status.Rescore] = 70
	statusProgressMap[status.ResultMake] = 90
	statusProgressMap[status.JoinResults] = 95
	statusProgressMap[status.Completed] = 100
}

//Convert return percentage value of a progress for status value
func Convert(status status.Status) int32 {
	pr, found := statusProgressMap[status]
	if found {
		return pr
	}
	return 0
}
