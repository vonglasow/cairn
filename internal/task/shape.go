package task

import (
	"time"

	"cairn/internal/domain"
)

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

type shapeMembrane struct {
	number, parent                      int
	objectives                          []int
	title, state, priority, description string
	labels                              []string
	blockers                            []int
	because                             string
	deadline, estimate, timeSpent       string
	archived                            bool
}

func (m *shapeMembrane) Numbered(n int)  { m.number = n }
func (m *shapeMembrane) Titled(s string) { m.title = s }
func (m *shapeMembrane) Archived()       { m.archived = true }
func (m *shapeMembrane) NotArchived()    { m.archived = false }

func (m *shapeMembrane) InColumn(s domain.State)   { m.state = string(s) }
func (m *shapeMembrane) Carries(p domain.Priority) { m.priority = string(p) }
func (m *shapeMembrane) Described(s string)        { m.description = s }
func (m *shapeMembrane) NotDescribed()             {}
func (m *shapeMembrane) Labelled(s string)         { m.labels = append(m.labels, s) }
func (m *shapeMembrane) NotLabelled()              {}
func (m *shapeMembrane) LabelsTold() {
	if m.labels == nil {
		m.labels = []string{}
	}
}
func (m *shapeMembrane) BlockedBy(n int) { m.blockers = append(m.blockers, n) }
func (m *shapeMembrane) NotBlocked()     {}
func (m *shapeMembrane) BlockersTold() {
	if m.blockers == nil {
		m.blockers = []int{}
	}
}
func (m *shapeMembrane) Because(s string)          { m.because = s }
func (m *shapeMembrane) NotBecause()               {}
func (m *shapeMembrane) Due(t time.Time)           { m.deadline = domain.FormatDay(t) }
func (m *shapeMembrane) NotDue()                   {}
func (m *shapeMembrane) Estimated(d time.Duration) { m.estimate = domain.FormatDuration(d) }
func (m *shapeMembrane) NotEstimated()             {}
func (m *shapeMembrane) Spent(d time.Duration)     { m.timeSpent = domain.FormatDuration(d) }
func (m *shapeMembrane) NothingSpent()             {}
func (m *shapeMembrane) Under(n int)               { m.parent = n }
func (m *shapeMembrane) UnderNothing()             {}
func (m *shapeMembrane) Serves(n int)              { m.objectives = append(m.objectives, n) }
func (m *shapeMembrane) ServesNothing()            {}
func (m *shapeMembrane) ObjectivesTold() {
	if m.objectives == nil {
		m.objectives = []int{}
	}
}

// OfTask turns a task into the shape a machine reads.
func OfTask(t Task) Shape {
	m := &shapeMembrane{}
	t.Tell(m)
	return Shape{
		Facts: Facts{
			Number: m.number, Parent: m.parent, Title: m.title,
			State: m.state, Archived: m.archived, Priority: m.priority,
			Labels: m.labels, BlockedBy: m.blockers, Because: m.because,
			Deadline:   m.deadline,
			Estimate:   m.estimate,
			TimeSpent:  m.timeSpent,
			Objectives: m.objectives,
		},
		Description: m.description,
	}
}
