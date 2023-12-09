package progress_test

import (
	"testing"

	"github.com/airenas/listgo/internal/pkg/progress"
	"github.com/airenas/listgo/internal/pkg/status"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name string
		args status.Status
		want int32
	}{
		{name: "any", args: status.From("olia"), want: 0},
		{name: "ChannelsSplit", args: status.SplitChannels, want: 6},
		{name: "AudioConvert", args: status.AudioConvert, want: 7},
		{name: "Rescore", args: status.Rescore, want: 70},
		{name: "ResultMake", args: status.ResultMake, want: 90},
		{name: "Completed", args: status.Completed, want: 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := progress.Convert(tt.args); got != tt.want {
				t.Errorf("Convert() = %v, want %v", got, tt.want)
			}
		})
	}
}
