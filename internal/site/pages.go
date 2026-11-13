// Package site builds cairn's documentation site: every page of docs/, written
// for a person using cairn, and a reference read off cairn itself.
package site

import (
	"path"
	"strings"
)

type source struct{ from, group, top string }

var published = []source{
	{"INSTALL.md", "Start", "install"},
	{"docs/usage.md", "Start", "guide"},
	{"docs/tui.md", "Guide", "guide"},
	{"docs/gui.md", "Guide", "guide"},
	{"docs/labels.md", "Guide", "guide"},
	{"docs/listing.md", "Guide", "guide"},
	{"docs/ordering.md", "Guide", "guide"},
	{"docs/steps.md", "Guide", "guide"},
	{"docs/blocking.md", "Guide", "guide"},
	{"docs/links.md", "Guide", "guide"},
	{"docs/objectives.md", "Guide", "guide"},
	{"docs/series.md", "Guide", "guide"},
	{"docs/trace.md", "Guide", "guide"},
	{"docs/context.md", "Guide", "guide"},
	{"docs/notes-obsidian.md", "Guide", "guide"},
	{"docs/claude-code.md", "Guide", "guide"},
	{"docs/mcp.md", "Reference", "reference"},
	{"docs/config.md", "Reference", "config"},
	{"docs/decisions.md", "Project", "decisions"},
}

const (
	commandsPage  = "commands.html"
	toolsPage     = "tools.html"
	themesPage    = "themes.html"
	changelogPage = "changelog.html"
	styleSheet    = "site.css"
	searchIndex   = "search.js"
	script        = "site.js"
)

var groups = []string{"Start", "Guide", "Reference", "Project"}

func pageFor(from string) (string, bool) {
	for _, s := range published {
		if s.from == from {
			return nameOf(from), true
		}
	}
	return "", false
}

func nameOf(from string) string {
	return strings.ToLower(strings.TrimSuffix(path.Base(from), ".md")) + ".html"
}

func windowPicture(slug string) string { return "shots/window-" + slug + ".png" }
