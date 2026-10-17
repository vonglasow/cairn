package engine

import (
	"errors"
	"fmt"
	"path/filepath"

	"cairn/internal/config"
)

// Home is what a surface knows about where it runs: the home directory, or why there is none, and how to read the environment.
type Home struct {
	Dir     string
	Missing error
	Lookup  func(string) (string, bool)
}

func (h Home) lookup() func(string) (string, bool) {
	if h.Lookup == nil {
		return func(string) (string, bool) { return "", false }
	}
	return h.Lookup
}

// Places is where an operation finds what it reaches: the store, the root every task's outside context sits under, and the home both were found from — kept because one operation, `config show`, answers how they were found rather than where they are.
type Places struct {
	Store  string
	Notes  string
	Plugin string
	Home   Home
}

// Draft is where a window keeps what was typed into the form of a task being added and not saved yet: beside the store it is meant for.
func (p Places) Draft() string { return p.Store + ".draft.json" }

// Locate works out where the store and the notes live: the environment, then the configuration file, then the default place under the home, in that order.
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
	plugin := found.Plugin()
	if plugin == "" {
		plugin = PluginBeside(gate.store.At)
	}
	return Places{Store: gate.store.At, Notes: notes, Plugin: plugin, Home: h}, nil
}

// DoFrom runs an operation for a surface that knows only its home: it refuses what it cannot read first, then finds the places, then runs it there.
func DoFrom(op Operation, h Home, in Input, acting Acting) (Answer, error) {
	if err := op.admitted(in); err != nil {
		return Answer{}, err
	}
	if op.Needs == AConfiguration {
		return doAt(op, Places{Home: h}, in, acting)
	}
	at, err := Locate(h)
	if err != nil {
		return Answer{}, err
	}
	return doAt(op, at, in, acting)
}

// Hold opens the store a surface that stays on screen holds for as long as it is open: found from the home, and refused the way an operation refuses it — a store behind this cairn says what to type about it. What it is asked to do is written down as done through that surface, by whoever acts.
func Hold(h Home, acting Acting) (*Engine, error) {
	at, err := Locate(h)
	if err != nil {
		return nil, err
	}
	held, err := OpenAt(at)
	if err != nil {
		return nil, err
	}
	held.acting = acting
	return held, nil
}
