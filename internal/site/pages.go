// Package site builds cairn's documentation site: every page of docs/, written
// for a person using cairn, and a reference read off cairn itself.
package site

import (
	"path"
	"strings"
)

type source struct{ from, group, says string }

var published = []source{
	{"docs/download.md", "Get started", "Homebrew or an archive, on macOS, Linux or Windows."},
	{"docs/usage.md", "Get started", "A task, its column, its focus; the board in a window and in the terminal."},
	{"docs/glossary.md", "Get started", "Task, column, objective, focus, series: one word for each, everywhere."},
	{"docs/plan-a-week.md", "Guides", "Choose the week's tasks and their order, set deadlines, focus, and see how long it took."},
	{"docs/follow-an-objective.md", "Guides", "Say what the work is for, split it, and read how far along it is."},
	{"docs/plug-in-an-agent.md", "Guides", "Claude Code or any MCP client on your board, and how far it may reach."},
	{"docs/ordering.md", "Using cairn", "Where a task stands, the order you set, and the task you focus."},
	{"docs/steps.md", "Using cairn", "Split a task into pieces to tick."},
	{"docs/blocking.md", "Using cairn", "A task in the way or a reason said; blocked, watching, paused; starting again."},
	{"docs/links.md", "Using cairn", "Tasks that go together, and a new task for a defect found after done."},
	{"docs/labels.md", "Using cairn", "Words on tasks and objectives, and their colours."},
	{"docs/listing.md", "Using cairn", "Narrow a list, search it, keep the filters under a name."},
	{"docs/objectives.md", "Using cairn", "What work is for, sub-objectives, what one waits on, its report."},
	{"docs/series.md", "Using cairn", "A task that comes back, by a rule you write in words."},
	{"docs/trace.md", "Using cairn", "What happened to a task, who did it, and the time each day took."},
	{"docs/context.md", "Using cairn", "The folder beside each task, for what was done and decided."},
	{"docs/notes-obsidian.md", "Using cairn", "Read and write the same notes from an Obsidian vault."},
	{"docs/gui.md", "The boards", ""},
	{"docs/tui.md", "The boards", ""},
	{"docs/claude-code.md", "Agents", ""},
	{"docs/mcp.md", "Agents", ""},
	{"docs/config.md", "Settings", ""},
	{"docs/backups.md", "Settings", ""},
}

type picture struct{ at, of, shows string }

type guide struct {
	from     string
	pictures []picture
}

const columns = "columns"

var byTask = []guide{
	{"docs/plan-a-week.md", []picture{
		{"", columns, "The week's columns in the window: todo, then current"},
		{"look-back", "time", "The week in the window: the time each day took"},
	}},
	{"docs/follow-an-objective.md", []picture{{"", "objectives", "The objectives in the window, each with how far along it is"}}},
	{"docs/plug-in-an-agent.md", []picture{{"", "tasks", "The board in the window, the task in focus first"}}},
}

var shots = []string{"tasks", "time", "objectives"}

const (
	homePage     = "index.html"
	docsPage     = "docs.html"
	agentsPage   = "agents.html"
	downloadPage = "download.html"
	commandsPage = "commands.html"
	toolsPage    = "tools.html"
	themesPage   = "themes.html"
	keysPage     = "keys.html"
	styleSheet   = "site.css"
	searchIndex  = "search.js"
	script       = "site.js"
)

var groups = []string{"Get started", "Guides", "Using cairn", "The boards", "Agents", "Settings", "Reference"}

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

func shotPicture(of string) string { return "shots/" + of + ".png" }

func drawing(of, shows string) string {
	if of == columns {
		return `<div class="crop"><img data-shot src="` + windowPicture(darkFirst) + `" alt="` + e(shows) + `"></div>`
	}
	return `<img class="shot" src="` + shotPicture(of) + `" alt="` + e(shows) + `">`
}
