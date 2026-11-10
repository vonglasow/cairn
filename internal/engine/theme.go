package engine

import (
	"path/filepath"

	"cairn/internal/config"
	"cairn/internal/theme"
)

// ThemeSeen is one scheme config show lists: its name, where it comes from, whether the window wears it, why it cannot be worn, and every contrast it falls short of.
type ThemeSeen struct {
	Slug    string   `json:"slug"`
	From    string   `json:"from" jsonschema:"shipped, for a scheme cairn carries, or the file a scheme of one's own is read from"`
	InForce bool     `json:"in_force,omitempty"`
	Problem string   `json:"problem,omitempty" jsonschema:"why the scheme cannot be worn"`
	Short   []string `json:"short,omitempty" jsonschema:"every contrast the scheme falls short of, the tokens named and the ratio they read at"`
}

func themesIn(home string) string {
	return filepath.Join(filepath.Dir(config.FileIn(home)), "themes")
}

func unworn(err error) string { return err.Error() + "; the window wears " + theme.DefaultSlug }

// Wearing is the colours the window wears for the store found from h, and, when the scheme named cannot be worn, why: the default is worn instead, and a configuration that cannot be read, said where it is read for the places, leaves the default too.
func Wearing(h Home) (theme.Colours, string) {
	if h.Dir == "" {
		return theme.Default(), ""
	}
	read, _, err := h.settled()
	if err != nil {
		return theme.Default(), ""
	}
	s, err := theme.Find(themesIn(h.Dir), read.Theme())
	if err != nil {
		return theme.Default(), unworn(err)
	}
	return s.Colours(), ""
}

// SchemeOf is the scheme config show listed, read from where it said it is: carried in the binary, or a person's own file.
func SchemeOf(seen ThemeSeen) (theme.Scheme, error) {
	if seen.From == theme.Carried {
		return theme.Find("", seen.Slug)
	}
	return theme.Find(filepath.Dir(seen.From), seen.Slug)
}

func (g *readingGate) themes(p config.Places) {
	folder, worn := themesIn(g.home.Dir), p.Theme()
	g.config.Window = &WindowConfig{KeepSizes: p.KeepSizes(), Theme: p.Theme()}
	if _, err := theme.Find(folder, p.Theme()); err != nil {
		g.config.Window.Unworn, worn = unworn(err), theme.DefaultSlug
	}
	for _, found := range theme.Seen(folder) {
		seen := ThemeSeen{Slug: found.Slug, From: found.From, InForce: found.Slug == worn}
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
