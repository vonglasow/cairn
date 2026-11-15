package workspace

import (
	"cairn/internal/store/files"
	"cairn/internal/theme"
	"path/filepath"
)

// ThemeSeen is one scheme config show lists: its name, where it comes from, whether the window wears it, why it cannot be worn, and every contrast it falls short of.
type ThemeSeen struct {
	Slug    string   `json:"slug"`
	From    string   `json:"from" jsonschema:"shipped, for a theme cairn carries, or the file a theme of one's own is read from"`
	InForce bool     `json:"in_force,omitempty"`
	Problem string   `json:"problem,omitempty" jsonschema:"why the theme cannot be worn"`
	Short   []string `json:"short,omitempty" jsonschema:"every contrast the theme falls short of, the tokens named and the ratio they read at"`
}

// ThemesIn is the folder of a person's own themes.
func ThemesIn(home string) string {
	return filepath.Join(filepath.Dir(files.FileIn(home)), "themes")
}

// Unworn says why a theme is not worn, and what the window wears instead.
func Unworn(err error) string { return err.Error() + "; the window wears " + theme.DefaultSlug }

// SchemeOf is the scheme config show listed, read from where it said it is: carried in the binary, or a person's own file.
func SchemeOf(seen ThemeSeen) (theme.Scheme, error) {
	if seen.From == theme.Carried {
		return theme.Find("", seen.Slug)
	}
	return theme.Find(filepath.Dir(seen.From), seen.Slug)
}
