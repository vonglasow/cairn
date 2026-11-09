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
	"slices"
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

// ErrUnknownKey is what a key the file holds and this cairn does not know is said to be: passed over, never quietly.
var ErrUnknownKey = errors.New("not a key this cairn knows")

// ErrNotTOML is returned when the file's content cannot be made sense of as TOML at all — not a specific rule the file broke, just the syntax itself.
var ErrNotTOML = errors.New("not a file cairn can make sense of")

// ErrArchive is returned when the file counts the days archiving waits for from something that is not a day, or in periods of no day at all.
var ErrArchive = errors.New("archiving counts from a day, written 2026-10-21, in periods of one day or more")

// ErrTime is returned when the file says a deleted task does something with its time other than keep it or take it.
var ErrTime = errors.New(`a deleted task's time is "kept" or "taken"`)

// TimeKept and TimeTaken are what a deleted task may do with its time: leave it on the sheet, or take it with it.
const (
	TimeKept  = "kept"
	TimeTaken = "taken"
)

// ErrWeekends is returned when the file says the weekends are something other than shown or hidden.
var ErrWeekends = errors.New(`the weekends are "shown" or "hidden"`)

// WeekendsShown and WeekendsHidden are what a week may do with its Saturday and Sunday: draw them, or leave them out.
const (
	WeekendsShown  = "shown"
	WeekendsHidden = "hidden"
)

// ErrTimeMode is returned when the file says the time is something other than one of its three words.
var ErrTimeMode = errors.New(`the time is "on", "hidden" or "off"`)

// TimeOn, TimeHidden and TimeOff are what the time may be: recorded and shown, recorded and not shown, or neither.
const (
	TimeOn     = "on"
	TimeHidden = "hidden"
	TimeOff    = "off"
)

// ErrKeepSizes is returned when the file says whether the window keeps its sizes with something that is neither true nor false.
var ErrKeepSizes = errors.New("the window keeps its sizes or not: true or false")

// ErrFocus is returned when the file says a focus session lasts something that is not a duration.
var ErrFocus = errors.New("a focus session lasts a duration, written 1h")

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
	store            string
	notes            string
	storeNamedBy     namedBy
	notesNamedBy     namedBy
	pagesOff         []string
	timeMode         string
	plugin           string
	archiveStart     string
	archiveDays      int
	focusSession     string
	approach         string
	colour           string
	timeDeleted      string
	timeWeekends     string
	keepSizes        bool
	written          []string
	ignored          []string
	workspaces       []Workspace
	defaultWorkspace string
	fileStore        string
	from             string
	said             file
	settings         string
	writtenHere      []string
}

type file struct {
	Store      *string           `toml:"store"`
	Notes      *string           `toml:"notes"`
	Plugin     *string           `toml:"plugin"`
	Pages      *pages            `toml:"pages"`
	Archive    *archive          `toml:"archive"`
	Focus      *focus            `toml:"focus"`
	Approach   *approach         `toml:"approach"`
	Time       *timed            `toml:"time"`
	Window     *window           `toml:"window"`
	Workspace  *workspace        `toml:"workspace"`
	Workspaces map[string]string `toml:"workspaces"`
}

type workspace struct {
	Colour *string `toml:"colour"`
}

func (w *workspace) colour() string {
	if w == nil || w.Colour == nil {
		return ""
	}
	return *w.Colour
}

type window struct {
	KeepSizes any `toml:"keep_sizes"`
}

func (w *window) keepSizes() (bool, error) {
	if w == nil || w.KeepSizes == nil {
		return false, nil
	}
	keep, ok := w.KeepSizes.(bool)
	if !ok {
		return false, fmt.Errorf("window.keep_sizes %v: %w", w.KeepSizes, ErrKeepSizes)
	}
	return keep, nil
}

type timed struct {
	Deleted  *string `toml:"deleted"`
	Weekends *string `toml:"weekends"`
}

type focus struct {
	Session *string `toml:"session"`
}

type approach struct {
	Label *string `toml:"label"`
}

func (a *approach) label() string {
	if a == nil || a.Label == nil {
		return ""
	}
	return *a.Label
}

type archive struct {
	Start any    `toml:"start"`
	Days  *int64 `toml:"days"`
}

