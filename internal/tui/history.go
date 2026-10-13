package tui

import (
	"cairn/internal/engine"
	"cairn/internal/trace"
)

// line is one act, written the way the detail shows it: when, what was done,
// what it went from and to when it was about something in particular, and who
// did it when it was not the person engine. What the act says is
// engine.SaidOf's to hear; how each piece is inked is the board's.
func (b Board) line(act trace.Act) string {
	said := engine.SaidOf(act)
	out := b.skin.Faint.Render(said.When) + "  " + b.skin.Bright.Render(said.Did)
	if said.Field != "" {
		out += "  " + b.skin.Says.Render(said.Field) + "  " +
			b.skin.Faint.Render(written(said.Was)) + b.skin.Says.Render(" → ") +
			b.skin.Bright.Render(written(said.Became))
	}
	if said.Through != "" {
		out += "  " + b.skin.Faint.Render("("+said.Through+")")
	}
	return out
}

// written is how a value that was not there is drawn in a history.
func written(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
