// Package config reads the file that says cairn's two places. It knows only
// where that file sits and what it says, never what a place means once read —
// expansion, defaults and the environment are somebody else's turn.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// DefaultFolder and defaultFile make up the place the store goes when
// nothing names it: no file at all, or a file that says nothing under
// `store`. Owned here, not borrowed from internal/engine, so the one place
// that decides the fallback also spells it. DefaultFolder is exported
// because a caller outside this package needs to name it; defaultFile has
// no such caller and stays private.
const (
	DefaultFolder = ".cairn"
	defaultFile   = "cairn.db"
)

// defaultNotesFolder is where notes go under DefaultFolder when nothing
// names them — the twin of defaultFile, for the other place this package
// resolves. No caller outside this package needs to name it, so it stays
// private like defaultFile does.
const defaultNotesFolder = "notes"

// StoreVariable is the environment variable that replaces whatever the file
// or the default answered, named here because this package now owns it —
// the surface that reads it lives elsewhere, but the name is spelled once.
const StoreVariable = "CAIRN_STORE"

// NotesVariable is the environment variable that replaces whatever the file
// or the default answered for notes — the twin of StoreVariable, for the
// other place this package resolves.
const NotesVariable = "CAIRN_NOTES"

// ErrUnknownKey is returned when the file names a key cairn does not know —
// the decoding is strict, so a typo or a leftover from another tool is
// refused rather than quietly ignored.
var ErrUnknownKey = errors.New("not a key cairn knows")

// ErrNotTOML is returned when the file's content cannot be made sense of as
// TOML at all — not a specific rule the file broke, just the syntax itself.
var ErrNotTOML = errors.New("not a file cairn can make sense of")

// ErrNotAbsolute is returned when a place the file names is not absolute,
// even once a leading "~" has been expanded — a configuration is read from
// everywhere, so what it names has to mean one thing regardless of where a
// command runs from.
var ErrNotAbsolute = errors.New("must be absolute")

// namedBy is which of the three provenances resolved an emplacement. Private:
// nothing outside this package names one directly, it is only ever told
// through Tell.
type namedBy int

const (
	byNothing namedBy = iota
	byTheFile
	byTheEnvironment
)

// Places is what the configuration file says, exactly as it was written. It
// tells what it carries; nothing reaches into it.
type Places struct {
	store        string
	notes        string
	storeNamedBy namedBy
	notesNamedBy namedBy
}

// file is the shape config.toml decodes into. Its tags are the only list of
// what cairn accepts — acceptedKeys reads them back rather than a second,
// hand-kept list drifting from this one.
//
// Store and Notes are pointers so a key that is absent — nil — is not
// confused with one written as the empty string: filepath.IsAbs("") is
// false, so an explicit empty value is refused rather than silently taking
// the default meant for "nothing said".
type file struct {
	Store *string `toml:"store"`
	Notes *string `toml:"notes"`
}

// acceptedKeys names, in the order the struct declares them, the keys cairn
// decodes config.toml into — derived from the tags above so a key added
// there needs no matching edit to a message spelled out here.
func acceptedKeys() []string {
	t := reflect.TypeFor[file]()
	keys := make([]string, 0, t.NumField())
	for i := range t.NumField() {
		keys = append(keys, t.Field(i).Tag.Get("toml"))
	}
	return keys
}

// Reading is the tolerant primitive Read is a gate over: it carries the path
// of the file it read, and either the emplacements the file named or the
// problem that stood in the way — never both, since emplacements resolved
// alongside a rule the file broke are not emplacements worth trusting. The
// path is carried regardless: "which file" is half of what a problem has to
// answer.
type Reading struct {
	path   string
	places Places
	err    error
}

// ReadingTold is the membrane of a Reading: whoever wants to know what a
// reading found implements this and is told it — its emplacements, or the
// problem it hit naming the file it read. Never both.
type ReadingTold interface {
	ReadFrom(path string)
	Resolved(Places)
	NotUnderstood(path string, err error)
}

// Tell says which file this reading read from, told first and regardless of
// what came of it, then what it found: its emplacements if it made sense of
// the file, or — naming the path again — the problem that stood in the way.
// "Which file" is answered the same way whether the reading made sense of it
// or not, since a reading that fell back to defaults still read from
// somewhere.
func (r Reading) Tell(to ReadingTold) {
	to.ReadFrom(r.path)
	if r.err != nil {
		to.NotUnderstood(r.path, r.err)
		return
	}
	to.Resolved(r.places)
}

