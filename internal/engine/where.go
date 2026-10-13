package engine

// Where an operation finds what it reaches, worked out once for every surface.
//
// A surface knows two things about where it runs — the home directory, and the
// environment — and nothing else about places. Turning those into a store and
// a root for the notes is a rule, so it lives here: it used to be the command
// line's, and the server had to be handed what the command line had found,
// which left `config show` answering an agent with nothing at all.

import (
	"errors"
	"fmt"
	"path/filepath"

	"cairn/internal/config"
)

// Home is what a surface knows about where it runs: the home directory, or why
// there is none, and how to read the environment.
type Home struct {
	// Dir is the home directory; empty when there is none to be found.
	Dir string
	// Missing says why there is no home, when Dir is empty.
	Missing error
	// Lookup reads the environment, the way os.LookupEnv does. Nil reads an
	// environment that says nothing.
	Lookup func(string) (string, bool)
}

// lookup reads the environment Home was handed, or one that says nothing.
func (h Home) lookup() func(string) (string, bool) {
	if h.Lookup == nil {
		return func(string) (string, bool) { return "", false }
	}
	return h.Lookup
}

// Places is where an operation finds what it reaches: the store, the root
// every task's outside context sits under, and the home both were found from —
// kept because one operation, `config show`, answers how they were found
// rather than where they are.
type Places struct {
	Store string
	Notes string
	Home  Home
}

// Locate works out where the store and the notes live: the environment, then
// the configuration file, then the default place under the home, in that order.
//
// The notes are only ever named, never opened, so finding them refuses
// nothing. With no home to put the default under, and nothing naming them,
// they are left unsaid: a place relative to wherever the command happened to
// run would send whoever reads it to look somewhere nobody writes.
//
// The home is only required when the environment does not already name the
// store: CAIRN_STORE set with no home to find — a cron entry, a container —
// still has to work.
//
// Only what the person wrote wrong is theirs to fix and passes through as it
// is; everything else about reading the file — permissions, a missing
// directory — is folded into ErrStore, the answer an unusable store gives
// everywhere else.
func Locate(h Home) (Places, error) {
	if h.Dir == "" {
		if said, ok := h.lookup()(config.StoreVariable); !ok || said == "" {
			return Places{}, fmt.Errorf("%w: %w", ErrStore, h.Missing)
		}
	}
	found, err := config.Read(h.Dir, h.lookup())
	if err != nil {
		for _, theirs := range []error{config.ErrNotTOML, config.ErrUnknownKey, config.ErrNotAbsolute} {
			if errors.Is(err, theirs) {
				return Places{}, err
			}
		}
		return Places{}, fmt.Errorf("%w: %w", ErrStore, err)
	}
	gate := &placesGate{}
	found.Tell(gate)
	notes := gate.notes.At
	if !filepath.IsAbs(notes) {
		notes = ""
	}
	return Places{Store: gate.store.At, Notes: notes, Home: h}, nil
}

// DoFrom runs an operation for a surface that knows only its home: it refuses
// what it cannot read first, then finds the places, then runs it there.
//
// `config show` is the one operation that does not need the places found: it
// answers how they would be found, and it has to stay reachable exactly when a
// file every other command refuses stands in the way.
func DoFrom(op Operation, h Home, in Input, through string) (Answer, error) {
	if err := op.admitted(in); err != nil {
		return Answer{}, err
	}
	if op.Needs == AConfiguration {
		return doAt(op, Places{Home: h}, in, through)
	}
	at, err := Locate(h)
	if err != nil {
		return Answer{}, err
	}
	return doAt(op, at, in, through)
}

// Hold opens the store a surface that stays on screen holds for as long as it
// is open: found from the home, and refused the way an operation refuses it —
// a store behind this cairn says what to type about it.
func Hold(h Home) (*Engine, error) {
	at, err := Locate(h)
	if err != nil {
		return nil, err
	}
	return OpenAt(at)
}
