package engine

// ConfigShow says which places cairn will use, and where it got them from —
// the file it read, and, for each place, what named it: the file, the
// environment, or nothing at all. A person reaches for this when something
// is not where they expected it.

import "cairn/internal/config"

// Emplacement is where a place resolved to, and what named it. Read off
// config.Places through Tell, never rebuilt as a comparison of strings: the
// three provenances a Told is given are the only source of truth for By.
type Emplacement struct {
	At string `json:"at"`
	By string `json:"named_by"`
}

// Config is how `config show` answers: the file cairn read, and each of its
// two places, alongside what named it. Problem is empty when the file made
// sense; when it did not, it carries what stood in the way, and Store and
// Notes are left at their zero value — an emplacement resolved alongside a
// file cairn could not make sense of is not one worth showing, since every
// other command refuses this same file.
type Config struct {
	File    string      `json:"file"`
	Store   Emplacement `json:"store"`
	Notes   Emplacement `json:"notes"`
	Problem string      `json:"problem,omitempty"`
}

// placesGate is a config.Told that captures each place's emplacement and
// what named it, so ConfigShow's Run can build a Config from what Places
// tells it rather than from a comparison of strings.
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

// readingGate is a config.ReadingTold that turns what a Reading told into a
// Config: the file it read from — told regardless of what came of it — and
// either its two emplacements when it made sense of the file, or what stood
// in the way when it did not.
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
}

func (g *readingGate) NotUnderstood(_ string, err error) {
	g.config.Problem = err.Error()
}

// ConfigShow shows which places cairn will use, and where it got them.
var ConfigShow = Operation{
	Thing:   "config",
	Verb:    "show",
	Short:   "Show which places cairn will use, and where it got them",
	Does:    "Show which places cairn will use — the store and the notes — the file read, and what named each place.",
	Level:   Read,
	Needs:   AConfiguration,
	Answers: Config{},
	Run: func(on On, _ Input) (Answer, error) {
		gate := &readingGate{}
		on.Reading.Tell(gate)
		return Answer{Config: &gate.config}, nil
	},
}
