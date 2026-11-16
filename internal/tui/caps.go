package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

const keyColumn = 14

func (b Board) cap(keys string) string {
	caps := strings.Fields(keys)
	for i, key := range caps {
		caps[i] = b.skin.Cap.Render(" " + key + " ")
	}
	return strings.Join(caps, " ")
}

func (b Board) sentence(said string, ink lipgloss.Style) string {
	runs := strings.Split(said, "`")
	for i, run := range runs {
		if i%2 == 1 {
			runs[i] = b.cap(run)
			continue
		}
		if run != "" {
			runs[i] = ink.Render(run)
		}
	}
	return strings.Join(runs, "")
}
