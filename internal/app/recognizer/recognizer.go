package recognizer

import "time"

//Info describes recognizer
type Info struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	DateCreated time.Time         `yaml:"date_created,omitempty"`
	Settings    map[string]string `yaml:"settings,flow"`
}
