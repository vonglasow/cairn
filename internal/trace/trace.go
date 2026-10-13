// Package trace holds the act: one thing that was done to a task, what it was before, what it became, when, and through which surface — and the membrane anything uses to read or build one.
package trace

import (
	"time"

	"cairn/internal/domain"
)

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
	at      time.Time
}

// Told is the membrane of an act: one method per fact it has to tell, absence included.
type Told interface {
	On(int)
	Did(string)
	Changing(string)
	NothingNamed()
	Was(string)
	WasNothing()
	Became(string)
	BecameNothing()
	Through(string)
	ByHand()
	At(time.Time)
}

// Tell says every fact this act carries to the given membrane, absence included, in the order a line lays them out.
func (a Act) Tell(to Told) {
	to.On(a.task)
	to.Did(a.did)
	domain.TellOrNot(a.field, to.Changing, to.NothingNamed)
	domain.TellOrNot(a.was, to.Was, to.WasNothing)
	domain.TellOrNot(a.became, to.Became, to.BecameNothing)
	domain.TellOrNot(a.through, to.Through, to.ByHand)
	to.At(a.at)
}
