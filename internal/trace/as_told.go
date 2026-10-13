package trace

import "time"

// AsTold builds an act one fact at a time.
type AsTold struct{ c Carried }

// From starts a builder on what an act already carries.
func From(a Act) AsTold { return AsTold{c: a.Carried()} }

// Act hands back what was built.
func (a *AsTold) Act() Act { return Made(a.c) }

// On is told which task the act was performed on.
func (a *AsTold) On(n int) { a.c.Task = n }

// Did is told what was done.
func (a *AsTold) Did(what string) { a.c.Did = what }

// Changing is told which attribute of the task the act is about.
func (a *AsTold) Changing(field string) { a.c.Field = field }

// Was is told what the attribute held before.
func (a *AsTold) Was(value string) { a.c.Was = value }

// Became is told what the attribute holds after.
func (a *AsTold) Became(value string) { a.c.Became = value }

// Through is told which surface performed the act.
func (a *AsTold) Through(surface string) { a.c.Through = surface }

// At is told when the act happened.
func (a *AsTold) At(when time.Time) { a.c.At = when }
