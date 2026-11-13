package engine

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cairn/internal/config"
	"cairn/internal/domain"
	"cairn/internal/theme"
)

// ConfigChanged is what setting a key or taking it out answers: the key, the value in force now and where it comes from, the file written, whether it was the global configuration, and whether the key was set rather than taken out.
type ConfigChanged struct {
	Key    string `json:"key"`
	Value  string `json:"value" jsonschema:"the value in force now"`
	From   string `json:"from" jsonschema:"where the value in force comes from: set in the store's settings, from the global configuration, or by default"`
	File   string `json:"file" jsonschema:"the file written"`
	Global bool   `json:"global"`
	Set    bool   `json:"set" jsonschema:"true when the key was set, false when it was taken out"`
}

var errAnsweredKey = errors.New("is the workspace's to say: name it in [workspaces], or set CAIRN_STORE or CAIRN_NOTES for one run")

var errStoreKey = errors.New("belongs to a store: it is said in the settings beside it, never in the global configuration")

var errEmptyGlobal = errors.New("is said empty only in a store's settings, for no label there: take it out of the global configuration instead")

var configKeyArguments = []Argument{
	{Name: "key", Kind: Text, Positional: true, Required: true, Means: "the key, dotted under its table: pages.time, window.theme, focus.session"},
}

var configGlobal = Argument{Name: "global", Kind: Bool, Means: "write the global configuration rather than the settings of the store in use"}

// ConfigSet writes one key of the configuration, in the settings of the store in use or in the global configuration, one line of the file and nothing else.
var ConfigSet = Operation{
	Thing:   "config",
	Verb:    "set",
	Short:   "Set a key of the configuration, one line of its file",
	Does:    "Set one key of the configuration to a value it takes, in the settings of the store in use when its workspace or the settings beside it say so, otherwise in the global configuration; --global writes the global one. The file is edited line by line, every comment kept, and nothing is written unless it reads back as meant. A key of the machine goes in the global configuration, the colour a workspace wears in its settings.",
	Example: "cairn config set pages.time hidden",
	Level:   Write,
	Needs:   AConfiguration,
	Again:   ChangesNothingMore,
	Answers: ConfigChanged{},
	Takes: append(append([]Argument{}, configKeyArguments...),
		Argument{Name: "value", Kind: Text, Positional: true, Required: true, Means: "the value, as the key takes it: a word, true or false, a duration, a day, a number, a path"},
		configGlobal),
	Run: func(on On, in Input) (Answer, error) { return changingConfig(on.Home, in, true) },
}

// ConfigUnset takes one key out of the settings of the store in use or out of the global configuration, its line and nothing else, so the value under it answers again.
var ConfigUnset = Operation{
	Thing:   "config",
	Verb:    "unset",
	Short:   "Take a key out of the configuration, its line and nothing else",
	Does:    "Take one key out of the settings of the store in use, or out of the global configuration with --global, so the value under it answers again: the global one, or the default. Its line goes and nothing else; a comment above it stays.",
	Example: "cairn config unset pages.time",
	Level:   Write,
	Needs:   AConfiguration,
	Again:   ChangesNothingMore,
	Answers: ConfigChanged{},
	Takes:   append(append([]Argument{}, configKeyArguments...), configGlobal),
	Run:     func(on On, in Input) (Answer, error) { return changingConfig(on.Home, in, false) },
}

// ConfigFill writes into a file of the configuration every key it can hold and does not say yet, as a comment carrying the value in force and where it comes from, so the file shows what can be changed in it.
var ConfigFill = Operation{
	Thing:   "config",
	Verb:    "fill",
	Short:   "Write every key a file can hold, as comments",
	Does:    "Write into the settings of the store in use, or into the global configuration with --global, every key the file can hold and does not say yet, as a comment carrying the value in force and where it comes from; uncommenting a line sets it. Lines and comments already there stay as they are.",
	Example: "cairn config fill --global",
	Level:   Write,
	Needs:   AConfiguration,
	Again:   ChangesNothingMore,
	Answers: ConfigWritten{},
	Takes:   []Argument{configGlobal},
	Run:     func(on On, in Input) (Answer, error) { return fillingConfig(on.Home, in.Yes("global")) },
}

func fillingConfig(h Home, global bool) (Answer, error) {
	held := map[string][]config.Key{}
	target := ""
	for _, key := range config.Keys() {
		path, inGlobal, err := h.writtenFor(key, global)
		if err != nil || inGlobal != global {
			continue
		}
		held[path] = append(held[path], key)
		target = path
	}
	if target == "" {
		return Answer{}, Refusing("config-nothing-to-fill", "no key of the configuration is written in that file")
	}
	change := func(was []byte) ([]byte, error) {
		said := was
		for _, key := range held[target] {
			value, from := h.inForce(key)
			literal, err := h.literal(key, value, global)
			if err != nil {
				literal = `""`
			}
			said = config.WithCommented(said, key.Name, literal, from)
		}
		return said, nil
	}
	if err := rewritten(target, h.Dir, !global, change); err != nil {
		return Answer{}, err
	}
	return Answer{Written: &target}, nil
}

