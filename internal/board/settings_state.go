package board

import (
	"cairn/internal/engine/workspace"
	"cairn/internal/theme"
)

// WornScheme is the scheme the window shows: the one previewed while the chooser is open, otherwise the one in force.
func (b Board) WornScheme() string {
	if b.ChooserOpen && b.Previewing != "" {
		return b.Previewing
	}
	for _, seen := range b.Configured.Themes {
		if seen.InForce {
			return seen.Slug
		}
	}
	return theme.DefaultSlug
}

// SettingsWritten is the file a setting is written to: the global one, or this workspace's.
func (b Board) SettingsWritten() string {
	if b.Global {
		return b.Configured.File
	}
	return workspace.SettingsOf(b.Configured)
}

// Trouble is the file the window cannot read while it is open, and why: nothing when every file reads.
func (b Board) Trouble() (workspace.Trouble, bool) {
	if b.Troubled == nil {
		return workspace.Trouble{}, false
	}
	return *b.Troubled, true
}

// ReadAgain is what the window says of a configuration it read again: the file, and each key that changed with what it was.
func (b Board) ReadAgain() (string, []workspace.ConfigChange) { return b.ChangedIn, b.Changed }

// OpenAsked is the file the window wants the system to open, if any.
func (b Board) OpenAsked() (string, bool) { return b.Opens, b.Opens != "" }

// PlaceOf is where the store or the notes are, as the configuration has it.
func (b Board) PlaceOf(key string) string {
	if key == "notes" {
		return b.Configured.Notes.At
	}
	return b.Configured.Store.At
}

// InForce is the colour scheme in force.
func (b Board) InForce() string {
	for _, seen := range b.Configured.Themes {
		if seen.InForce {
			return seen.Slug
		}
	}
	return theme.DefaultSlug
}
