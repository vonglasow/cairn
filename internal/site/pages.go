// Package site builds cairn's documentation site: every page of docs/, written
// for a person using cairn, and a reference read off cairn itself.
package site

import (
	"path"
	"strings"
)

type source struct{ from, group, top string }

var published = []source{
	{"docs/install.md", "Start", "start"},
	{"docs/usage.md", "Start", "start"},
	{"docs/glossary.md", "Start", "start"},
	{"docs/plan-a-week.md", "Guides", "guides"},
	{"docs/follow-an-objective.md", "Guides", "guides"},
	{"docs/plug-in-an-agent.md", "Guides", "guides"},
	{"docs/ordering.md", "Guides", "guides"},
	{"docs/steps.md", "Guides", "guides"},
	{"docs/blocking.md", "Guides", "guides"},
	{"docs/links.md", "Guides", "guides"},
	{"docs/labels.md", "Guides", "guides"},
	{"docs/listing.md", "Guides", "guides"},
	{"docs/objectives.md", "Guides", "guides"},
	{"docs/series.md", "Guides", "guides"},
	{"docs/trace.md", "Guides", "guides"},
	{"docs/context.md", "Guides", "guides"},
	{"docs/notes-obsidian.md", "Guides", "guides"},
	{"docs/config.md", "Guides", "guides"},
	{"docs/tui.md", "Boards", "boards"},
	{"docs/gui.md", "Boards", "boards"},
	{"docs/claude-code.md", "Agents", "agents"},
	{"docs/mcp.md", "Agents", "agents"},
}

type guide struct{ from, page, shows string }

var byTask = []guide{
	{"docs/plan-a-week.md", "time", "The week in the window: the time each day took"},
	{"docs/follow-an-objective.md", "objectives", "The objectives in the window, each with how far along it is"},
	{"docs/plug-in-an-agent.md", "tasks", "The board in the window, the task in focus first"},
}

const (
	commandsPage = "commands.html"
	toolsPage    = "tools.html"
	themesPage   = "themes.html"
	keysPage     = "keys.html"
	styleSheet   = "site.css"
	searchIndex  = "search.js"
	script       = "site.js"
)

var groups = []string{"Start", "Guides", "Boards", "Agents", "Reference"}

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

func guidePicture(from string) string {
	return "shots/guide-" + strings.TrimSuffix(nameOf(from), ".html") + ".png"
}
