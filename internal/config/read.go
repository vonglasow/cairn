// Package config reads the file that says cairn's two places, and which of its pages are turned off.
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

// DefaultFolder and defaultFile make up the place the store goes when nothing names it: no file at all, or a file that says nothing under `store`.
const (
	DefaultFolder = ".cairn"
	defaultFile   = "cairn.db"
)

const defaultNotesFolder = "notes"

// StoreVariable is the environment variable that replaces whatever the file or the default answered, named here because this package now owns it — the surface that reads it lives elsewhere, but the name is spelled once.
const StoreVariable = "CAIRN_STORE"

// NotesVariable is the environment variable that replaces whatever the file or the default answered for notes — the twin of StoreVariable, for the other place this package resolves.
const NotesVariable = "CAIRN_NOTES"

// ErrUnknownKey is returned when the file names a key cairn does not know — the decoding is strict, so a typo or a leftover from another tool is refused rather than quietly ignored.
var ErrUnknownKey = errors.New("not a key cairn knows")

// ErrNotTOML is returned when the file's content cannot be made sense of as TOML at all — not a specific rule the file broke, just the syntax itself.
var ErrNotTOML = errors.New("not a file cairn can make sense of")

// ErrArchive is returned when the file counts the days archiving waits for from something that is not a day, or in periods of no day at all.
var ErrArchive = errors.New("archiving counts from a day, written 2026-10-21, in periods of one day or more")

const defaultArchiveDays = 14

// ErrNotAbsolute is returned when a place the file names is not absolute, even once a leading "~" has been expanded — a configuration is read from everywhere, so what it names has to mean one thing regardless of where a command runs from.
var ErrNotAbsolute = errors.New("must be absolute")

type namedBy int

const (
	byNothing namedBy = iota
	byTheFile
	byTheEnvironment
)

// Places is what the configuration file says, exactly as it was written.
type Places struct {
	store        string
	notes        string
	storeNamedBy namedBy
	notesNamedBy namedBy
	pagesOff     []string
	plugin       string
	archiveStart string
	archiveDays  int
}

type file struct {
	Store   *string  `toml:"store"`
	Notes   *string  `toml:"notes"`
	Plugin  *string  `toml:"plugin"`
	Pages   *pages   `toml:"pages"`
	Archive *archive `toml:"archive"`
}

type archive struct {
	Start any    `toml:"start"`
	Days  *int64 `toml:"days"`
}

func (s *archive) read() (string, int, error) {
	if s == nil {
		return "", 0, nil
	}
	var start toml.LocalDate
	switch said := s.Start.(type) {
	case toml.LocalDate:
		start = said
	case string:
		if err := start.UnmarshalText([]byte(said)); err != nil {
			return "", 0, fmt.Errorf("archive.start %q: %w", said, ErrArchive)
		}
	default:
		return "", 0, fmt.Errorf("archive.start %v: %w", said, ErrArchive)
	}
	days := int64(defaultArchiveDays)
	if s.Days != nil {
		days = *s.Days
	}
	if days < 1 {
		return "", 0, fmt.Errorf("archive.days %d: %w", days, ErrArchive)
	}
	return start.String(), int(days), nil
}

type pages struct {
	Tasks      *bool `toml:"tasks"`
	Objectives *bool `toml:"objectives"`
	Time       *bool `toml:"time"`
}

func (p *pages) off() []string {
	if p == nil {
		return nil
	}
	var off []string
	for _, page := range []struct {
		name string
		on   *bool
	}{{"tasks", p.Tasks}, {"objectives", p.Objectives}, {"time", p.Time}} {
		if page.on != nil && !*page.on {
			off = append(off, page.name)
		}
	}
	return off
}

// Plugin is the folder the file names for the Claude Code plugin, and nothing when it names none: the plugin then goes beside the store.
func (p Places) Plugin() string { return p.plugin }

// Archive is the day the file counts from, written 2026-10-21, and how many days each period lasts; nothing when the file does not archive on its own.
func (p Places) Archive() (string, int, bool) {
	return p.archiveStart, p.archiveDays, p.archiveDays > 0
}

// PagesOff is the pages the file turns off, in the order the boards draw them: tasks, objectives, time.
func (p Places) PagesOff() []string { return p.pagesOff }

func acceptedKeys() []string {
	t := reflect.TypeFor[file]()
	keys := make([]string, 0, t.NumField())
	for i := range t.NumField() {
		keys = append(keys, t.Field(i).Tag.Get("toml"))
	}
	return keys
}

// Reading is the tolerant primitive Read is a gate over: it carries the path of the file it read, and either the emplacements the file named or the problem that stood in the way — never both, since emplacements resolved alongside a rule the file broke are not emplacements worth trusting.
type Reading struct {
	path   string
	places Places
	err    error
}

// ReadingTold is the membrane of a Reading: whoever wants to know what a reading found implements this and is told it — its emplacements, or the problem it hit naming the file it read.
type ReadingTold interface {
	ReadFrom(path string)
	Resolved(Places)
	NotUnderstood(path string, err error)
}

