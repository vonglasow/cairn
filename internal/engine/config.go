package engine

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/config"
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
	Recurrence  *RecurrenceConfig `json:"recurrence,omitempty"`
	Workspace   *WorkspaceConfig  `json:"workspace,omitempty" jsonschema:"what the store's settings say of its workspace; nothing for a store no workspace names and whose settings say no colour"`
	Workspaces  []Workspace       `json:"workspaces,omitempty"`
	Themes      []ThemeSeen       `json:"themes,omitempty" jsonschema:"every colour scheme cairn sees, the ones it carries and one's own, sorted by name"`
	Problem     string            `json:"problem,omitempty"`
	Settings    string            `json:"settings,omitempty" jsonschema:"the settings beside the store read over the file, when there are any"`
	Written     []string          `json:"written,omitempty"`
	WrittenHere []string          `json:"written_here,omitempty" jsonschema:"every key the settings beside the store write, dotted under its table: those win over the file"`
	Global      map[string]string `json:"global,omitempty" jsonschema:"for each key the settings beside the store write, the value the global configuration, or its default, would give"`
	Ignored     []string          `json:"ignored,omitempty" jsonschema:"every key the file holds that this cairn does not know and passed over, one sentence each"`
}

// WorkspaceConfig is the colour the workspace wears: the one its settings say, or one read off its name.
type WorkspaceConfig struct {
	Colour string `json:"colour"`
}

// NotSaid is how a key with no value, and none by default, is told.
const NotSaid = "not said"

var quotedKeys = []string{"pages.time", "focus.session", "approach.label", "time.deleted", "time.weekends", "window.theme", "window.intro"}

