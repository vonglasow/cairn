package trace

import "time"

// Shape is how an act crosses to a machine. Every fact a line shows, and the
// moment as a machine reads one rather than as a person does.
type Shape struct {
	Task    int    `json:"task"`
	Did     string `json:"did"`
	Field   string `json:"field,omitempty"`
	Was     string `json:"was,omitempty"`
	Became  string `json:"became,omitempty"`
	Through string `json:"through,omitempty"`
	At      string `json:"at"`
}

// OfAct turns an act into what crosses.
func OfAct(a Act) Shape {
	var m shapeMembrane
	a.Tell(&m)
	return m.shape()
}

// shapeMembrane is an act's membrane for what crosses to a machine: it is told
// every fact and keeps them as the strings JSON carries.
type shapeMembrane struct {
	task                             int
	did, field, was, became, through string
	at                               time.Time
}

func (m *shapeMembrane) On(n int)               { m.task = n }
func (m *shapeMembrane) Did(what string)        { m.did = what }
func (m *shapeMembrane) Changing(field string)  { m.field = field }
func (m *shapeMembrane) NothingNamed()          { m.field = "" }
func (m *shapeMembrane) Was(value string)       { m.was = value }
func (m *shapeMembrane) WasNothing()            { m.was = "" }
func (m *shapeMembrane) Became(value string)    { m.became = value }
func (m *shapeMembrane) BecameNothing()         { m.became = "" }
func (m *shapeMembrane) Through(surface string) { m.through = surface }
func (m *shapeMembrane) ByHand()                { m.through = "" }
func (m *shapeMembrane) At(when time.Time)      { m.at = when }

// shape lays out what was told. The moment goes out in the one format a
// machine reads without guessing, never the one a person reads.
func (m *shapeMembrane) shape() Shape {
	return Shape{
		Task: m.task, Did: m.did, Field: m.field, Was: m.was,
		Became: m.became, Through: m.through, At: m.at.UTC().Format(time.RFC3339),
	}
}
