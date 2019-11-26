package api

import "time"

//Recognizer describes recognizer for upload service response
type Recognizer struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	DateCreated time.Time `json:"date_created,omitempty"`
}
