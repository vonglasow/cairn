package engine

import (
	"time"

	"cairn/internal/domain"
	"cairn/internal/trace"
)

// Said is one act of a task's history, the way a person reads it: when, what was done, what about, what it went from and to, and which surface did it when it was not the person reading.
type Said struct {
	When, Did, Field, Was, Became, Through string
}

// SaidOf is what an act says, heard through its own membrane — the only way in.
func SaidOf(act trace.Act) Said {
	var h saying
	act.Tell(&h)
	return h.said
}

// Line is the act on one line, the way a board writes it when it has no ink to tell the pieces apart with.
func (s Said) Line() string {
	out := s.When + "  " + s.Did
	if s.Field != "" {
		out += "  " + s.Field + "  " + domain.OrNothing(s.Was) + " → " + domain.OrNothing(s.Became)
	}
	if s.Through != "" {
		out += "  (" + s.Through + ")"
	}
	return out
}

type saying struct{ said Said }

func (h *saying) On(int)                {}
func (h *saying) Did(what string)       { h.said.Did = what }
func (h *saying) Changing(field string) { h.said.Field = field }
func (h *saying) NothingNamed()         {}
func (h *saying) Was(value string)      { h.said.Was = domain.Shown(value) }
func (h *saying) WasNothing()           {}
func (h *saying) Became(value string)   { h.said.Became = domain.Shown(value) }
func (h *saying) BecameNothing()        {}
func (h *saying) Through(who string)    { h.said.Through = who }
func (h *saying) ByHand()               {}
func (h *saying) At(when time.Time)     { h.said.When = domain.FormatMoment(when) }
