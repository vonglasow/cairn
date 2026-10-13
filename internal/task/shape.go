package task

import "cairn/internal/domain"

// Shape is how a task crosses to a machine.
type Shape struct {
	Facts
	Description string `json:"description" jsonschema:"what is written about the task; empty when nothing is"`
}

// Facts is everything a task carries but what it says: the part of it a listing hands over for every row.
type Facts struct {
	Number     int      `json:"number" jsonschema:"the task's number, never handed out twice"`
	Parent     int      `json:"parent" jsonschema:"the number of the task this one is a part of; 0 when it is a task in its own right"`
	Title      string   `json:"title" jsonschema:"what the task is called"`
	State      string   `json:"state" jsonschema:"the column it stands in: backlog, todo, current, blocked, watching, paused or done"`
	Archived   bool     `json:"archived" jsonschema:"true when it was archived: kept, and off the columns"`
	Priority   string   `json:"priority" jsonschema:"low, normal, high or critical"`
	Labels     []string `json:"labels" jsonschema:"the labels it carries, lower case and sorted; empty or null when it carries none"`
	BlockedBy  []int    `json:"blocked_by" jsonschema:"the numbers of the tasks in its way; empty or null when none is"`
	Because    string   `json:"because" jsonschema:"what is in its way that is not a task cairn holds; empty when nothing is said"`
	Deadline   string   `json:"deadline" jsonschema:"the last day it may be done, as 2026-09-01; empty when it has none"`
	Estimate   string   `json:"estimate" jsonschema:"how long it was expected to take, as 1h30m; empty when nobody said"`
	TimeSpent  string   `json:"time_spent" jsonschema:"how long has been spent on it, as 1h30m; empty when nothing has"`
	Objectives []int    `json:"objectives" jsonschema:"the numbers of the objectives it serves; empty or null when it serves none"`
}

// OfTask turns a task into the shape a machine reads.
func OfTask(t Task) Shape {
	c := t.Carried()
	return Shape{
		Facts: Facts{
			Number: c.Number, Parent: c.Parent, Title: c.Title,
			State: string(c.State), Archived: c.Archived, Priority: string(c.Priority),
			Labels: orNone(c.Labels), BlockedBy: orNone(c.Blockers), Because: c.Because,
			Deadline:   domain.FormatDay(c.Deadline),
			Estimate:   domain.FormatDuration(c.Estimate),
			TimeSpent:  domain.FormatDuration(c.Spent),
			Objectives: orNone(c.Objectives),
		},
		Description: c.Description,
	}
}

func orNone[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
