package engine

import (
	"cairn/internal/domain"
	"cairn/internal/trace"
)

// said is one act of a task's history, the way a person reads it: when, what was done, what about, what it went from and to, who did it and through which surface.
type said struct {
	When, Did, Field, Was, Became, Through, By string
}

// SaidOf is what an act says, heard through its own membrane — the only way in.
func SaidOf(act trace.Act) said {
	c := act.Carried()
	return said{
		When: domain.FormatMoment(c.At), Did: c.Did, Field: c.Field,
		Was: domain.Shown(c.Was), Became: domain.Shown(c.Became), Through: c.Through, By: c.By,
	}
}

// Line is the act on one line, the way a board writes it when it has no ink to tell the pieces apart with.
func (s said) Line() string {
	out := s.When + "  " + s.Did
	if s.Field != "" {
		out += "  " + s.Field + "  " + domain.OrNothing(s.Was) + " → " + domain.OrNothing(s.Became)
	} else if s.Became != "" {
		out += "  " + s.Became
	}
	return out + trace.Credited(s.By, s.Through)
}
