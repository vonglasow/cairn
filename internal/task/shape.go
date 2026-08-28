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
// wire — the one place that reaches both notions — carries it instead.
type Shape struct {
	Number      int      `json:"number"`
	Parent      int      `json:"parent"`
	Title       string   `json:"title"`
	State       string   `json:"state"`
	Priority    string   `json:"priority"`
	Description string   `json:"description"`
	Labels      []string `json:"labels"`
	BlockedBy   []int    `json:"blocked_by"`
	Because     string   `json:"because"`
	Deadline    string   `json:"deadline"`
	Estimate    string   `json:"estimate"`
	TimeSpent   string   `json:"time_spent"`
	Objective   int      `json:"objective"`
}

// shapeMembrane is a task's membrane for the shape a machine reads: it is
// told every fact and keeps it in the form Shape needs, ready to be read off
// once telling is done.
type shapeMembrane struct {
	number, parent, objective           int
	title, state, priority, description string
	labels                              []string
	blockers                            []int
	because                             string
	deadline, estimate, timeSpent       string
}

func (m *shapeMembrane) Numbered(n int)            { m.number = n }
func (m *shapeMembrane) Titled(s string)           { m.title = s }
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
func (m *shapeMembrane) Serves(n int)              { m.objective = n }
func (m *shapeMembrane) ServesNothing()            {}

// OfTask turns a task into the shape a machine reads.
func OfTask(t Task) Shape {
	m := &shapeMembrane{}
	t.Tell(m)
	return Shape{
		Number: m.number, Parent: m.parent, Title: m.title,
		State: m.state, Priority: m.priority,
		Description: m.description, Labels: m.labels,
		BlockedBy: m.blockers, Because: m.because,
		Deadline:  m.deadline,
		Estimate:  m.estimate,
		TimeSpent: m.timeSpent,
		Objective: m.objective,
	}
}