// NewReading decodes <home>/.config/cairn/config.toml, tolerating what it
// cannot make sense of by reporting it through the Reading it hands back
// rather than returning it. It still returns an error outright when the file
// could not be read at all beyond not existing — that case is frozen out of
// this tolerance, and surfaces exactly as it always has.
func NewReading(home string, lookupEnv func(string) (string, bool)) (Reading, error) {
	path := filepath.Join(home, ".config", "cairn", "config.toml")
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Reading{}, err
	}

	var f file
	if err == nil {
		if err := toml.NewDecoder(bytes.NewReader(data)).DisallowUnknownFields().Decode(&f); err != nil {
			// *StrictMissingError carries the names of the keys the file used
			// that the struct above does not know — checked before the more
			// general *toml.DecodeError below, since StrictMissingError also
			// unwraps to one and would otherwise be caught there first,
			// answering with a message that never names the key.
			var strict *toml.StrictMissingError
			if errors.As(err, &strict) {
				return Reading{path: path, err: fmt.Errorf("%q: %w — cairn accepts %s",
					strings.Join(strict.Errors[0].Key(), "."), ErrUnknownKey, strings.Join(acceptedKeys(), ", "))}, nil
			}
			// *toml.DecodeError is what is left once StrictMissingError has
			// already been ruled out above: the content itself could not be
			// made sense of as TOML, rather than a rule cairn's own strict
			// decoding added on top of it.
			var decode *toml.DecodeError
			if errors.As(err, &decode) {
				return Reading{path: path, err: fmt.Errorf("%q: %w", path, ErrNotTOML)}, nil
			}
			return Reading{}, err
		}
	}

	// answer is what the file said, or the default when it said nothing —
	// worked out completely, before the environment gets a say. Named by
	// nothing until one of the branches below says otherwise.
	answer := Places{
		store:        filepath.Join(home, DefaultFolder, defaultFile),
		notes:        filepath.Join(home, DefaultFolder, defaultNotesFolder),
		storeNamedBy: byNothing,
		notesNamedBy: byNothing,
	}
	if f.Store != nil {
		expanded := expandTilde(*f.Store, home)
		if !filepath.IsAbs(expanded) {
			return Reading{path: path, err: fmt.Errorf("%q: %w", *f.Store, ErrNotAbsolute)}, nil
		}
		answer.store = expanded
		answer.storeNamedBy = byTheFile
	}
	if f.Notes != nil {
		expanded := expandTilde(*f.Notes, home)
		if !filepath.IsAbs(expanded) {
			return Reading{path: path, err: fmt.Errorf("%q: %w", *f.Notes, ErrNotAbsolute)}, nil
		}
		answer.notes = expanded
		answer.notesNamedBy = byTheFile
	}

	// The environment is not a third rung after the file and the default: it
	// replaces the answer above outright, rather than taking a turn before it.
	// Set but empty says nothing, so answer stands as it was worked out. What
	// it names is read the way the file's places are: expanded, then judged.
	if said, ok := lookupEnv(StoreVariable); ok && said != "" {
		expanded := expandTilde(said, home)
		if !filepath.IsAbs(expanded) {
			return Reading{path: path, err: fmt.Errorf("%s=%q: %w", StoreVariable, said, ErrNotAbsolute)}, nil
		}
		answer.store = expanded
		answer.storeNamedBy = byTheEnvironment
	}
	if said, ok := lookupEnv(NotesVariable); ok && said != "" {
		expanded := expandTilde(said, home)
		if !filepath.IsAbs(expanded) {
			return Reading{path: path, err: fmt.Errorf("%s=%q: %w", NotesVariable, said, ErrNotAbsolute)}, nil
		}
		answer.notes = expanded
		answer.notesNamedBy = byTheEnvironment
	}
	return Reading{path: path, places: answer}, nil
}

// readGate is a ReadingTold that captures whichever single fact a Reading
// told it, so Read can hand that fact back the way it always has: the
// emplacements, or the problem that stood in the way.
type readGate struct {
	places Places
	err    error
}

func (g *readGate) ReadFrom(_ string)                 {}
func (g *readGate) Resolved(p Places)                 { g.places = p }
func (g *readGate) NotUnderstood(_ string, err error) { g.err = err }

// Read decodes <home>/.config/cairn/config.toml and hands back the places it
// names, unchanged from what the file wrote. It is a gate over the tolerant
// NewReading, a few lines above it: if a problem stood in the way, return it;
// otherwise return the emplacements it resolved.
func Read(home string, lookupEnv func(string) (string, bool)) (Places, error) {
	reading, err := NewReading(home, lookupEnv)
	if err != nil {
		return Places{}, err
	}
	gate := &readGate{}
	reading.Tell(gate)
	return gate.places, gate.err
}

// Told is the membrane of a Places: one method per fact it has to tell about
// each emplacement it resolved, absence of a namer included. Whoever wants to
// know who named a place implements this and is told it; nothing reaches
// into Places to read it directly.
type Told interface {
	StoreNamedByTheFile(string)
	StoreNamedByTheEnvironment(string)
	StoreNamedByNothing(string)
	NotesNamedByTheFile(string)
	NotesNamedByTheEnvironment(string)
	NotesNamedByNothing(string)
}

// Tell says, for each emplacement, which of the three provenances named it,
// carrying the emplacement it resolved to.
func (p Places) Tell(to Told) {
	switch p.storeNamedBy {
	case byTheFile:
		to.StoreNamedByTheFile(p.store)
	case byTheEnvironment:
		to.StoreNamedByTheEnvironment(p.store)
	default:
		to.StoreNamedByNothing(p.store)
	}
	switch p.notesNamedBy {
	case byTheFile:
		to.NotesNamedByTheFile(p.notes)
	case byTheEnvironment:
		to.NotesNamedByTheEnvironment(p.notes)
	default:
		to.NotesNamedByNothing(p.notes)
	}
}

// expandTilde replaces a "~" at the very start of value with home. Anywhere
// else, "~" is just a character the file wrote — this is not shell syntax.
func expandTilde(value, home string) string {
	rest, ok := strings.CutPrefix(value, "~/")
	if !ok {
		return value
	}
	return filepath.Join(home, rest)
}
