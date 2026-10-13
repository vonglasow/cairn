package tui

import (
	"cairn/internal/engine"
	"cairn/internal/trace"
)

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

func written(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