// Tell says which file this reading read from, told first and regardless of what came of it, then what it found: its emplacements if it made sense of the file, or — naming the path again — the problem that stood in the way.
func (r Reading) Tell(to ReadingTold) {
	to.ReadFrom(r.path)
	if r.err != nil {
		to.NotUnderstood(r.path, r.err)
		return
	}
	to.Resolved(r.places)
}

// FileIn is where the configuration file of a home is: <home>/.config/cairn/config.toml.
func FileIn(home string) string { return filepath.Join(home, ".config", "cairn", "config.toml") }

// NewReading decodes <home>/.config/cairn/config.toml, tolerating what it cannot make sense of by reporting it through the Reading it hands back rather than returning it.
func NewReading(home string, lookupEnv func(string) (string, bool)) (Reading, error) {
	path := FileIn(home)
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return Reading{}, err
	}

	var f file
	if err == nil {
		if err := toml.NewDecoder(bytes.NewReader(data)).DisallowUnknownFields().Decode(&f); err != nil {
			var strict *toml.StrictMissingError
			if errors.As(err, &strict) {
				return Reading{path: path, err: fmt.Errorf("%q: %w — cairn accepts %s",
					strings.Join(strict.Errors[0].Key(), "."), ErrUnknownKey, strings.Join(acceptedKeys(), ", "))}, nil
			}
			var decode *toml.DecodeError
			if errors.As(err, &decode) {
				return Reading{path: path, err: fmt.Errorf("%q: %w", path, ErrNotTOML)}, nil
			}
			return Reading{}, err
		}
	}

	start, days, err := f.Archive.read()
	if err != nil {
		return Reading{path: path, err: err}, nil
	}
	answer := Places{
		archiveStart: start,
		archiveDays:  days,
		store:        filepath.Join(home, DefaultFolder, defaultFile),
		notes:        filepath.Join(home, DefaultFolder, defaultNotesFolder),
		storeNamedBy: byNothing,
		notesNamedBy: byNothing,
		pagesOff:     f.Pages.off(),
	}
	for _, c := range []claim{
		{f.Store, &answer.store, &answer.storeNamedBy, byTheFile, ""},
		{f.Notes, &answer.notes, &answer.notesNamedBy, byTheFile, ""},
		{fromEnv(lookupEnv, StoreVariable), &answer.store, &answer.storeNamedBy, byTheEnvironment, StoreVariable},
		{fromEnv(lookupEnv, NotesVariable), &answer.notes, &answer.notesNamedBy, byTheEnvironment, NotesVariable},
		{f.Plugin, &answer.plugin, new(namedBy), byTheFile, ""},
	} {
		if c.said == nil {
			continue
		}
		expanded := expandTilde(*c.said, home)
		if !filepath.IsAbs(expanded) {
			return Reading{path: path, err: fmt.Errorf("%s: %w", c.quoted(), ErrNotAbsolute)}, nil
		}
		*c.place, *c.namedBy = expanded, c.by
	}
	return Reading{path: path, places: answer}, nil
}

type claim struct {
	said     *string
	place    *string
	namedBy  *namedBy
	by       namedBy
	variable string
}

func (c claim) quoted() string {
	if c.variable == "" {
		return fmt.Sprintf("%q", *c.said)
	}
	return fmt.Sprintf("%s=%q", c.variable, *c.said)
}

func fromEnv(lookupEnv func(string) (string, bool), name string) *string {
	if said, ok := lookupEnv(name); ok && said != "" {
		return &said
	}
	return nil
}

type readGate struct {
	places Places
	err    error
}

func (g *readGate) ReadFrom(_ string)                 {}
func (g *readGate) Resolved(p Places)                 { g.places = p }
func (g *readGate) NotUnderstood(_ string, err error) { g.err = err }

// Read decodes <home>/.config/cairn/config.toml and hands back the places it names, unchanged from what the file wrote.
func Read(home string, lookupEnv func(string) (string, bool)) (Places, error) {
	reading, err := NewReading(home, lookupEnv)
	if err != nil {
		return Places{}, err
	}
	gate := &readGate{}
	reading.Tell(gate)
	return gate.places, gate.err
}

// Told is the membrane of a Places: one method per fact it has to tell about each emplacement it resolved, absence of a namer included.
type Told interface {
	StoreNamedByTheFile(string)
	StoreNamedByTheEnvironment(string)
	StoreNamedByNothing(string)
	NotesNamedByTheFile(string)
	NotesNamedByTheEnvironment(string)
	NotesNamedByNothing(string)
}

// Tell says, for each emplacement, which of the three provenances named it, carrying the emplacement it resolved to.
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

func expandTilde(value, home string) string {
	rest, ok := strings.CutPrefix(value, "~/")
	if !ok {
		return value
	}
	return filepath.Join(home, rest)
}
