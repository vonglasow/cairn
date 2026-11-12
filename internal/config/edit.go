package config

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// ErrNotEdited is returned when a key cannot be written in a file without breaking it: a shape the editor does not know, an inline table, a string over several lines, or a file that is not TOML.
var ErrNotEdited = errors.New("cairn cannot write this key in this file without breaking it; change it by hand")

// ErrNotSet is returned when a key is taken out of a file that does not say it.
var ErrNotSet = errors.New("is not said in")

var aTableHeading = regexp.MustCompile(`^\s*\[\s*([A-Za-z0-9_.-]+)\s*\]\s*(#.*)?$`)

func split(key string) (table, name string) {
	if table, name, ok := strings.Cut(key, "."); ok {
		return table, name
	}
	return "", key
}

func regionOf(lines []string, table string) (int, int, bool) {
	first := slices.IndexFunc(lines, aHeading.MatchString)
	if first < 0 {
		first = len(lines)
	}
	if table == "" {
		return 0, first, true
	}
	head := slices.IndexFunc(lines, func(l string) bool {
		m := aTableHeading.FindStringSubmatch(strings.TrimSuffix(l, "\n"))
		return m != nil && m[1] == table
	})
	if head < 0 {
		return 0, 0, false
	}
	end := len(lines)
	if next := slices.IndexFunc(lines[head+1:], aHeading.MatchString); next >= 0 {
		end = head + 1 + next
	}
	return head + 1, end, true
}

func keyLine(name string, commented bool) *regexp.Regexp {
	if commented {
		return regexp.MustCompile(`^\s*#\s*` + regexp.QuoteMeta(name) + `\s*=`)
	}
	return regexp.MustCompile(`^(\s*` + regexp.QuoteMeta(name) + `\s*=\s*)(.*)\n?$`)
}

func valueEnds(rest string) (int, bool) {
	switch {
	case strings.HasPrefix(rest, `"""`), strings.HasPrefix(rest, `'''`), strings.HasPrefix(rest, "{"), strings.HasPrefix(rest, "["):
		return 0, false
	case strings.HasPrefix(rest, `"`):
		for i := 1; i < len(rest); i++ {
			switch rest[i] {
			case '\\':
				i++
			case '"':
				return i + 1, true
			}
		}
		return 0, false
	case strings.HasPrefix(rest, "'"):
		end := strings.Index(rest[1:], "'")
		return end + 2, end >= 0
	}
	value, _, _ := strings.Cut(rest, "#")
	return len(strings.TrimRight(value, " \t\r\n")), true
}

func insertedAt(lines []string, start, end int) int {
	at := end
	if end < len(lines) {
		i := end
		for i > start && comment(lines[i-1]) {
			i--
		}
		if i > start && blank(lines[i-1]) {
			at = i
		}
	}
	for at > start && blank(lines[at-1]) {
		at--
	}
	return at
}

// WithKey is the file with the key given written to the value given, already as TOML writes it: its line changed in place, its comment kept; a key written commented taken out of its comment where it stands; or a line joining the end of its table, the table joining the end of the file when it is not there. Every other byte is the file's, and a file the change would not read back as meant is refused.
func WithKey(data []byte, key, literal string) ([]byte, error) {
	table, name := split(key)
	entry := name + " = " + literal + "\n"
	lines := linesOf(data)
	start, end, found := regionOf(lines, table)
	if !found {
		return sameButFor(data, []byte(strings.Join(withTable(lines, table), "")+entry), key, literal)
	}
	said, err := setIn(lines, start, end, name, literal, entry)
	if err != nil {
		return nil, err
	}
	return sameButFor(data, []byte(strings.Join(said, "")), key, literal)
}

// WithCommented is the file with a key it does not say yet, neither written nor commented, joining the end of its table as a comment carrying the value given and where it comes from, so uncommenting it sets it; a key already there leaves the file as it is.
func WithCommented(data []byte, key, literal, from string) []byte {
	table, name := split(key)
	entry := "# " + name + " = " + literal + "   # " + from + "\n"
	lines := linesOf(data)
	start, end, found := regionOf(lines, table)
	if !found {
		return []byte(strings.Join(withTable(lines, table), "") + entry)
	}
	for _, line := range lines[start:end] {
		if keyLine(name, false).MatchString(line) || keyLine(name, true).MatchString(line) {
			return data
		}
	}
	at := insertedAt(lines, start, end)
	return []byte(strings.Join(slices.Concat(lines[:at], []string{entry}, lines[at:]), ""))
}

