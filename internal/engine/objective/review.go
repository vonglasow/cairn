package objective

import (
	"cairn/internal/engine/act"
	"time"
)

// MonthShape is the time spent in one month of the period, and for a parent how it splits over its sub-objectives and itself.
type MonthShape struct {
	Month         string      `json:"month" jsonschema:"the month, as 2026-08"`
	TimeSpent     string      `json:"time_spent"`
	SubObjectives []PartShape `json:"sub_objectives,omitempty" jsonschema:"the time spent through each sub-objective that month; left out for an objective holding none"`
	Itself        string      `json:"itself,omitempty" jsonschema:"the time spent on the tasks put on the parent itself that month; left out when none was"`
}

// PartShape is the time a month spent through one sub-objective.
type PartShape struct {
	Number    int    `json:"number"`
	TimeSpent string `json:"time_spent"`
}

// DoneShape is a task put in done within the period, and the day it was.
type DoneShape struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	On     string `json:"on"`
}

// MonthsOf is the time spent each month on the tasks serving.
func MonthsOf(sheet act.Sheet, serving map[int]bool) []Month {
	var months []Month
	for _, day := range sheet.Days {
		var spent time.Duration
		for _, w := range day.Tasks {
			if serving[w.Task] {
				spent += w.Took
			}
		}
		if spent == 0 {
			continue
		}
		month := time.Date(day.Day.Year(), day.Day.Month(), 1, 0, 0, 0, 0, day.Day.Location())
		if n := len(months); n > 0 && months[n-1].Month.Equal(month) {
			months[n-1].Spent += spent
			continue
		}
		months = append(months, Month{Month: month, Spent: spent})
	}
	return months
}
