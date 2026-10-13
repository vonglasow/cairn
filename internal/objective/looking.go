package objective

import (
	"strconv"

	"cairn/internal/domain"
)

// Said is everything a search looks in, read the way a person reads an objective rather than the way it is stored.
func Said(o Objective) []string {
	c := o.Carried()
	said := []string{strconv.Itoa(c.Number), c.Title, string(c.State), string(c.Priority)}
	if c.Description != "" {
		said = append(said, c.Description)
	}
	said = append(said, c.Labels...)
	for _, s := range []string{domain.FormatDay(c.Deadline), domain.FormatDuration(c.Estimate)} {
		if s != "" {
			said = append(said, s)
		}
	}
	return said
}

// Matching says whether o carries every one of named, case-insensitive, wherever among its facts each one happens to sit.
func Matching(o Objective, named []string) bool {
	return domain.Reaches(Said(o), named)
}