func valuesOf(p config.Places) map[string]string {
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
		"window.theme": p.Theme(), "window.intro": p.Intro(), "window.ghost": strconv.Itoa(p.Ghost()), "recurrence.on": strconv.FormatBool(p.Recurrence()),
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

type placesGate struct {
	store Emplacement
	notes Emplacement
}

func (g *placesGate) StoreNamedByTheFile(at string) { g.store = Emplacement{At: at, By: "the file"} }
func (g *placesGate) StoreNamedByTheEnvironment(at string) {
	g.store = Emplacement{At: at, By: "the environment"}
}
func (g *placesGate) StoreNamedByNothing(at string) { g.store = Emplacement{At: at, By: "nothing"} }
func (g *placesGate) NotesNamedByTheFile(at string) { g.notes = Emplacement{At: at, By: "the file"} }
func (g *placesGate) NotesNamedByTheEnvironment(at string) {
	g.notes = Emplacement{At: at, By: "the environment"}
}
func (g *placesGate) NotesNamedByNothing(at string) { g.notes = Emplacement{At: at, By: "nothing"} }

type readingGate struct {
	home   Home
	config Config
}

func (g *readingGate) ReadFrom(path string) {
	g.config.File = path
}

func (g *readingGate) Resolved(p config.Places) {
	places, err := g.home.answering(p)
	if err != nil {
		g.config.Problem = err.Error()
		return
	}
	global := valuesOf(p)
	p, err = p.WithSettings(config.SettingsBeside(places.store.At))
	if err != nil {
		g.config.Problem = err.Error()
		return
	}
	for _, key := range p.WrittenHere() {
		if value, ok := global[key]; ok {
			if g.config.Global == nil {
				g.config.Global = map[string]string{}
			}
			g.config.Global[key] = value
		}
	}
	if colour := colourOf(places.workspace, places.store.At); colour != "" {
		g.config.Workspace = &WorkspaceConfig{Colour: colour}
	}
	g.config.Store = places.store
	g.config.Notes = places.notes
	g.config.Workspaces = workspacesIn(p)
	g.config.Plugin = Emplacement{At: PluginBeside(places.store.At), By: "nothing"}
	if told := p.Plugin(); told != "" {
		g.config.Plugin = Emplacement{At: told, By: "the file"}
	}
	pages := p.Pages()
	g.config.Pages = &PagesConfig{Tasks: pages[0], Objectives: pages[1], Time: p.TimeMode()}
	start, days, _ := p.Archive()
	g.config.Archive = &Archive{Start: start, Days: days}
	g.config.Time = &TimeConfig{Deleted: p.TimeDeleted(), Weekends: p.TimeWeekends()}
	g.config.Focus = &FocusConfig{Session: p.FocusSession()}
	g.config.Approach = &ApproachConfig{Label: p.ApproachLabel()}
	g.config.Recurrence = &RecurrenceConfig{On: p.Recurrence()}
	g.themes(p)
	g.config.Written = p.Written()
	g.config.Settings = p.Settings()
	g.config.WrittenHere = p.WrittenHere()
	g.config.Ignored = p.Ignored()
}

func (g *readingGate) NotUnderstood(_ string, err error) {
	g.config.Problem = err.Error()
}

// ConfigShow shows which places cairn will use, and where it got them.
var ConfigShow = Operation{
	Thing:   "config",
	Verb:    "show",
	Short:   "Show the whole configuration cairn uses, and where it got its places",
	Does:    "Show the whole configuration cairn uses, written as the file is: every key under its table with the value in force — what the file says, or what is used when it says nothing — the file read, and what named each place: the file, the environment, or nothing.",
	Example: "cairn config show",
	Level:   Read,
	Needs:   AConfiguration,
	Answers: Config{},
	Run: func(on On, _ Input) (Answer, error) {
		reading, err := config.NewReading(on.Home.Dir, on.Home.lookup())
		if err != nil {
			return Answer{}, err
		}
		gate := &readingGate{home: on.Home}
		reading.Tell(gate)
		return Answer{Config: &gate.config}, nil
	},
}

func configText(c Config) string {
	if c.Problem != "" {
		return fmt.Sprintf(
			"file    %s\n\nthe file cannot be made sense of: %s\nthe other commands will refuse until it is\n",
			c.File, c.Problem,
		)
	}
	places := [][2]string{
		{fmt.Sprintf("store = %q", c.Store.At), c.Store.By},
		{fmt.Sprintf("notes = %q", c.Notes.At), c.Notes.By},
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
		fmt.Fprintf(&said, "%-*s   # named by %s\n", wide, p[0], p[1])
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
		switch {
		case c.Approach.Label != "":
			fmt.Fprintf(&said, "label = %q%s\n", c.Approach.Label, c.from("approach.label"))
		case slices.Contains(c.WrittenHere, "approach.label"):
			said.WriteString("# label is said empty in the store's settings: no label asks for the approach to be agreed first" + c.global("approach.label") + "\n")
		default:
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
		fmt.Fprintf(&said, "\n[recurrence]\non = %t%s\n", c.Recurrence.On, c.from("recurrence.on"))
	}
	if c.Workspace != nil {
		from := "read off the workspace's name"
		if slices.Contains(c.WrittenHere, "workspace.colour") {
			from = "set in the store's settings"
		}
		fmt.Fprintf(&said, "\n[workspace]\ncolour = %q   # %s\n", c.Workspace.Colour, from)
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

func whereSaid(p config.Places, keys ...string) string {
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

// ConfigInit writes the whole configuration file, every key set to what is used when nothing is said.
var ConfigInit = Operation{
	Thing:   "config",
	Verb:    "init",
	Short:   "Write the configuration file, every key in it",
	Does:    "Write the configuration file where cairn reads it, every key commented and set to what cairn uses when nothing is said, to change from there, with one workspace, main, on ~/.cairn, the default, its folder made. A configuration already there is never written over.",
	Example: "cairn config init",
	Level:   Write,
	Needs:   AConfiguration,
	Again:   ChangesNothingMore,
	Answers: ConfigWritten{},
	Run: func(on On, _ Input) (Answer, error) {
		path, err := on.Home.file()
		if err != nil {
			return Answer{}, err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return Answer{}, err
		}
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if errors.Is(err, fs.ErrExist) {
			return Answer{}, fmt.Errorf("%s: %w", path, ErrConfigAlreadyThere)
		}
		if err != nil {
			return Answer{}, err
		}
		if _, err := file.WriteString(config.Example); err != nil {
			_ = file.Close()
			return Answer{}, err
		}
		if err := file.Close(); err != nil {
			return Answer{}, err
		}
		read, err := on.Home.configuration()
		if err != nil {
			return Answer{}, err
		}
		if one, err := workspaceNamed(read, read.DefaultWorkspace()); err == nil {
			if err := workspaceFolder(one.Folder); err != nil {
				return Answer{}, err
			}
		}
		return Answer{Written: &path}, nil
	},
}

func configWrittenText(path string) string {
	return "the configuration is written in " + path + "\n" +
		"every key in it says what it is, set to what cairn uses when nothing is said: change what you want\n"
}
