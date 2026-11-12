package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// DefaultWorkspaceKey is the key of [workspaces] that names the workspace used when nothing else says, and so a name no workspace can take.
const DefaultWorkspaceKey = "default"

// ErrNotWritten is returned when the file says its workspaces in a way cairn cannot change without breaking it, or is not TOML to begin with.
var ErrNotWritten = errors.New("cairn cannot change the workspaces of this file without breaking it; change it by hand")

// Workspace is one entry of [workspaces]: a name, and the folder holding its store and its notes, ~ expanded.
type Workspace struct {
	Name   string
	Folder string
}

// Workspaces is every workspace the file names, by name, the default key left out.
func (p Places) Workspaces() []Workspace { return p.workspaces }

// FileStore is the store the file names under store, ~ expanded, whatever answers over it; nothing when it names none.
func (p Places) FileStore() string { return p.fileStore }

// DefaultWorkspace is the name the file gives under default in [workspaces], exactly as written; nothing when it gives none.
func (p Places) DefaultWorkspace() string { return p.defaultWorkspace }

// Absolute is a place as cairn reads it: a leading ~ expanded under the home, and refused unless absolute.
func Absolute(said, home string) (string, error) {
	expanded := expandTilde(said, home)
	if !filepath.IsAbs(expanded) {
		return "", fmt.Errorf("%q: %w", said, ErrNotAbsolute)
	}
	return expanded, nil
}

func workspacesOf(said map[string]string, home string) ([]Workspace, string, error) {
	var named []Workspace
	for name, folder := range said {
		if name == DefaultWorkspaceKey {
			continue
		}
		at, err := Absolute(folder, home)
		if err != nil {
			return nil, "", fmt.Errorf("workspaces.%s %w", name, err)
		}
		named = append(named, Workspace{Name: name, Folder: at})
	}
	slices.SortFunc(named, func(a, b Workspace) int { return strings.Compare(a.Name, b.Name) })
	return named, said[DefaultWorkspaceKey], nil
}

const workspacesTable = "workspaces"

var (
	aHeading = regexp.MustCompile(`^\s*\[`)
	aDefault = regexp.MustCompile(`^\s*default\s*=`)
)

func blank(line string) bool   { return strings.TrimSpace(line) == "" }
func comment(line string) bool { return strings.HasPrefix(strings.TrimSpace(line), "#") }

func linesOf(data []byte) []string {
	lines := strings.SplitAfter(string(data), "\n")
	if last := len(lines) - 1; lines[last] == "" {
		return lines[:last]
	}
	lines[len(lines)-1] += "\n"
	return lines
}

// WithWorkspace is the file with one more workspace in [workspaces], at the end of the table, or in a table of its own after the rest; every other byte is the file's.
func WithWorkspace(data []byte, name, folder string) ([]byte, error) {
	entry := fmt.Sprintf("%s = %q\n", name, folder)
	lines := linesOf(data)
	start, end, found := regionOf(lines, workspacesTable)
	if !found {
		return checked([]byte(strings.Join(withTable(lines, workspacesTable), "")+entry), data, name, folder)
	}
	said := slices.Insert(slices.Clone(lines), insertedAt(lines, start, end), entry)
	return checked([]byte(strings.Join(said, "")), data, name, folder)
}

// WithDefault is the file with default in [workspaces] naming the workspace given, where it was said or under the table's heading; every other byte is the file's.
func WithDefault(data []byte, name string) ([]byte, error) {
	lines := linesOf(data)
	start, end, found := regionOf(lines, workspacesTable)
	if !found {
		return nil, ErrNotWritten
	}
	entry := fmt.Sprintf("%s = %q\n", DefaultWorkspaceKey, name)
	said := slices.Clone(lines)
	if at := slices.IndexFunc(lines[start:end], aDefault.MatchString); at >= 0 {
		said[start+at] = entry
	} else {
		said = slices.Insert(said, start, entry)
	}
	return checked([]byte(strings.Join(said, "")), data, DefaultWorkspaceKey, name)
}

func checked(said, was []byte, key, value string) ([]byte, error) {
	var before, after struct {
		Workspaces map[string]any `toml:"workspaces"`
	}
	if toml.Unmarshal(was, &before) != nil || toml.Unmarshal(said, &after) != nil || after.Workspaces[key] != value {
		return nil, ErrNotWritten
	}
	return said, nil
}

// WithoutWorkspace is the file with the workspace named taken out of [workspaces], its line and nothing else; every other byte is the file's.
func WithoutWorkspace(data []byte, name string) ([]byte, error) {
	lines := linesOf(data)
	start, end, found := regionOf(lines, workspacesTable)
	if !found {
		return nil, ErrNotWritten
	}
	entry := regexp.MustCompile(`^\s*` + regexp.QuoteMeta(name) + `\s*=`)
	at := slices.IndexFunc(lines[start:end], entry.MatchString)
	if at < 0 {
		return nil, ErrNotWritten
	}
	said := []byte(strings.Join(slices.Delete(slices.Clone(lines), start+at, start+at+1), ""))
	var after struct {
		Workspaces map[string]any `toml:"workspaces"`
	}
	if toml.Unmarshal(said, &after) != nil {
		return nil, ErrNotWritten
	}
	if _, still := after.Workspaces[name]; still {
		return nil, ErrNotWritten
	}
	return said, nil
}
