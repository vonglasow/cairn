package workspace

import (
	"cairn/internal/store/files"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Emplacement is where a place resolved to, and what named it.
type Emplacement struct {
	At string `json:"at"`
	By string `json:"named_by"`
}

// Config is how `config show` answers: the file cairn read, each of its places alongside what named it, and every other key with the value cairn uses — what the file says, or what is used when it says nothing.
type Config struct {
	File        string            `json:"file"`
	Store       Emplacement       `json:"store"`
	Notes       Emplacement       `json:"notes"`
	Plugin      Emplacement       `json:"plugin"`
	Pages       *PagesConfig      `json:"pages,omitempty"`
	Archive     *Archive          `json:"archive,omitempty"`
	Focus       *FocusConfig      `json:"focus,omitempty"`
	Approach    *ApproachConfig   `json:"approach,omitempty"`
	Time        *TimeConfig       `json:"time,omitempty"`
	Window      *WindowConfig     `json:"window,omitempty"`
	Recurrence  *RecurrenceConfig `json:"series,omitempty"`
	Workspace   *Colouring        `json:"workspace,omitempty" jsonschema:"what the store's settings say of its workspace; nothing for a store no workspace names and whose settings say no colour"`
	Workspaces  []Workspace       `json:"workspaces,omitempty"`
	Themes      []ThemeSeen       `json:"themes,omitempty" jsonschema:"every theme cairn sees, the ones it carries and one's own, sorted by name"`
	Problem     string            `json:"problem,omitempty"`
	Settings    string            `json:"settings,omitempty" jsonschema:"the settings beside the store read over the file, when there are any"`
	Written     []string          `json:"written,omitempty"`
	WrittenHere []string          `json:"written_here,omitempty" jsonschema:"every key the settings beside the store write, dotted under its table: those win over the file"`
	Global      map[string]string `json:"global,omitempty" jsonschema:"for each key the settings beside the store write, the value the global configuration, or its default, would give"`
	Ignored     []string          `json:"ignored,omitempty" jsonschema:"every key the file holds that this cairn does not know and passed over, one sentence each"`
}

// Colouring is the colour the workspace wears: the one its settings say, or one read off its name.
type Colouring struct {
	Colour string `json:"colour"`
	Name   string `json:"name,omitempty" jsonschema:"the name the store's settings give it, shown where no workspace names it"`
}

// NotSaid is how a key with no value, and none by default, is told.
const NotSaid = "not said"

var quotedKeys = []string{"pages.time", "focus.session", "approach.label", "time.deleted", "time.weekends", "window.theme", "window.intro"}

// ValuesOf is each key config show lists, and the value said for it.
func ValuesOf(p files.Places) map[string]string {
	pages := p.Pages()
	start, days, _ := p.Archive()
	unless := func(said string) string {
		if said == "" {
			return NotSaid
		}
		return said
	}
	return map[string]string{
		"pages.tasks": strconv.FormatBool(pages[0]), "pages.objectives": strconv.FormatBool(pages[1]), "pages.time": p.TimeMode(),
		"archive.start": unless(start), "archive.days": strconv.Itoa(days),
		"focus.session": unless(p.FocusSession()), "approach.label": unless(p.ApproachLabel()),
		"time.deleted": p.TimeDeleted(), "time.weekends": p.TimeWeekends(), "window.keep_sizes": strconv.FormatBool(p.KeepSizes()),
		"window.theme": p.Theme(), "window.intro": p.Intro(), "window.ghost": strconv.Itoa(p.Ghost()), "series.on": strconv.FormatBool(p.Recurrence()),
	}
}

// PagesConfig is whether each page is on, and what the time is: on, hidden or off.
type PagesConfig struct {
	Tasks      bool   `json:"tasks"`
	Objectives bool   `json:"objectives"`
	Time       string `json:"time" jsonschema:"on, recorded and shown; hidden, recorded and not shown; off, neither"`
}

// TimeConfig is what a deleted task does with its time, kept or taken, and whether a week is drawn with its weekend, shown or hidden.
type TimeConfig struct {
	Deleted  string `json:"deleted"`
	Weekends string `json:"weekends"`
}

// WindowConfig is whether the window opens with the sizes it was left with, the scheme it is named to wear, and why that scheme cannot be worn when it cannot.
type WindowConfig struct {
	KeepSizes bool   `json:"keep_sizes"`
	Theme     string `json:"theme"`
	Intro     string `json:"intro" jsonschema:"how the board arrives when the window opens"`
	Ghost     int    `json:"ghost" jsonschema:"how much of what is dragged shows, in percent"`
	Unworn    string `json:"unworn,omitempty" jsonschema:"why the scheme named cannot be worn; the window then wears default"`
}

// RecurrenceConfig is whether a task may come back as a series.
type RecurrenceConfig struct {
	On bool `json:"on"`
}

// FocusConfig is how long a focus session lasts before it asks to be renewed, as the file wrote it; nothing when it is not said.
type FocusConfig struct {
	Session string `json:"session,omitempty"`
}

// ApproachConfig is the label asking for the approach to be agreed before any code, as the file wrote it; nothing when it is not said.
type ApproachConfig struct {
	Label string `json:"label,omitempty" jsonschema:"the label asking for the approach to be agreed before any code; empty when none is"`
}

// Archive is the day archiving counts from, nothing when it is not said and nothing is archived on its own, and how many days each period lasts.
type Archive struct {
	Start string `json:"start,omitempty"`
	Days  int    `json:"days"`
}

// PlacesGate is told where the store and the notes are, and what named each.
type PlacesGate struct {
	Store Emplacement
	Notes Emplacement
}

// StoreNamedByTheFile is told the store the configuration names.
func (g *PlacesGate) StoreNamedByTheFile(at string) { g.Store = Emplacement{At: at, By: "the file"} }

// StoreNamedByTheEnvironment is told the store CAIRN_STORE names.
func (g *PlacesGate) StoreNamedByTheEnvironment(at string) {
	g.Store = Emplacement{At: at, By: "the environment"}
}

// StoreNamedByNothing is told the store cairn takes when nothing names one.
func (g *PlacesGate) StoreNamedByNothing(at string) { g.Store = Emplacement{At: at, By: "nothing"} }

// NotesNamedByTheFile is told the notes the configuration names.
func (g *PlacesGate) NotesNamedByTheFile(at string) { g.Notes = Emplacement{At: at, By: "the file"} }

// NotesNamedByTheEnvironment is told the notes CAIRN_NOTES names.
func (g *PlacesGate) NotesNamedByTheEnvironment(at string) {
	g.Notes = Emplacement{At: at, By: "the environment"}
}

// NotesNamedByNothing is told the notes cairn takes when nothing names them.
func (g *PlacesGate) NotesNamedByNothing(at string) { g.Notes = Emplacement{At: at, By: "nothing"} }

// ConfigText is config show for a person.
func ConfigText(c Config) string {
	if c.Problem != "" {
		return fmt.Sprintf(
			"file    %s\n\nthe file cannot be made sense of: %s\nthe other commands will refuse until it is\n",
			c.File, c.Problem,
		)
	}
	places := [][2]string{
		{fmt.Sprintf("# store = %q", c.Store.At), c.Store.By},
		{fmt.Sprintf("# notes = %q", c.Notes.At), c.Notes.By},
		{fmt.Sprintf("plugin = %q", c.Plugin.At), c.Plugin.By},
	}
	wide := 0
	for _, p := range places {
		wide = max(wide, len(p[0]))
	}
	var said strings.Builder
	fmt.Fprintf(&said, "# the configuration cairn uses, every key with the value in force\n# the file it reads: %s\n", c.File)
	if c.Settings != "" {
		fmt.Fprintf(&said, "# the store's settings: %s\n", c.Settings)
	}
	said.WriteString("\n")
	for _, p := range places {
		named := "# named by "
		if strings.HasPrefix(p[0], "#") {
			named = "  named by "
		}
		fmt.Fprintf(&said, "%-*s   %s%s\n", wide, p[0], named, p[1])
	}
	for _, ignored := range c.Ignored {
		fmt.Fprintf(&said, "# ignored: %s\n", ignored)
	}
	if c.Pages != nil {
		fmt.Fprintf(&said, "\n[pages]\ntasks = %t%s\nobjectives = %t%s\ntime = %q%s\n",
			c.Pages.Tasks, c.from("pages.tasks"), c.Pages.Objectives, c.from("pages.objectives"), c.Pages.Time, c.from("pages.time"))
	}
	if c.Archive != nil {
		said.WriteString("\n[archive]\n")
		if c.Archive.Start == "" {
			said.WriteString("# start is not said: nothing is archived on its own\n")
		} else {
			fmt.Fprintf(&said, "start = %s%s\n", c.Archive.Start, c.from("archive.start"))
		}
		fmt.Fprintf(&said, "days = %d%s\n", c.Archive.Days, c.from("archive.days"))
	}
	if c.Focus != nil {
		said.WriteString("\n[focus]\n")
		if c.Focus.Session == "" {
			said.WriteString("# session is not said: a focus session is never asked to be renewed\n")
		} else {
			fmt.Fprintf(&said, "session = %q%s\n", c.Focus.Session, c.from("focus.session"))
		}
	}
	if c.Approach != nil {
		said.WriteString("\n[approach]\n")
		if c.Approach.Label != "" {
			fmt.Fprintf(&said, "label = %q%s\n", c.Approach.Label, c.from("approach.label"))
		} else if slices.Contains(c.WrittenHere, "approach.label") {
			said.WriteString("# label is said empty in the store's settings: no label asks for the approach to be agreed first" + c.global("approach.label") + "\n")
		} else {
			said.WriteString("# label is not said: no label asks for the approach to be agreed first\n")
		}
	}
	if c.Time != nil {
		fmt.Fprintf(&said, "\n[time]\ndeleted = %q%s\nweekends = %q%s\n", c.Time.Deleted, c.from("time.deleted"), c.Time.Weekends, c.from("time.weekends"))
	}
	if c.Window != nil {
		fmt.Fprintf(&said, "\n[window]\nkeep_sizes = %t%s\ntheme = %q%s\n", c.Window.KeepSizes, c.from("window.keep_sizes"), c.Window.Theme, c.from("window.theme"))
		said.WriteString(themesText(c))
		fmt.Fprintf(&said, "intro = %q%s\n", c.Window.Intro, c.from("window.intro"))
		fmt.Fprintf(&said, "ghost = %d%s\n", c.Window.Ghost, c.from("window.ghost"))
	}
	if c.Recurrence != nil {
		fmt.Fprintf(&said, "\n[series]\non = %t%s\n", c.Recurrence.On, c.from("series.on"))
	}
	if c.Workspace != nil {
		from := "read off the workspace's name"
		if slices.Contains(c.WrittenHere, "workspace.colour") {
			from = "set in the store's settings"
		}
		fmt.Fprintf(&said, "\n[workspace]\ncolour = %q   # %s\n", c.Workspace.Colour, from)
		if c.Workspace.Name != "" {
			fmt.Fprintf(&said, "name = %q   # set in the store's settings\n", c.Workspace.Name)
		}
	}
	said.WriteString(workspacesTable(c.Workspaces))
	return said.String()
}

func themesText(c Config) string {
	var said strings.Builder
	if c.Window.Unworn != "" {
		fmt.Fprintf(&said, "# theme: %s\n", c.Window.Unworn)
	}
	if len(c.Themes) == 0 {
		return said.String()
	}
	said.WriteString("# the schemes cairn sees, any of them a theme:\n")
	for _, seen := range c.Themes {
		fmt.Fprintf(&said, "#   %-18s %s", seen.Slug, seen.From)
		if seen.InForce {
			said.WriteString(", in force")
		}
		if seen.Problem != "" {
			said.WriteString(", cannot be worn: " + seen.Problem)
		}
		said.WriteString("\n")
	}
	for _, seen := range c.Themes {
		for _, short := range seen.Short {
			fmt.Fprintf(&said, "# %s falls short: %s\n", seen.Slug, short)
		}
	}
	return said.String()
}

func (c Config) from(key string) string {
	if c.Settings == "" {
		return ""
	}
	return "   # " + saidWhere(c.Written, c.WrittenHere, key) + c.global(key)
}

func (c Config) global(key string) string {
	value, ok := c.Global[key]
	if !ok {
		return ""
	}
	if value != NotSaid && slices.Contains(quotedKeys, key) {
		value = strconv.Quote(value)
	}
	return "; global: " + value
}

// WhereSaid is where a key is said: the global configuration, or the settings beside a store.
func WhereSaid(p files.Places, keys ...string) string {
	return saidWhere(p.Written(), p.WrittenHere(), keys...)
}

func saidWhere(written, here []string, keys ...string) string {
	among := func(said []string) bool {
		return slices.ContainsFunc(keys, func(key string) bool { return slices.Contains(said, key) })
	}
	switch {
	case among(here):
		return "set in the store's settings"
	case among(written):
		return "from the global configuration"
	}
	return "by default"
}

// ErrConfigAlreadyThere is returned when config init finds a configuration already written: it is never written over.
var ErrConfigAlreadyThere = errors.New("a configuration is already there; change it, or move it away first")

// ConfigWritten says where config init wrote the configuration.
type ConfigWritten struct {
	Written string `json:"written"`
}

// ConfigWrittenText is config init for a person.
func ConfigWrittenText(path string) string {
	return "the configuration is written in " + path + "\n" +
		"every key in it says what it is, set to what cairn uses when nothing is said: change what you want\n"
}
