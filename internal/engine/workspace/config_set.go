package workspace

import (
	"cairn/internal/store/files"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// ErrAnsweredKey is returned when config set names a key a workspace answers.
var ErrAnsweredKey = errors.New("is the workspace's to say: name it in [workspaces], or set CAIRN_STORE or CAIRN_NOTES for one run")

// ErrStoreKey is returned when config set names, in the global configuration, a key of a store.
var ErrStoreKey = errors.New("belongs to a store: it is said in the settings beside it, never in the global configuration")

// ErrEmptyGlobal is returned when config set says a key empty in the global configuration.
var ErrEmptyGlobal = errors.New("is said empty only in a store's settings, for no label there: take it out of the global configuration instead")

// KeysSaid names every key cairn sets.
func KeysSaid() string {
	var names []string
	for _, key := range files.Keys() {
		names = append(names, key.Name)
	}
	return "cairn sets " + strings.Join(names, ", ")
}

// Rewritten changes a configuration file, and refuses one it cannot read.
func Rewritten(path, home string, settings bool, change func([]byte) ([]byte, error)) error {
	was, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := files.Readable(was, home, settings); err != nil {
		return fmt.Errorf("%s cannot be read, and is left for the hand: %w", path, err)
	}
	said, err := change(was)
	if err != nil {
		return err
	}
	if err := files.Readable(said, home, settings); err != nil {
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

// PlaceOf is the store or the notes the configuration read names.
func PlaceOf(read files.Places, name string) string {
	gate := &PlacesGate{}
	read.Tell(gate)
	switch name {
	case "store":
		return gate.Store.At
	case "notes":
		return gate.Notes.At
	}
	return read.Plugin()
}

// ConfigChangedText is config set for a person.
func ConfigChangedText(c ConfigChanged) string {
	if c.Set {
		key, _ := files.KeyNamed(c.Key)
		value := c.Value
		if key.Kind != files.KindBool && key.Kind != files.KindNumber && key.Kind != files.KindDay {
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