func (s *archive) read() (string, int, error) {
	if s == nil {
		return "", defaultArchiveDays, nil
	}
	days := int64(defaultArchiveDays)
	if s.Days != nil {
		days = *s.Days
	}
	if days < 1 {
		return "", 0, fmt.Errorf("archive.days %d: %w", days, ErrArchive)
	}
	var start toml.LocalDate
	switch said := s.Start.(type) {
	case nil:
		return "", int(days), nil
	case toml.LocalDate:
		start = said
	case string:
		if err := start.UnmarshalText([]byte(said)); err != nil {
			return "", 0, fmt.Errorf("archive.start %q: %w", said, ErrArchive)
		}
	default:
		return "", 0, fmt.Errorf("archive.start %v: %w", said, ErrArchive)
	}
	return start.String(), int(days), nil
}

type pages struct {
	Tasks      *bool `toml:"tasks"`
	Objectives *bool `toml:"objectives"`
	Time       any   `toml:"time"`
}

func (p *pages) time() (string, error) {
	if p == nil || p.Time == nil {
		return TimeOn, nil
	}
	switch said := p.Time.(type) {
	case bool:
		if said {
			return TimeOn, nil
		}
		return TimeHidden, nil
	case string:
		if slices.Contains([]string{TimeOn, TimeHidden, TimeOff}, said) {
			return said, nil
		}
	}
	return "", fmt.Errorf("pages.time %#v: %w", p.Time, ErrTimeMode)
}

func (p *pages) off(timeOn bool) []string {
	if p == nil {
		return nil
	}
	var off []string
	for _, page := range []struct {
		name string
		on   *bool
	}{{"tasks", p.Tasks}, {"objectives", p.Objectives}, {"time", &timeOn}} {
		if page.on != nil && !*page.on {
			off = append(off, page.name)
		}
	}
	return off
}

// Plugin is the folder the file names for the Claude Code plugin, and nothing when it names none: the plugin then goes beside the store.
func (p Places) Plugin() string { return p.plugin }

// Archive is the day the file counts from, written 2026-10-21, and how many days each period lasts, and whether anything is archived on its own: only when the file says the day.
func (p Places) Archive() (string, int, bool) {
	return p.archiveStart, p.archiveDays, p.archiveStart != ""
}

// Pages says whether each page is on, in the order the boards draw them: tasks, objectives, time.
func (p Places) Pages() [3]bool {
	on := [3]bool{true, true, true}
	for i, page := range []string{"tasks", "objectives", "time"} {
		on[i] = !slices.Contains(p.pagesOff, page)
	}
	return on
}

// TimeMode is what the time is: TimeOn unless the file says TimeHidden or TimeOff, false read as TimeHidden.
func (p Places) TimeMode() string { return p.timeMode }

// TimeDeleted is what a deleted task does with its time: TimeKept, unless the file says TimeTaken.
func (p Places) TimeDeleted() string { return p.timeDeleted }

// TimeWeekends says whether a week is drawn with its Saturday and Sunday.
func (p Places) TimeWeekends() string { return p.timeWeekends }

// FocusSession is how long a focus session lasts before it asks to be renewed, written 1h, exactly as the file wrote it; nothing when the file does not say.
func (p Places) FocusSession() string { return p.focusSession }

// ApproachLabel is the label asking for the approach to be agreed before any code, exactly as the file wrote it; nothing when the file does not say.
func (p Places) ApproachLabel() string { return p.approach }

// Colour is the colour the store's settings say its workspace wears, exactly as written; nothing when they say none, and never the global file's, which a store's colour is not.
func (p Places) Colour() string { return p.colour }

// KeepSizes says whether the window opens with the sizes it was left with: false unless the file says true.
func (p Places) KeepSizes() bool { return p.keepSizes }

// Written is every key the file writes, dotted under its table and sorted: what is not among them holds its default, or what the environment says.
func (p Places) Written() []string { return p.written }

// Ignored says, one sentence each and sorted, every key the file holds that this cairn does not know and passed over: a typo, or a key a newer cairn wrote.
func (p Places) Ignored() []string { return p.ignored }

func passedOver(strict *toml.StrictMissingError) []string {
	said := make([]string, 0, len(strict.Errors))
	for _, missing := range strict.Errors {
		key := missing.Key()
		said = append(said, fmt.Sprintf("%q is %s — %s; it is ignored", strings.Join(key, "."), ErrUnknownKey, belongs(key[len(key)-1])))
	}
	slices.Sort(said)
	return said
}