func changingConfig(h Home, in Input, set bool) (Answer, error) {
	key, ok := config.KeyNamed(in.Given["key"])
	if !ok {
		return Answer{}, fmt.Errorf("%q %w: %s", in.Given["key"], config.ErrNotAKey, keysSaid())
	}
	path, global, err := h.writtenFor(key, in.Yes("global"))
	if err != nil {
		return Answer{}, err
	}
	change := func(was []byte) ([]byte, error) {
		said, err := config.WithoutKey(was, key.Name)
		if errors.Is(err, config.ErrNotSet) {
			return nil, fmt.Errorf("%s %w %s: there is nothing to take out", key.Name, config.ErrNotSet, path)
		}
		return said, err
	}
	if set {
		literal, err := h.literal(key, in.Given["value"], global)
		if err != nil {
			return Answer{}, err
		}
		change = func(was []byte) ([]byte, error) { return config.WithKey(was, key.Name, literal) }
	}
	if err := rewritten(path, h.Dir, !global, change); err != nil {
		return Answer{}, err
	}
	changed := ConfigChanged{Key: key.Name, File: path, Global: global, Set: set}
	changed.Value, changed.From = h.inForce(key)
	return Answer{ConfigChanged: &changed}, nil
}

func keysSaid() string {
	var names []string
	for _, key := range config.Keys() {
		names = append(names, key.Name)
	}
	return "cairn sets " + strings.Join(names, ", ")
}

func (h Home) writtenFor(key config.Key, global bool) (string, bool, error) {
	file, err := h.file()
	if err != nil {
		return "", false, err
	}
	if key.Answered {
		return "", false, fmt.Errorf("%s %w", key.Name, errAnsweredKey)
	}
	if global && key.Store {
		return "", false, fmt.Errorf("%s %w", key.Name, errStoreKey)
	}
	if global || key.Machine {
		return file, true, nil
	}
	settings, named, err := h.settingsHere()
	if err != nil {
		return "", false, err
	}
	if _, there := os.Stat(settings); key.Store || named || there == nil {
		return settings, false, nil
	}
	return file, true, nil
}

func (h Home) settingsHere() (string, bool, error) {
	if k := h.Kept; k != nil {
		return config.SettingsBeside(k.Store.At), k.Workspace != "" && !k.Implicit, nil
	}
	found, err := h.configuration()
	if err != nil {
		return "", false, err
	}
	a, err := h.answering(found)
	if err != nil {
		return "", false, err
	}
	return config.SettingsBeside(a.store.At), a.workspace != "" && !a.implicit(), nil
}

func (h Home) literal(key config.Key, value string, global bool) (string, error) {
	if global && value == "" && key.Kind == config.KindText {
		return "", fmt.Errorf("%s %w", key.Name, errEmptyGlobal)
	}
	literal, err := key.LiteralHere(value)
	if err != nil {
		return "", err
	}
	switch key.Kind {
	case config.KindDuration:
		if length, err := time.ParseDuration(value); err != nil || length <= 0 {
			return "", config.Refused("%s: %q is not a duration — write 1h, 45m, 1h30m", key.Name, value)
		}
	case config.KindColour:
		if _, err := domain.ParseColour(value); err != nil {
			return "", fmt.Errorf("%s %w", key.Name, err)
		}
	case config.KindTheme:
		if _, err := theme.Find(themesIn(h.Dir), value); err != nil {
			return "", Refusing("config-value", fmt.Sprintf("%s: %v", key.Name, err))
		}
	}
	return literal, nil
}

func rewritten(path, home string, settings bool, change func([]byte) ([]byte, error)) error {
	was, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := config.Readable(was, home, settings); err != nil {
		return fmt.Errorf("%s cannot be read, and is left for the hand: %w", path, err)
	}
	said, err := change(was)
	if err != nil {
		return err
	}
	if err := config.Readable(said, home, settings); err != nil {
		return err
	}
	return writtenWhole(path, said)
}

func writtenWhole(path string, said []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	beside, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(beside.Name()) }()
	if _, err := beside.Write(said); err != nil {
		_ = beside.Close()
		return err
	}
	if err := beside.Close(); err != nil {
		return err
	}
	if err := os.Chmod(beside.Name(), 0o600); err != nil {
		return err
	}
	return os.Rename(beside.Name(), path)
}

func (h Home) inForce(key config.Key) (string, string) {
	read, _, err := h.settled()
	if err != nil {
		return "", ""
	}
	from := whereSaid(read, key.Name)
	switch key.Name {
	case "store", "notes", "plugin":
		return placeOf(read, key.Name), from
	case "workspace.colour":
		return read.Colour(), from
	case "workspace.name":
		return read.Name(), from
	}
	value := valuesOf(read)[key.Name]
	if value == NotSaid {
		value = ""
	}
	return value, from
}

func placeOf(read config.Places, name string) string {
	gate := &placesGate{}
	read.Tell(gate)
	switch name {
	case "store":
		return gate.store.At
	case "notes":
		return gate.notes.At
	}
	return read.Plugin()
}

func configChangedText(c ConfigChanged) string {
	if c.Set {
		key, _ := config.KeyNamed(c.Key)
		value := c.Value
		if key.Kind != config.KindBool && key.Kind != config.KindNumber && key.Kind != config.KindDay {
			value = strconv.Quote(value)
		}
		return fmt.Sprintf("%s = %s, written in %s\n", c.Key, value, c.File)
	}
	value := c.Value
	if value == "" {
		value = "not said"
	}
	return fmt.Sprintf("%s is no longer said in %s: %s, %s\n", c.Key, c.File, value, c.From)
}
