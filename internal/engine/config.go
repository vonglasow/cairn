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

// Config is how `config show` answers: the file cairn read, and each of its places, alongside what named it.
type Config struct {
	File     string       `json:"file"`
	Store    Emplacement  `json:"store"`
	Notes    Emplacement  `json:"notes"`
	Plugin   Emplacement  `json:"plugin"`
	PagesOff []string     `json:"pages_off,omitempty"`
	Archive  *Archive     `json:"archive,omitempty"`
	Focus    *FocusConfig `json:"focus,omitempty"`
	Time     *TimeConfig  `json:"time,omitempty"`
	Problem  string       `json:"problem,omitempty"`
}

// TimeConfig is what a deleted task does with its time, when the file says it takes it.
type TimeConfig struct {
	Deleted string `json:"deleted"`
}

// FocusConfig is how long a focus session lasts before it asks to be renewed, as the file wrote it.
type FocusConfig struct {
	Session string `json:"session"`
}

// Archive is the day archiving counts from, and how many days each period lasts.
type Archive struct {
	Start string `json:"start"`
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
	g.config.PagesOff = p.PagesOff()
	if start, days, ok := p.Archive(); ok {
		g.config.Archive = &Archive{Start: start, Days: days}
	}
	if p.TimeDeleted() == config.TimeTaken {
		g.config.Time = &TimeConfig{Deleted: config.TimeTaken}
	}
	if session := p.FocusSession(); session != "" {
		g.config.Focus = &FocusConfig{Session: session}
	}
}

func (g *readingGate) NotUnderstood(_ string, err error) {
	g.config.Problem = err.Error()
}

// ConfigShow shows which places cairn will use, and where it got them.
var ConfigShow = Operation{
	Thing:   "config",
	Verb:    "show",
	Short:   "Show which places cairn will use, and where it got them",
	Does:    "Show which places cairn will use — the store, the notes and the Claude Code plugin — the file read, and what named each place.",
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
	wide := max(len(c.Store.At), len(c.Notes.At), len(c.Plugin.At))
	said := fmt.Sprintf(
		"file    %s\n\nstore   %-*s   named by %s\nnotes   %-*s   named by %s\nplugin  %-*s   named by %s\n",
		c.File, wide, c.Store.At, c.Store.By, wide, c.Notes.At, c.Notes.By, wide, c.Plugin.At, c.Plugin.By,
	)
	if len(c.PagesOff) > 0 {
		said += "\npages off   " + strings.Join(c.PagesOff, ", ") + "\n"
	}
	if c.Archive == nil {
		said += "\nnothing is archived on its own\n"
	} else {
		said += fmt.Sprintf("\narchive     every %d days from %s\n", c.Archive.Days, c.Archive.Start)
	}
	if c.Time != nil {
		said += "\ntime        a deleted task takes its time with it\n"
	}
	if c.Focus != nil {
		said += fmt.Sprintf("\nfocus       asks to be renewed after %s\n", c.Focus.Session)
	}
	return said
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
