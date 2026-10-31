package engine

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
	File    string        `json:"file"`
	Store   Emplacement   `json:"store"`
	Notes   Emplacement   `json:"notes"`
	Plugin  Emplacement   `json:"plugin"`
	Pages   *PagesConfig  `json:"pages,omitempty"`
	Archive *Archive      `json:"archive,omitempty"`
	Focus   *FocusConfig  `json:"focus,omitempty"`
	Time    *TimeConfig   `json:"time,omitempty"`
	Window  *WindowConfig `json:"window,omitempty"`
	Problem string        `json:"problem,omitempty"`
}

// PagesConfig is whether each page is on.
type PagesConfig struct {
	Tasks      bool `json:"tasks"`
	Objectives bool `json:"objectives"`
	Time       bool `json:"time"`
}

// TimeConfig is what a deleted task does with its time, kept or taken, and whether a week is drawn with its weekend, shown or hidden.
type TimeConfig struct {
	Deleted  string `json:"deleted"`
	Weekends string `json:"weekends"`
}

// WindowConfig is whether the window opens with the sizes it was left with.
type WindowConfig struct {
	KeepSizes bool `json:"keep_sizes"`
}

// FocusConfig is how long a focus session lasts before it asks to be renewed, as the file wrote it; nothing when it is not said.
type FocusConfig struct {
	Session string `json:"session,omitempty"`
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
	config Config
}

func (g *readingGate) ReadFrom(path string) {
	g.config.File = path
}

func (g *readingGate) Resolved(p config.Places) {
	places := &placesGate{}
	p.Tell(places)
	g.config.Store = places.store
	g.config.Notes = places.notes
	g.config.Plugin = Emplacement{At: PluginBeside(places.store.At), By: "nothing"}
	if told := p.Plugin(); told != "" {
		g.config.Plugin = Emplacement{At: told, By: "the file"}
	}
	pages := p.Pages()
	g.config.Pages = &PagesConfig{Tasks: pages[0], Objectives: pages[1], Time: pages[2]}
	start, days, _ := p.Archive()
	g.config.Archive = &Archive{Start: start, Days: days}
	g.config.Time = &TimeConfig{Deleted: p.TimeDeleted(), Weekends: p.TimeWeekends()}
	g.config.Focus = &FocusConfig{Session: p.FocusSession()}
	g.config.Window = &WindowConfig{KeepSizes: p.KeepSizes()}
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
		gate := &readingGate{}
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
	fmt.Fprintf(&said, "# the configuration cairn uses, every key with the value in force\n# the file it reads: %s\n\n", c.File)
	for _, p := range places {
		fmt.Fprintf(&said, "%-*s   # named by %s\n", wide, p[0], p[1])
	}
	if c.Pages != nil {
		fmt.Fprintf(&said, "\n[pages]\ntasks = %t\nobjectives = %t\ntime = %t\n", c.Pages.Tasks, c.Pages.Objectives, c.Pages.Time)
	}
	if c.Archive != nil {
		said.WriteString("\n[archive]\n")
		if c.Archive.Start == "" {
			said.WriteString("# start is not said: nothing is archived on its own\n")
		} else {
			fmt.Fprintf(&said, "start = %s\n", c.Archive.Start)
		}
		fmt.Fprintf(&said, "days = %d\n", c.Archive.Days)
	}
	if c.Focus != nil {
		said.WriteString("\n[focus]\n")
		if c.Focus.Session == "" {
			said.WriteString("# session is not said: a focus session is never asked to be renewed\n")
		} else {
			fmt.Fprintf(&said, "session = %q\n", c.Focus.Session)
		}
	}
	if c.Time != nil {
		fmt.Fprintf(&said, "\n[time]\ndeleted = %q\nweekends = %q\n", c.Time.Deleted, c.Time.Weekends)
	}
	if c.Window != nil {
		fmt.Fprintf(&said, "\n[window]\nkeep_sizes = %t\n", c.Window.KeepSizes)
	}
	return said.String()
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
	Does:    "Write the configuration file where cairn reads it, every key commented and set to what cairn uses when nothing is said, to change from there. A configuration already there is never written over.",
	Example: "cairn config init",
	Level:   Write,
	Needs:   AConfiguration,
	Again:   ChangesNothingMore,
	Answers: ConfigWritten{},
	Run: func(on On, _ Input) (Answer, error) {
		if on.Home.Dir == "" {
			return Answer{}, fmt.Errorf("%w: %w", ErrStore, on.Home.Missing)
		}
		path := config.FileIn(on.Home.Dir)
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
		return Answer{Written: &path}, nil
	},
}

func configWrittenText(path string) string {
	return "the configuration is written in " + path + "\n" +
		"every key in it says what it is, set to what cairn uses when nothing is said: change what you want\n"
}
