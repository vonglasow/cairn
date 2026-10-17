// Package trace holds the act: one thing that was done to a task, what it was before, what it became, when, by whom and through which surface — and the membrane anything uses to read or build one.
package trace

import "time"

// The acts that are not about an attribute a task carries.
const (
	// Created is a task coming into being.
	Created = "created"
	// Moved is a task changing place among the ones it sits with.
	Moved = "moved"
	// FocusTaken is a task becoming the one being worked on.
	FocusTaken = "focus taken"
	// FocusReleased is that task being let go.
	FocusReleased = "focus released"
	// Deleted is a task being removed.
	Deleted = "deleted"
	// Changed is one attribute of a task taking another value.
	Changed = "changed"
	// TimeAdded is time spent on a task without it being in focus, added by hand: the day it was spent is what it was, the length what it became.
	TimeAdded = "time added"
)

// Act is one thing that was done to a task.
type Act struct {
	task    int
	did     string
	field   string
	was     string
	became  string
	through string
	by      string
	at      time.Time
}

// Carried is everything an act carries, as plain facts.
type Carried struct {
	Task                             int
	Did, Field, Was, Became, Through string
	By                               string
	At                               time.Time
}

// Carried hands over what the act carries.
func (a Act) Carried() Carried {
	return Carried{Task: a.task, Did: a.did, Field: a.field, Was: a.was, Became: a.became, Through: a.through, By: a.by, At: a.at}
}

// Made is the act that carries what is given.
func Made(c Carried) Act {
	return Act{task: c.Task, did: c.Did, field: c.Field, was: c.Was, became: c.Became, through: c.Through, by: c.By, at: c.At}
}
