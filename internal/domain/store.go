package domain

import (
	"fmt"
	"time"
)

// Held is what a store, or a backup of one, contains. Tasks and objectives are
// counted apart for the same reason they are everywhere else: they are not the
// same thing, and one total would hide which of the two somebody is about to
// overwrite. The tags are the wire as much as the fields are: a machine reads
// this shape directly, on its own or nested under a backup's.
type Held struct {
	Tasks      int `json:"tasks"`
	Objectives int `json:"objectives"`
}

// HeldText writes what is held the way a person reads it. The zero is written
// rather than left out: this is an inventory, and a line missing at zero reads
// as a kind of thing that could not have been held at all.
func HeldText(h Held) string {
	return fmt.Sprintf("holds %d task(s) and %d objective(s)", h.Tasks, h.Objectives)
}

// Backup is a backup of the store, kept beside it and referred to by its number.
// It lives here rather than beside the code that writes it, because a surface
// has to name one without being allowed to know how a store is made.
type Backup struct {
	Number int
	Made   time.Time
	Path   string
}

// FormatMoment writes a moment the way a person reads one: local time, to the
// minute. A backup and a change to the store are both dated with it, so the
// format lives here once rather than in each place that reads one.
func FormatMoment(t time.Time) string {
	return t.Local().Format(momentLayout)
}

// momentLayout is how a moment is written for a person: local time, to the
// minute.
const momentLayout = "2006-01-02 15:04"

// ParseMoment reads back a moment FormatMoment wrote, in local time.
func ParseMoment(s string) (time.Time, error) {
	return time.ParseInLocation(momentLayout, s, time.Local)
}
