package task

import (
	"time"

	"cairn/internal/domain"
)

// Shape is how a task crosses to a machine. The field names are the wire, so
// they are written here once rather than derived from the entity — an entity
// that knew its wire format would drag a protocol into every test.
//
// How far along a task is has no field here: it is not the task's to say, and
// the engine's answer — the one place that reaches both notions — carries it
// instead.
//
// It is the task's facts and what the task says, kept apart: a listing hands
// over the facts of every task and leaves what each says to be shown, and a
// shape that held both in one list of fields could only be cut down after the
// fact, by a second list kept in step with this one.
type Shape struct {
	Facts
	Description string `json:"description"`
}

// Facts is everything a task carries but what it says: the part of it a
// listing hands over for every row.
type Facts struct {
	Number int    `json:"number"`
	Parent int    `json:"parent"`
	Title  string `json:"title"`
	State  string `json:"state"`
	// Archived is beside the state, so a machine is told both.
	Archived   bool     `json:"archived"`
	Priority   string   `json:"priority"`
	Labels     []string `json:"labels"`
	BlockedBy  []int    `json:"blocked_by"`
	Because    string   `json:"because"`
	Deadline   string   `json:"deadline"`
	Estimate   string   `json:"estimate"`
	TimeSpent  string   `json:"time_spent"`
	Objectives []int    `json:"objectives"`
}

// shapeMembrane is a task's membrane for the shape a machine reads: it is
// told every fact and keeps it in the form Shape needs, ready to be read off
// once telling is done.
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
