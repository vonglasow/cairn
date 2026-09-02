package board

import (
	"time"

	"cairn/internal/domain"
	"cairn/internal/trace"
)

// Reading a history. An act tells its facts rather than handing them over, so
// the board listens the same way everything else does — through a membrane of
// its own, which is what lets it take the moment as a moment and write it the
// way a person reads one, rather than the way a machine does.

// heard is what one act said.
type heard struct {
	did    string
	field  string
	was    string
	became string
	at     time.Time
}

func (h *heard) On(int)                {}
func (h *heard) Did(what string)       { h.did = what }
func (h *heard) Changing(field string) { h.field = field }
func (h *heard) NothingNamed()         { h.field = "" }
func (h *heard) Was(value string)      { h.was = value }
func (h *heard) WasNothing()           { h.was = "" }
func (h *heard) Became(value string)   { h.became = value }
func (h *heard) BecameNothing()        { h.became = "" }
func (h *heard) Through(string)        {}
func (h *heard) ByHand()               {}
func (h *heard) At(when time.Time)     { h.at = when }

// line is one act, written the way the detail shows it: when, what was done,
// and what it went from and to when it was about something in particular.
func (b Board) line(act trace.Act) string {
	var h heard
	act.Tell(&h)
	out := b.skin.Faint.Render(domain.FormatMoment(h.at)) + "  " + b.skin.Bright.Render(h.did)
	if h.field == "" {
		return out
	}
	return out + "  " + b.skin.Says.Render(h.field) + "  " +
		b.skin.Faint.Render(written(h.was)) + b.skin.Says.Render(" → ") +
		b.skin.Bright.Render(written(h.became))
}

// written is how a value that was not there is drawn in a history.
func written(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
