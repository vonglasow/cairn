package files

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Kind is what sort of value a key of the configuration takes, which says how it is written and how a surface offers it.
type Kind string

// The kinds a key of the configuration takes.
const (
	KindBool     Kind = "bool"
	KindWords    Kind = "words"
	KindDuration Kind = "duration"
	KindDay      Kind = "day"
	KindNumber   Kind = "number"
	KindText     Kind = "text"
	KindPath     Kind = "path"
	KindColour   Kind = "colour"
	KindTheme    Kind = "theme"
	KindIntro    Kind = "intro"
	KindPercent  Kind = "percent"
)

// Key is one key of the configuration as a person sets it: its dotted name, the kind of value it takes and the words it takes when it takes words, whether it belongs to the machine or to a store alone, whether it is a place the workspace answering gives, shown and never written, and how a surface names it: its section, its name, a line saying what it does, its unit, the words it shows for the words it takes, and whether it is shown as a switch, on for the first word.
type Key struct {
	Name     string
	Kind     Kind
	Words    []string
	Machine  bool
	Store    bool
	Answered bool
	Section  string
	Says     string
	Help     string
	Unit     string
	Shown    []string
	Switch   bool
}

// The sections the keys are shown under, in the order they are shown.
const (
	SectionAppearance = "Look"
	SectionPages      = "Tabs"
	SectionTime       = "Time"
	SectionArchive    = "Archive"
	SectionAgents     = "Agent"
	SectionPlaces     = "Files"
)

// Sections is every section the keys are shown under, in order.
func Sections() []string {
	return []string{SectionAppearance, SectionPages, SectionTime, SectionArchive, SectionAgents, SectionPlaces}
}

var keys = []Key{
	{Name: "window.theme", Kind: KindTheme, Section: SectionAppearance, Says: "Theme", Help: "The colours of the window and the terminal"},
	{Name: "window.intro", Kind: KindIntro, Words: Intros(), Section: SectionAppearance, Says: "Intro", Help: "What plays when the window opens"},
	{Name: "workspace.colour", Kind: KindColour, Store: true, Section: SectionAppearance, Says: "Workspace colour", Help: "The dot beside its name, in the menu and the tabs"},
	{Name: "workspace.name", Kind: KindText, Store: true, Section: SectionAppearance, Says: "Workspace name", Help: "Shown in the menu and the title bar"},
	{Name: "window.keep_sizes", Kind: KindBool, Section: SectionAppearance, Says: "Remember column widths", Help: "The widths you drag stay the next time"},
	{Name: "window.ghost", Kind: KindPercent, Section: SectionAppearance, Says: "Dragged card opacity", Help: "How much of a card shows while you drag it", Unit: "%"},
	{Name: "pages.tasks", Kind: KindBool, Section: SectionPages, Says: "Tasks", Help: "The board and its columns"},
	{Name: "pages.objectives", Kind: KindBool, Section: SectionPages, Says: "Objectives", Help: "The objectives and their tree"},
	{Name: "pages.time", Kind: KindWords, Words: []string{TimeOn, TimeHidden, TimeOff}, Shown: []string{"Shown", "Counted only", "Off"}, Section: SectionPages, Says: "Time", Help: "The timesheet, or only the time counted"},
	{Name: "series.on", Kind: KindBool, Section: SectionPages, Says: "Series", Help: "Tasks that come back, and their tab"},
	{Name: "time.weekends", Kind: KindWords, Words: []string{WeekendsShown, WeekendsHidden}, Shown: []string{"Shown", "Hidden"}, Section: SectionTime, Says: "Weekends", Help: "In the week and month of the Time tab"},
	{Name: "time.deleted", Kind: KindWords, Words: []string{TimeKept, TimeTaken}, Switch: true, Section: SectionTime, Says: "Count the time of deleted tasks", Help: "On: the time spent on a task you delete stays in the Time tab. Off: it goes with the task."},
	{Name: "focus.session", Kind: KindDuration, Section: SectionTime, Says: "Ask if still working after", Help: "A focus held this long asks to be renewed"},
	{Name: "archive.start", Kind: KindDay, Section: SectionArchive, Says: "Archive done tasks, starting", Help: "Empty: done tasks stay on the board"},
	{Name: "archive.days", Kind: KindNumber, Section: SectionArchive, Says: "Every", Help: "How often the done tasks are archived", Unit: "days"},
	{Name: "approach.label", Kind: KindText, Section: SectionAgents, Says: "Label that asks the agent for an approach first", Help: "A task carrying it is not coded before you agree on how"},
	{Name: "store", Kind: KindPath, Machine: true, Answered: true, Section: SectionPlaces, Says: "Board file", Help: "Where the tasks are kept"},
	{Name: "notes", Kind: KindPath, Machine: true, Answered: true, Section: SectionPlaces, Says: "Notes folder", Help: "Where each task's notes are written"},
	{Name: "plugin", Kind: KindPath, Machine: true, Section: SectionPlaces, Says: "Claude Code plugin folder", Help: "Where cairn writes its plugin for Claude Code"},
}

