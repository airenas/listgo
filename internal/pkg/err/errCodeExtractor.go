package err

import (
	"strings"
)

const (
	// DefaultCode is a default service error code
	DefaultCode    string = "SERVICE_ERROR"
	errorCodeStart string = "[[[ErrorCode:"
	errorCodeEnd   string = "]]]"
)

//CodeExtractor get the error code from erro message
type CodeExtractor struct {
}

//Get searches for [[[ErrCode:xxx]]] in string and returns xxx or SERVICE_ERROR
func (ece CodeExtractor) Get(err string) string {
	i := strings.Index(err, errorCodeStart)
	if i > -1 {
		ec := err[i+len(errorCodeStart):]
		ie := strings.Index(ec, errorCodeEnd)
		if ie > -1 {
			ec = strings.TrimSpace(ec[:ie])
			if len(ec) > 0 {
				return ec
			}
		}
	}
	return DefaultCode
}
