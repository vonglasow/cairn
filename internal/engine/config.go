package engine

import (
	"fmt"
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
	File     string      `json:"file"`
	Store    Emplacement `json:"store"`
	Notes    Emplacement `json:"notes"`
	Plugin   Emplacement `json:"plugin"`
	PagesOff []string    `json:"pages_off,omitempty"`
	Problem  string      `json:"problem,omitempty"`
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
	return said
}
