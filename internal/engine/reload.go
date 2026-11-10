package engine

import (
	"crypto/sha256"
	"os"
	"strconv"
	"strings"
	"time"

	"cairn/internal/config"
	"cairn/internal/theme"
)

// PagesOf is the pages an answer of config show says are shown, how many days a week shows, and whether tasks come back: what a board that read the configuration again draws from.
func PagesOf(c Config) Pages {
	var p Pages
	if c.Pages != nil {
		for _, page := range []struct {
			name string
			on   bool
		}{{PageTasks, c.Pages.Tasks}, {PageObjectives, c.Pages.Objectives}, {PageTime, c.Pages.Time == config.TimeOn}} {
			if !page.on {
				p.off = append(p.off, page.name)
			}
		}
	}
	p.noWeekend = c.Time != nil && c.Time.Weekends == config.WeekendsHidden
	p.noRecurrence = c.Recurrence == nil || !c.Recurrence.On
	return p
}

// ConfigChange is one key whose value in force moved between two answers of config show.
type ConfigChange struct {
	Key, Was, Now string
}

// String is the change as a board says it: the key, what it was, and what it is.
func (c ConfigChange) String() string { return c.Key + " " + c.Was + " → " + c.Now }

func valuesIn(c Config) map[string]string {
	values := map[string]string{}
	if c.Pages != nil {
		values["pages.tasks"], values["pages.objectives"], values["pages.time"] = strconv.FormatBool(c.Pages.Tasks), strconv.FormatBool(c.Pages.Objectives), c.Pages.Time
	}
	if c.Archive != nil {
		values["archive.start"], values["archive.days"] = notSaid(c.Archive.Start), strconv.Itoa(c.Archive.Days)
	}
	if c.Focus != nil {
		values["focus.session"] = notSaid(c.Focus.Session)
	}
	if c.Approach != nil {
		values["approach.label"] = notSaid(c.Approach.Label)
	}
	if c.Time != nil {
		values["time.deleted"], values["time.weekends"] = c.Time.Deleted, c.Time.Weekends
	}
	if c.Window != nil {
		values["window.keep_sizes"], values["window.theme"], values["window.intro"] = strconv.FormatBool(c.Window.KeepSizes), c.Window.Theme, c.Window.Intro
	}
	if c.Recurrence != nil {
		values["recurrence.on"] = strconv.FormatBool(c.Recurrence.On)
	}
	if c.Workspace != nil {
		values["workspace.colour"] = c.Workspace.Colour
	}
	values["plugin"] = c.Plugin.At
	return values
}

func notSaid(said string) string {
	if said == "" {
		return NotSaid
	}
	return said
}

// ValueIn is the value in force an answer of config show gives the key named, as a board writes it; NotSaid for a key with no value.
func ValueIn(c Config, key string) string { return valuesIn(c)[key] }

// ConfigChanges is every key whose value in force differs from one answer of config show to the next, in the order the catalogue lists the keys; nothing when none does.
func ConfigChanges(was, now Config) []ConfigChange {
	before, after := valuesIn(was), valuesIn(now)
	var changed []ConfigChange
	for _, key := range config.Keys() {
		if before[key.Name] != after[key.Name] {
			changed = append(changed, ConfigChange{Key: key.Name, Was: before[key.Name], Now: after[key.Name]})
		}
	}
	return changed
}

// SettingsOf is where the settings of the store an answer of config show was given for are, whether a file is there or not; nothing when it names no store.
func SettingsOf(c Config) string {
	if c.Store.At == "" {
		return ""
	}
	return config.SettingsBeside(c.Store.At)
}

type seenFile struct {
	path  string
	size  int64
	at    time.Time
	there bool
	sum   [sha256.Size]byte
}

func statOf(path string) seenFile {
	f := seenFile{path: path}
	if info, err := os.Stat(path); err == nil {
		f.size, f.at, f.there = info.Size(), info.ModTime(), true
	}
	return f
}

func (f seenFile) summed() seenFile {
	data, _ := os.ReadFile(f.path)
	f.sum = sha256.Sum256(data)
	return f
}

// ConfigStamp is what a board knows of the files its configuration was read from — the global file, the store's settings, and a scheme of one's own worn — to tell, at its beat, whether any of them says something else now.
type ConfigStamp struct {
	files []seenFile
}

// StampOf is the stamp of the files an answer of config show was read from, as they are now.
func StampOf(c Config) ConfigStamp {
	paths := []string{c.File, SettingsOf(c)}
	for _, one := range c.Themes {
		if one.InForce && one.From != theme.Carried {
			paths = append(paths, one.From)
		}
	}
	var s ConfigStamp
	for _, path := range paths {
		if path != "" {
			s.files = append(s.files, statOf(path).summed())
		}
	}
	return s
}

// Moved is the stamp as the files are now, and whether what one of them says changed: a size or a time that moved is read again, and a file saved with nothing changed does not move it.
func (s ConfigStamp) Moved() (ConfigStamp, bool) {
	now := ConfigStamp{files: make([]seenFile, len(s.files))}
	moved := false
	for i, was := range s.files {
		f := statOf(was.path)
		if f.size == was.size && f.at.Equal(was.at) && f.there == was.there {
			f.sum = was.sum
		} else {
			f = f.summed()
			moved = moved || f.sum != was.sum || f.there != was.there
		}
		now.files[i] = f
	}
	return now, moved
}

// Trouble is a configuration a board cannot read while it is open: the file, the line the key at fault is written on when it can be told, and why, in the reader's words.
type Trouble struct {
	File string
	Line int
	Said string
}

// String is the trouble as a board says it under the file's name: the line, then why.
func (t Trouble) String() string {
	if t.Line == 0 {
		return t.Said
	}
	return "line " + strconv.Itoa(t.Line) + ": " + t.Said
}

// TroubleIn is what an answer of config show that could not read its configuration says is wrong, told against the last answer that could: which file, which line, and why.
func TroubleIn(broken, last Config) Trouble {
	t := Trouble{File: broken.File, Said: broken.Problem}
	if settings := SettingsOf(last); settings != "" && strings.HasPrefix(broken.Problem, settings) {
		t.File = settings
	}
	t.Said = strings.TrimPrefix(strings.TrimPrefix(t.Said, t.File), ": ")
	data, err := os.ReadFile(t.File)
	if err != nil {
		return t
	}
	for _, key := range config.Keys() {
		if strings.Contains(t.Said, key.Name) {
			t.Line = config.LineOf(data, key.Name)
			return t
		}
	}
	return t
}
