package engine

import (
	"cairn/internal/engine/workspace"
	"cairn/internal/store/files"
	"cairn/internal/theme"
)

// Wearing is the colours the window wears for the store found from h, and, when the scheme named cannot be worn, why: the default is worn instead, and a configuration that cannot be read, said where it is read for the places, leaves the default too.
func Wearing(h Home) (theme.Colours, string) {
	if h.Dir == "" {
		return theme.Default(), ""
	}
	read, _, err := h.settled()
	if err != nil {
		return theme.Default(), ""
	}
	s, err := theme.Find(workspace.ThemesIn(h.Dir), read.Theme())
	if err != nil {
		return theme.Default(), workspace.Unworn(err)
	}
	return s.Colours(), ""
}

func (g *readingGate) themes(p files.Places) {
	folder, worn := workspace.ThemesIn(g.home.Dir), p.Theme()
	g.config.Window = &workspace.WindowConfig{KeepSizes: p.KeepSizes(), Theme: p.Theme(), Intro: p.Intro(), Ghost: p.Ghost()}
	if _, err := theme.Find(folder, p.Theme()); err != nil {
		g.config.Window.Unworn, worn = workspace.Unworn(err), theme.DefaultSlug
	}
	for _, found := range theme.Seen(folder) {
		seen := workspace.ThemeSeen{Slug: found.Slug, From: found.From, InForce: found.Slug == worn}
		if found.Err != nil {
			seen.Problem = found.Err.Error()
		} else {
			for _, short := range theme.Shortfalls(found.Scheme.Colours()) {
				seen.Short = append(seen.Short, short.String())
			}
		}
		g.config.Themes = append(g.config.Themes, seen)
	}
}
