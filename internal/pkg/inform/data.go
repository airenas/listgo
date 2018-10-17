package inform

import (
	"time"
)

//Data keeps data for email generation
type Data struct {
	ID      string
	MsgType string
	Email   string
	MsgTime time.Time
}
