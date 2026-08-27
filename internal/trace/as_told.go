package trace

import "time"

// AsTold builds an act from what it is told. It is the same membrane the other
// way round: whoever has an act's facts — a store reading a row it wrote, or
// the engine recording something that just happened — tells them, and is handed
// the act they make. Nothing has to reach inside an act to put one together,
// which is what lets an act keep its insides to itself.
type AsTold struct {
	act Act
}

// Act hands back what was built.
func (a *AsTold) Act() Act { return a.act }

// On is told which task the act was performed on.
func (a *AsTold) On(n int) { a.act.task = n }

// Did is told what was done.
func (a *AsTold) Did(what string) { a.act.did = what }

// Changing is told which attribute of the task the act is about.
func (a *AsTold) Changing(field string) { a.act.field = field }

// NothingNamed is told the act is about no attribute in particular.
func (a *AsTold) NothingNamed() { a.act.field = "" }

// Was is told what the attribute held before.
func (a *AsTold) Was(value string) { a.act.was = value }

// WasNothing is told the attribute held nothing before.
func (a *AsTold) WasNothing() { a.act.was = "" }

// Became is told what the attribute holds after, or what the act carries when
// it is about no attribute — where a task was moved to, what it started as.
func (a *AsTold) Became(value string) { a.act.became = value }

// BecameNothing is told the attribute holds nothing after.
func (a *AsTold) BecameNothing() { a.act.became = "" }

// Through is told which surface performed the act.
func (a *AsTold) Through(surface string) { a.act.through = surface }

// ByHand is told the act was performed at the terminal, which a history says
// nothing about.
func (a *AsTold) ByHand() { a.act.through = "" }

// At is told when the act happened.
func (a *AsTold) At(when time.Time) { a.act.at = when }
