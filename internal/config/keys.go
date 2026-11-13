package config

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

// Key is one key of the configuration as a person sets it: its dotted name, the kind of value it takes and the words it takes when it takes words, whether it belongs to the machine or to a store alone, whether it is a place the workspace answering gives, shown and never written, and the section and words a surface names it by.
type Key struct {
	Name     string
	Kind     Kind
	Words    []string
	Machine  bool
	Store    bool
	Answered bool
	Section  string
	Says     string
	Unit     string
}

// The sections the keys are shown under, in the order they are shown.
const (
	SectionAppearance = "Appearance"
	SectionPages      = "Tabs"
	SectionTime       = "Time"
	SectionArchive    = "Archive"
	SectionAgents     = "Agents"
	SectionPlaces     = "Places"
)

// Sections is every section the keys are shown under, in order.
func Sections() []string {
	return []string{SectionAppearance, SectionPages, SectionTime, SectionArchive, SectionAgents, SectionPlaces}
}

var keys = []Key{
	{Name: "window.theme", Kind: KindTheme, Section: SectionAppearance, Says: "Theme"},
	{Name: "window.intro", Kind: KindIntro, Words: Intros(), Section: SectionAppearance, Says: "Intro"},
	{Name: "workspace.colour", Kind: KindColour, Store: true, Section: SectionAppearance, Says: "Colour"},
	{Name: "workspace.name", Kind: KindText, Store: true, Section: SectionAppearance, Says: "Name"},
	{Name: "window.keep_sizes", Kind: KindBool, Section: SectionAppearance, Says: "Keep column sizes"},
	{Name: "window.ghost", Kind: KindPercent, Section: SectionAppearance, Says: "What is dragged shows at", Unit: "%"},
	{Name: "pages.tasks", Kind: KindBool, Section: SectionPages, Says: "Tasks"},
	{Name: "pages.objectives", Kind: KindBool, Section: SectionPages, Says: "Objectives"},
	{Name: "pages.time", Kind: KindWords, Words: []string{TimeOn, TimeHidden, TimeOff}, Section: SectionPages, Says: "Time"},
	{Name: "series.on", Kind: KindBool, Section: SectionPages, Says: "Series"},
	{Name: "time.weekends", Kind: KindWords, Words: []string{WeekendsShown, WeekendsHidden}, Section: SectionTime, Says: "Weekends"},
	{Name: "time.deleted", Kind: KindWords, Words: []string{TimeKept, TimeTaken}, Section: SectionTime, Says: "A deleted task's time"},
	{Name: "focus.session", Kind: KindDuration, Section: SectionTime, Says: "Focus session asks after"},
	{Name: "archive.start", Kind: KindDay, Section: SectionArchive, Says: "Archive on its own, counting from"},
	{Name: "archive.days", Kind: KindNumber, Section: SectionArchive, Says: "In periods of", Unit: "days"},
	{Name: "approach.label", Kind: KindText, Section: SectionAgents, Says: "Label asking for the approach first"},
	{Name: "store", Kind: KindPath, Machine: true, Answered: true, Section: SectionPlaces, Says: "Store"},
	{Name: "notes", Kind: KindPath, Machine: true, Answered: true, Section: SectionPlaces, Says: "Notes"},
	{Name: "plugin", Kind: KindPath, Machine: true, Section: SectionPlaces, Says: "Claude Code plugin"},
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
