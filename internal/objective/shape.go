package objective

import (
	"time"

	"cairn/internal/domain"
)

// Shape is how an objective crosses to a machine.
type Shape struct {
	Number      int      `json:"number" jsonschema:"the objective's number, never handed out twice"`
	Title       string   `json:"title" jsonschema:"what the objective is called"`
	State       string   `json:"state" jsonschema:"where it stands: active, on-hold, reached or dropped"`
	Priority    string   `json:"priority" jsonschema:"low, normal, high or critical"`
	Description string   `json:"description" jsonschema:"what is written about it; empty when nothing is"`
	Labels      []string `json:"labels" jsonschema:"the labels it carries, lower case and sorted; empty or null when it carries none"`
	Deadline    string   `json:"deadline" jsonschema:"the day it is due, as 2026-12-01; empty when it has none"`
	Estimate    string   `json:"estimate" jsonschema:"how long it was expected to take, as 40h; empty when nobody said"`
}

type shapeMembrane struct {
	number                              int
	title, state, priority, description string
	labels                              []string
	deadline, estimate                  string
}

func (m *shapeMembrane) Numbered(n int)            { m.number = n }
func (m *shapeMembrane) Titled(s string)           { m.title = s }
func (m *shapeMembrane) Standing(s State)          { m.state = string(s) }
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
func (m *shapeMembrane) Due(t time.Time)           { m.deadline = domain.FormatDay(t) }
func (m *shapeMembrane) NotDue()                   {}
func (m *shapeMembrane) Estimated(d time.Duration) { m.estimate = domain.FormatDuration(d) }
func (m *shapeMembrane) NotEstimated()             {}

// OfObjective turns an objective into the shape a machine reads.
func OfObjective(o Objective) Shape {
	m := &shapeMembrane{}
	o.Tell(m)
	return Shape{
		Number: m.number, Title: m.title, State: m.state,
		Priority: m.priority, Description: m.description, Labels: m.labels,
		Deadline: m.deadline,
		Estimate: m.estimate,
	}
}
