package engine

import (
	"cairn/internal/engine/label"
	"cairn/internal/engine/workspace"
	"errors"
	"fmt"
	"os"
	"time"

	"cairn/internal/store/files"
	"cairn/internal/theme"
)

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
	Level:   write,
	Needs:   aConfiguration,
	Again:   ChangesNothingMore,
	Answers: workspace.ConfigChanged{},
	Takes: append(append([]Argument{}, configKeyArguments...),
		Argument{Name: "value", Kind: Text, Positional: true, Required: true, Means: "the value, as the key takes it: a word, true or false, a duration, a day, a number, a path"},
		configGlobal),
	Run: func(on workedOn, in Input) (Answer, error) { return changingConfig(on.Home, in, true) },
}

// ConfigUnset takes one key out of the settings of the store in use or out of the global configuration, its line and nothing else, so the value under it answers again.
var ConfigUnset = Operation{
	Thing:   "config",
	Verb:    "unset",
	Short:   "Take a key out of the configuration, its line and nothing else",
	Does:    "Take one key out of the settings of the store in use, or out of the global configuration with --global, so the value under it answers again: the global one, or the default. Its line goes and nothing else; a comment above it stays.",
	Example: "cairn config unset pages.time",
	Level:   write,
	Needs:   aConfiguration,
	Again:   ChangesNothingMore,
	Answers: workspace.ConfigChanged{},
	Takes:   append(append([]Argument{}, configKeyArguments...), configGlobal),
	Run:     func(on workedOn, in Input) (Answer, error) { return changingConfig(on.Home, in, false) },
}

// ConfigFill writes into a file of the configuration every key it can hold and does not say yet, as a comment carrying the value in force and where it comes from, so the file shows what can be changed in it.
var ConfigFill = Operation{
	Thing:   "config",
	Verb:    "fill",
	Short:   "Write every key a file can hold, as comments",
	Does:    "Write into the settings of the store in use, or into the global configuration with --global, every key the file can hold and does not say yet, as a comment carrying the value in force and where it comes from; uncommenting a line sets it. Lines and comments already there stay as they are.",
	Example: "cairn config fill --global",
	Level:   write,
	Needs:   aConfiguration,
	Again:   ChangesNothingMore,
	Answers: workspace.ConfigWritten{},
	Takes:   []Argument{configGlobal},
	Run:     func(on workedOn, in Input) (Answer, error) { return fillingConfig(on.Home, in.yes("global")) },
}

func fillingConfig(h Home, global bool) (Answer, error) {
	held := map[string][]files.Key{}
	target := ""
	for _, key := range files.Keys() {
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
			said = files.WithCommented(said, key.Name, literal, from)
		}
		return said, nil
	}
	if err := workspace.Rewritten(target, h.Dir, !global, change); err != nil {
		return Answer{}, err
	}
	return Answer{Written: &target}, nil
}

func changingConfig(h Home, in Input, set bool) (Answer, error) {
	key, ok := files.KeyNamed(in.Given["key"])
	if !ok {
		return Answer{}, fmt.Errorf("%q %w: %s", in.Given["key"], files.ErrNotAKey, workspace.KeysSaid())
	}
	path, global, err := h.writtenFor(key, in.yes("global"))
	if err != nil {
		return Answer{}, err
	}
	change := func(was []byte) ([]byte, error) {
		said, err := files.WithoutKey(was, key.Name)
		if errors.Is(err, files.ErrNotSet) {
			return nil, fmt.Errorf("%s %w %s: there is nothing to take out", key.Name, files.ErrNotSet, path)
		}
		return said, err
	}
	if set {
		literal, err := h.literal(key, in.Given["value"], global)
		if err != nil {
			return Answer{}, err
		}
		change = func(was []byte) ([]byte, error) { return files.WithKey(was, key.Name, literal) }
	}
	if err := workspace.Rewritten(path, h.Dir, !global, change); err != nil {
		return Answer{}, err
	}
	changed := workspace.ConfigChanged{Key: key.Name, File: path, Global: global, Set: set}
	changed.Value, changed.From = h.inForce(key)
	return Answer{ConfigChanged: &changed}, nil
}

func (h Home) writtenFor(key files.Key, global bool) (string, bool, error) {
	file, err := h.file()
	if err != nil {
		return "", false, err
	}
	if key.Answered {
		return "", false, fmt.Errorf("%s %w", key.Name, workspace.ErrAnsweredKey)
	}
	if global && key.Store {
		return "", false, fmt.Errorf("%s %w", key.Name, workspace.ErrStoreKey)
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
		return files.SettingsBeside(k.Store.At), k.Workspace != "" && !k.Implicit, nil
	}
	found, err := h.configuration()
	if err != nil {
		return "", false, err
	}
	a, err := h.answering(found)
	if err != nil {
		return "", false, err
	}
	return files.SettingsBeside(a.Store.At), a.Workspace != "" && !a.Implicit(), nil
}

func (h Home) literal(key files.Key, value string, global bool) (string, error) {
	if global && value == "" && key.Kind == files.KindText {
		return "", fmt.Errorf("%s %w", key.Name, workspace.ErrEmptyGlobal)
	}
	literal, err := key.LiteralHere(value)
	if err != nil {
		return "", err
	}
	switch key.Kind {
	case files.KindDuration:
		if length, err := time.ParseDuration(value); err != nil || length <= 0 {
			return "", files.Refused("%s: %q is not a duration — write 1h, 45m, 1h30m", key.Name, value)
		}
	case files.KindColour:
		if _, err := label.ParseColour(value); err != nil {
			return "", fmt.Errorf("%s %w", key.Name, err)
		}
	case files.KindTheme:
		if _, err := theme.Find(workspace.ThemesIn(h.Dir), value); err != nil {
			return "", Refusing("config-value", fmt.Sprintf("%s: %v", key.Name, err))
		}
	}
	return literal, nil
}

func (h Home) inForce(key files.Key) (string, string) {
	read, _, err := h.settled()
	if err != nil {
		return "", ""
	}
	from := workspace.WhereSaid(read, key.Name)
	switch key.Name {
	case "store", "notes", "plugin":
		return workspace.PlaceOf(read, key.Name), from
	case "workspace.colour":
		return read.Colour(), from
	case "workspace.name":
		return read.Name(), from
	}
	value := workspace.ValuesOf(read)[key.Name]
	if value == workspace.NotSaid {
		value = ""
	}
	return value, from
}
