package domain

import (
	"fmt"
	"time"
)

// Held is what a store, or a backup of one, contains.
type Held struct {
	Tasks      int `json:"tasks"`
	Objectives int `json:"objectives"`
}

// HeldText writes what is held the way a person reads it.
func HeldText(h Held) string {
	return fmt.Sprintf("holds %d task(s) and %d objective(s)", h.Tasks, h.Objectives)
}

// Backup is a backup of the store, kept beside it and referred to by its number.
type Backup struct {
	Number int
	Made   time.Time
	Path   string
}

// FormatMoment writes a moment the way a person reads one: local time, to the minute.
func FormatMoment(t time.Time) string {
	return t.Local().Format(momentLayout)
}

const momentLayout = "2006-01-02 15:04"

// ParseMoment reads back a moment FormatMoment wrote, in local time.
func ParseMoment(s string) (time.Time, error) {
	return time.ParseInLocation(momentLayout, s, time.Local)
}
