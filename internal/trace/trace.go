// Package trace holds the act: one thing that was done to a task, what it was
// before, what it became, when, and through which surface — and the membrane
// anything uses to read or build one. It knows only domain, for writing a
// moment the way a person reads one.
package trace

import "time"

// The acts that are not about an attribute a task carries. Named here because
// both the engine that records them and the text that renders them have to say
// the same words, and a word spelt twice is a word that drifts.
const (
	// Created is a task coming into being. It carries what the task started
	// as, so putting one back together never needs to know what the defaults
	// were on the day it was made.
	Created = "created"
	// Moved is a task changing place among the ones it sits with.
	Moved = "moved"
	// FocusTaken is a task becoming the one being worked on.
	FocusTaken = "focus taken"
	// FocusReleased is that task being let go.
	FocusReleased = "focus released"
	// Deleted is a task being removed. Its acts stay: otherwise the past would
	// change every time something was deleted.
	Deleted = "deleted"
	// Changed is one attribute of a task taking another value. The attribute
	// is told alongside.
	Changed = "changed"
	// TimeAdded is time spent on a task without it being in focus, added by
	// hand: the day it was spent is what it was, the length what it became.
	TimeAdded = "time added"
)

// Act is one thing that was done to a task.
//
// It names the task by number rather than carrying the task: a task may be
// gone by the time its acts are read, and an act that held one could not
// outlive it. Numbers are never handed out twice, so a number goes on naming
// the one task it always named.
type Act struct {
	task    int
	did     string
	field   string
	was     string
	became  string
	through string
	at      time.Time
}

// Told is the membrane of an act: one method per fact it has to tell, absence
// included. Whoever wants to know something about an act implements this and
// is told it; nothing reads an act's insides, and an act knows none of them.
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

// TellOrNot tells a value to a membrane, or tells its absence when there is
// none.
//
// It lives here rather than in each caller because four facts of an act are
// absent in the same way, and everyone who builds an act has the same four ifs
// to write. It was written twice before this — once in the store reading a row,
// once in the engine recording something — and two copies of one rule is one
// place for it to start saying something different.
func TellOrNot(value string, told func(string), absent func()) {
	if value == "" {
		absent()
		return
	}
	told(value)
}

// Tell says every fact this act carries to the given membrane, absence
// included, in the order a line lays them out.
func (a Act) Tell(to Told) {
	to.On(a.task)
	to.Did(a.did)
	if a.field == "" {
		to.NothingNamed()
	} else {
		to.Changing(a.field)
	}
	if a.was == "" {
		to.WasNothing()
	} else {
		to.Was(a.was)
	}
	if a.became == "" {
		to.BecameNothing()
	} else {
		to.Became(a.became)
	}
	if a.through == "" {
		to.ByHand()
	} else {
		to.Through(a.through)
	}
	to.At(a.at)
}