func withTable(lines []string, table string) []string {
	heading := "[" + table + "]\n"
	if len(lines) == 0 {
		return []string{heading}
	}
	return append(slices.Clone(lines), "\n", heading)
}

func setIn(lines []string, start, end int, name, literal, entry string) ([]string, error) {
	said := slices.Clone(lines)
	line := keyLine(name, false)
	if at := slices.IndexFunc(lines[start:end], line.MatchString); at >= 0 {
		m := line.FindStringSubmatch(strings.TrimSuffix(lines[start+at], "\n"))
		ends, ok := valueEnds(m[2])
		if !ok {
			return nil, ErrNotEdited
		}
		said[start+at] = m[1] + literal + m[2][ends:] + "\n"
		return said, nil
	}
	if at := slices.IndexFunc(lines[start:end], keyLine(name, true).MatchString); at >= 0 {
		said[start+at] = entry
		return said, nil
	}
	at := insertedAt(lines, start, end)
	if at == 0 && end < len(lines) {
		entry += "\n"
	}
	return slices.Insert(said, at, entry), nil
}

// WithoutKey is the file with the line of the key given taken out, and nothing else of it: a comment above it stays, as the table it was under does.
func WithoutKey(data []byte, key string) ([]byte, error) {
	table, name := split(key)
	lines := linesOf(data)
	start, end, found := regionOf(lines, table)
	at := -1
	if found {
		at = slices.IndexFunc(lines[start:end], keyLine(name, false).MatchString)
	}
	if at < 0 {
		return nil, ErrNotSet
	}
	said := slices.Delete(slices.Clone(lines), start+at, start+at+1)
	return sameButFor(data, []byte(strings.Join(said, "")), key, "")
}

func sameButFor(was, said []byte, key, literal string) ([]byte, error) {
	var before, after map[string]any
	if toml.Unmarshal(was, &before) != nil || toml.Unmarshal(said, &after) != nil {
		return nil, ErrNotEdited
	}
	if before == nil {
		before = map[string]any{}
	}
	table, name := split(key)
	into := before
	if table != "" {
		inner, _ := before[table].(map[string]any)
		inner = cloned(inner)
		before[table] = inner
		into = inner
	}
	if literal == "" {
		delete(into, name)
	} else {
		var value map[string]any
		if toml.Unmarshal([]byte("v = "+literal), &value) != nil {
			return nil, ErrNotEdited
		}
		into[name] = value["v"]
	}
	if !reflect.DeepEqual(before, after) {
		return nil, ErrNotEdited
	}
	return said, nil
}

func cloned(table map[string]any) map[string]any {
	out := make(map[string]any, len(table))
	for k, v := range table {
		out[k] = v
	}
	return out
}

// Readable says why the file's content would be refused by the reader, nothing when it reads: the global configuration with its places under the home given, or a store's settings, which may not say what belongs to the machine.
func Readable(data []byte, home string, settings bool) error {
	f, _, err := decoded(data)
	if err != nil {
		return err
	}
	if settings {
		if key := machineKey(f); key != "" {
			return fmt.Errorf("%q %w", key, ErrMachineKey)
		}
		if key := emptyIn(f); key != "" {
			return fmt.Errorf("%s %w", key, ErrEmpty)
		}
	} else {
		f.Workspace = nil
		for _, said := range []*string{f.Store, f.Notes, f.Plugin} {
			if said == nil {
				continue
			}
			if _, err := Absolute(*said, home); err != nil {
				return err
			}
		}
		if _, _, err := workspacesOf(f.Workspaces, home); err != nil {
			return err
		}
	}
	var p Places
	return p.settle(f)
}

// LineOf is the line, counted from one, the key given is written on in the file; nought when it is not written there.
func LineOf(data []byte, key string) int {
	table, name := split(key)
	lines := linesOf(data)
	start, end, found := regionOf(lines, table)
	if !found {
		return 0
	}
	at := slices.IndexFunc(lines[start:end], keyLine(name, false).MatchString)
	if at < 0 {
		return 0
	}
	return start + at + 1
}
