package progress

import (
	"bitbucket.org/airenas/listgo/internal/pkg/status"
)

var statusProgressMap = make(map[string]int32)

func init() {
	statusProgressMap[status.Uploaded.Name] = 5
	statusProgressMap[status.AudioConvert.Name] = 6
	statusProgressMap[status.Diarization.Name] = 35
	statusProgressMap[status.Transcription.Name] = 50
	statusProgressMap[status.ResultMake.Name] = 90
	statusProgressMap[status.Completed.Name] = 100
}

//Convert return percentage value of a progress for status value
func Convert(status string) int32 {
	pr, found := statusProgressMap[status]
	if found {
		return pr
	}
	return 0
}
