package status

import "bitbucket.org/airenas/listgo/internal/app/status/api"

// Provider provides transcription result for ID
type Provider interface {
	Get(ID string) (*api.TranscriptionResult, error)
}
