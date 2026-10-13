package engine

import (
	"cairn/internal/domain"
	"cairn/internal/trace"
)

// Said is one act of a task's history, the way a person reads it: when, what was done, what about, what it went from and to, and which surface did it when it was not the person reading.
type Said struct {
	When, Did, Field, Was, Became, Through string
}

// SaidOf is what an act says, heard through its own membrane — the only way in.
func SaidOf(act trace.Act) Said {
	c := act.Carried()
	return Said{
		When: domain.FormatMoment(c.At), Did: c.Did, Field: c.Field,
		Was: domain.Shown(c.Was), Became: domain.Shown(c.Became), Through: c.Through,
	}
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
