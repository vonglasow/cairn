package task

import (
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/item"
	"strconv"
)

// Said is everything a search looks in, read the way a person reads a task rather than the way it is stored.
func Said(t Task) []string {
	c := t.Carried()
	said := []string{strconv.Itoa(c.Number), c.Title, string(c.State), string(c.Priority)}
	if c.Description != "" {
		said = append(said, c.Description)
	}
	said = append(said, c.Labels...)
	for _, step := range c.Steps {
		said = append(said, step.Title)
	}
	for _, s := range []string{c.Because, calendar.FormatDay(c.Deadline), calendar.FormatDuration(c.Estimate), calendar.FormatDuration(c.Spent)} {
		if s != "" {
			said = append(said, s)
		}
	}
	return said
}

// Matching says whether t carries every one of named, case-insensitive, wherever among its facts each one happens to sit.
func Matching(t Task, named []string) bool {
	return item.Reaches(Said(t), named)
}