// Keys is every key a person sets, in the order the sections show them; the workspaces are not among them, which cairn workspace add, use and remove write.
func Keys() []Key { return slices.Clone(keys) }

// KeyNamed is the key of the catalogue by its dotted name, and whether there is one.
func KeyNamed(name string) (Key, bool) {
	at := slices.IndexFunc(keys, func(k Key) bool { return k.Name == name })
	if at < 0 {
		return Key{}, false
	}
	return keys[at], true
}

// ErrNotAKey is returned when a key is set that the catalogue does not hold.
var ErrNotAKey = errors.New("is not a key cairn sets")

// ErrValue is returned when a value is set that its key does not take.
var ErrValue = errors.New("is not what the key takes")

type valueError struct{ said string }

func (v valueError) Error() string        { return v.said }
func (v valueError) Is(target error) bool { return target == ErrValue }

// Refused is a value its key does not take, said in the words given.
func Refused(format string, args ...any) error { return valueError{said: fmt.Sprintf(format, args...)} }

// Literal is the value as the file writes it for this key, or why the key does not take it, in words naming the key.
func (k Key) Literal(value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("%s %w", k.Name, ErrEmpty)
	}
	switch k.Kind {
	case KindBool:
		if value != "true" && value != "false" {
			return "", Refused("%s: %q is true or false", k.Name, value)
		}
		return value, nil
	case KindWords, KindIntro:
		if !slices.Contains(k.Words, value) {
			return "", Refused("%q is not a word %s takes (%s)", value, k.Name, strings.Join(k.Words, ", "))
		}
	case KindDay:
		var day toml.LocalDate
		if day.UnmarshalText([]byte(value)) != nil {
			return "", Refused("%s: %q is not a day — write 2026-10-21", k.Name, value)
		}
		return day.String(), nil
	case KindNumber:
		if n, err := strconv.Atoi(value); err != nil || n < 1 {
			return "", Refused("%s: %q is not a number of days, one or more", k.Name, value)
		}
		return value, nil
	case KindPercent:
		if n, err := strconv.Atoi(value); err != nil || n < LeastGhost || n > MostGhost {
			return "", Refused("%s: %q is not a percentage from %d to %d", k.Name, value, LeastGhost, MostGhost)
		}
		return value, nil
	case KindPath:
		if _, err := Absolute(value, "/"); err != nil {
			return "", fmt.Errorf("%s %w", k.Name, err)
		}
	}
	return strconv.Quote(value), nil
}

// LiteralHere is the value as a store's settings write it: as Literal does, but for the approach label, which a store's settings may say empty for no label here.
func (k Key) LiteralHere(value string) (string, error) {
	if value == "" && k.Name == noLabelHere {
		return `""`, nil
	}
	return k.Literal(value)
}