func writtenIn(data []byte) []string {
	var said map[string]any
	if toml.Unmarshal(data, &said) != nil {
		return nil
	}
	var keys []string
	for key, value := range said {
		table, ok := value.(map[string]any)
		if !ok {
			keys = append(keys, key)
			continue
		}
		for inner := range table {
			keys = append(keys, key+"."+inner)
		}
	}
	slices.Sort(keys)
	return keys
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

func belongs(key string) string {
	t := reflect.TypeFor[file]()
	for i := range t.NumField() {
		field := t.Field(i)
		if field.Tag.Get("toml") == key && field.Type.Kind() == reflect.Map {
			return fmt.Sprintf("%s is a table of its own, [%s]", key, key)
		}
		if field.Type.Kind() == reflect.Map {
			continue
		}
		if field.Tag.Get("toml") == key && field.Type.Elem().Kind() != reflect.Struct {
			return key + " belongs before any table"
		}
		if field.Type.Elem().Kind() != reflect.Struct {
			continue
		}
		table := field.Type.Elem()
		for j := range table.NumField() {
			if table.Field(j).Tag.Get("toml") == key {
				return fmt.Sprintf("%s belongs under [%s]", key, field.Tag.Get("toml"))
			}
		}
	}
	return "cairn accepts " + strings.Join(acceptedKeys(), ", ")
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
	var ignored []string
	if err == nil {
		f, ignored, err = decoded(data)
		if errors.Is(err, ErrNotTOML) {
			return Reading{path: path, err: fmt.Errorf("%q: %w", path, ErrNotTOML)}, nil
		}
		if err != nil {
			return Reading{}, err
		}
		ignored = append(ignored, storesOwn(data)...)
		f.Workspace = nil
	}

	answer := Places{
		from:         path,
		said:         f,
		store:        filepath.Join(home, DefaultFolder, defaultFile),
		notes:        filepath.Join(home, DefaultFolder, defaultNotesFolder),
		storeNamedBy: byNothing,
		notesNamedBy: byNothing,
		written:      writtenIn(data),
		ignored:      ignored,
	}
	if err := answer.settle(f); err != nil {
		return Reading{path: path, err: err}, nil
	}
	answer.workspaces, answer.defaultWorkspace, err = workspacesOf(f.Workspaces, home)
	if err != nil {
		return Reading{path: path, err: err}, nil
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
		expanded, err := Absolute(*c.said, home)
		if err != nil {
			return Reading{path: path, err: fmt.Errorf("%s: %w", c.quoted(), ErrNotAbsolute)}, nil
		}
		*c.place, *c.namedBy = expanded, c.by
		if c.said == f.Store {
			answer.fileStore = expanded
		}
	}
	return Reading{path: path, places: answer}, nil
}

func decoded(data []byte) (file, []string, error) {
	var f file
	err := toml.NewDecoder(bytes.NewReader(data)).DisallowUnknownFields().Decode(&f)
	if err == nil {
		return f, nil, nil
	}
	var ignored []string
	var strict *toml.StrictMissingError
	if errors.As(err, &strict) {
		ignored = passedOver(strict)
		f = file{}
		err = toml.NewDecoder(bytes.NewReader(data)).Decode(&f)
	}
	var decode *toml.DecodeError
	if errors.As(err, &decode) {
		return file{}, nil, ErrNotTOML
	}
	return f, ignored, err
}

func (p *Places) settle(f file) error {
	start, days, err := f.Archive.read()
	if err != nil {
		return err
	}
	deleted, err := f.Time.deleted()
	if err != nil {
		return err
	}
	weekends, err := f.Time.weekends()
	if err != nil {
		return err
	}
	keep, err := f.Window.keepSizes()
	if err != nil {
		return err
	}
	mode, err := f.Pages.time()
	if err != nil {
		return err
	}
	p.timeMode = mode
	p.archiveStart, p.archiveDays, p.timeDeleted, p.timeWeekends, p.keepSizes = start, days, deleted, weekends, keep
	p.focusSession, p.approach, p.pagesOff = f.Focus.session(), f.Approach.label(), f.Pages.off(mode == TimeOn)
	p.colour = f.Workspace.colour()
	return nil
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

func (f *focus) session() string {
	if f == nil || f.Session == nil {
		return ""
	}
	return *f.Session
}

func (t *timed) deleted() (string, error) {
	if t == nil || t.Deleted == nil || *t.Deleted == "" {
		return TimeKept, nil
	}
	if *t.Deleted != TimeKept && *t.Deleted != TimeTaken {
		return "", fmt.Errorf("time.deleted %q: %w", *t.Deleted, ErrTime)
	}
	return *t.Deleted, nil
}

func (t *timed) weekends() (string, error) {
	if t == nil || t.Weekends == nil || *t.Weekends == "" {
		return WeekendsShown, nil
	}
	if *t.Weekends != WeekendsShown && *t.Weekends != WeekendsHidden {
		return "", fmt.Errorf("time.weekends %q: %w", *t.Weekends, ErrWeekends)
	}
	return *t.Weekends, nil
}
