package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
)

// SettingsFile is the name of the file beside a store that holds the settings belonging to that store.
const SettingsFile = "settings.toml"

// SettingsHeader is what cairn workspace add writes in a workspace's settings: nothing set, and what the file is for.
const SettingsHeader = `# The settings of the store beside this file, written as in the global
# configuration: [pages], [archive], [focus], [approach], [time], [window].
# A key said here is used for this store instead of the global one. store,
# notes, plugin and [workspaces] belong to the machine, and stay there.
# Nothing is said yet: everything comes from the global configuration until
# this file says otherwise.
`

const noLabelHere = "approach.label"

// ErrMachineKey is returned when a store's settings say something that belongs to the machine, and so to the global configuration.
var ErrMachineKey = errors.New("belongs to the machine, not to a store")

// ErrEmpty is returned when a store's settings write a key as the empty string, which says nothing; only approach.label may, for no label here.
var ErrEmpty = errors.New("is said empty: write a value, or leave the key out")

// SettingsBeside is where the settings of the store given are: settings.toml, in the store's folder.
func SettingsBeside(store string) string { return filepath.Join(filepath.Dir(store), SettingsFile) }

// Settings is the store's settings read over the global configuration, nothing when there were none.
func (p Places) Settings() string { return p.settings }

// WrittenHere is every key the store's settings write, dotted under its table and sorted: those win over the global configuration.
func (p Places) WrittenHere() []string { return p.writtenHere }

// WithSettings is the configuration with the store's settings at path read over it, key by key; no file there leaves it as it was.
func (p Places) WithSettings(path string) (Places, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return p, nil
	}
	if err != nil {
		return Places{}, err
	}
	s, ignored, err := decoded(data)
	if err != nil {
		return Places{}, fmt.Errorf("%q: %w", path, err)
	}
	if key := machineKey(s); key != "" {
		return Places{}, fmt.Errorf("%s: %q %w: it goes in %s", path, key, ErrMachineKey, p.from)
	}
	if key := emptyIn(s); key != "" {
		return Places{}, fmt.Errorf("%s: %s %w", path, key, ErrEmpty)
	}
	if err := p.settle(p.said.under(s)); err != nil {
		return Places{}, fmt.Errorf("%s: %w", path, err)
	}
	p.ignored = slices.Clone(p.ignored)
	for _, one := range ignored {
		p.ignored = append(p.ignored, one+", in "+path)
	}
	p.settings, p.writtenHere = path, writtenIn(data)
	return p, nil
}

func storeLevel(t reflect.Type) bool {
	return t.Kind() == reflect.Pointer && t.Elem().Kind() == reflect.Struct
}

func machineKey(s file) string {
	said := reflect.ValueOf(s)
	for i := range said.NumField() {
		if !storeLevel(said.Field(i).Type()) && !said.Field(i).IsNil() {
			return said.Type().Field(i).Tag.Get("toml")
		}
	}
	return ""
}

func emptyIn(s file) string {
	said := reflect.ValueOf(s)
	for i := range said.NumField() {
		table := said.Field(i)
		if !storeLevel(table.Type()) || table.IsNil() {
			continue
		}
		for j := range table.Elem().NumField() {
			value := table.Elem().Field(j)
			key := said.Type().Field(i).Tag.Get("toml") + "." + table.Elem().Type().Field(j).Tag.Get("toml")
			if empty, ok := value.Interface().(*string); ok && empty != nil && *empty == "" && key != noLabelHere {
				return key
			}
		}
	}
	return ""
}

func (f file) under(s file) file {
	into, said := reflect.ValueOf(&f).Elem(), reflect.ValueOf(s)
	for i := range said.NumField() {
		table := said.Field(i)
		if !storeLevel(table.Type()) || table.IsNil() {
			continue
		}
		merged := reflect.New(table.Type().Elem())
		if !into.Field(i).IsNil() {
			merged.Elem().Set(into.Field(i).Elem())
		}
		for j := range table.Elem().NumField() {
			if value := table.Elem().Field(j); !value.IsNil() {
				merged.Elem().Field(j).Set(value)
			}
		}
		into.Field(i).Set(merged)
	}
	return f
}
