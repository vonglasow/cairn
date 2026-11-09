package gui

import (
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strconv"

	"cairn/internal/engine"
)

// ConfigurationFoot is what the panel of the configuration says last: what the window reads of the file, it reads when it opens.
const ConfigurationFoot = "Changes take effect when cairn is opened again."

// OpenSettings is the offer handing the settings of the workspace shown to the system, to open in its editor.
const OpenSettings = "Open settings.toml"

// SetHere and FromGlobal say where a key of the settings of a workspace comes from: its own settings, or the global configuration under them.
const (
	SetHere    = "set here"
	FromGlobal = "from the global configuration"
)

type configurationState struct {
	configured engine.Config
	configHere bool
	opens      string
}

func (w Window) openingConfiguration() Window {
	got, err := w.work.Ask(engine.ConfigShow, engine.Input{})
	if err != nil {
		return w.failing(err)
	}
	w.configured = *got.Config
	_, missing := os.Stat(w.configured.File)
	w.configHere = missing == nil
	w.screen = OnConfiguration
	return w
}

func (w Window) writingConfiguration() Window {
	if _, err := w.work.Ask(engine.ConfigInit, engine.Input{}); err != nil {
		return w.failing(err)
	}
	return w.openingConfiguration()
}

func (w Window) onConfiguration(string) Window { return w }

func emplacedFrom(at engine.Emplacement) string {
	switch at.By {
	case "nothing":
		return "default"
	case "the file":
		return "file"
	}
	return at.By
}

func notSaid(said string) string {
	if said == "" {
		return "not said"
	}
	return said
}

// ConfigLine is one key of the configuration as its panel lays it out: the key, the value in force, and where the value came from — the file, the environment, or the default — each in its column.
type ConfigLine struct {
	Key, Value, From       string
	KeyAt, ValueAt, FromAt Box
}

func (w Window) configKeys() []ConfigLine {
	c := w.configured
	if c.Problem != "" {
		return []ConfigLine{{Key: "problem", Value: c.Problem}}
	}
	written := func(key, value string) ConfigLine {
		from := "default"
		switch {
		case slices.Contains(c.WrittenHere, key):
			from = SetHere
		case slices.Contains(c.Written, key):
			from = FromGlobal
		}
		return ConfigLine{Key: key, Value: value, From: from}
	}
	lines := []ConfigLine{
		{Key: "store", Value: c.Store.At, From: emplacedFrom(c.Store)},
		{Key: "notes", Value: c.Notes.At, From: emplacedFrom(c.Notes)},
		{Key: "plugin", Value: c.Plugin.At, From: emplacedFrom(c.Plugin)},
	}
	if c.Pages == nil || c.Archive == nil || c.Focus == nil || c.Time == nil || c.Window == nil {
		return lines
	}
	return append(lines,
		written("pages.tasks", strconv.FormatBool(c.Pages.Tasks)),
		written("pages.objectives", strconv.FormatBool(c.Pages.Objectives)),
		written("pages.time", c.Pages.Time),
		written("archive.start", notSaid(c.Archive.Start)),
		written("archive.days", strconv.Itoa(c.Archive.Days)),
		written("focus.session", notSaid(c.Focus.Session)),
		written("time.deleted", c.Time.Deleted),
		written("time.weekends", c.Time.Weekends),
		written("window.keep_sizes", strconv.FormatBool(c.Window.KeepSizes)),
	)
}

// ConfigLines is every key of the configuration where its panel lays it out: the keys down one column, their values down the next, where each came from at the right.
func (w Window) ConfigLines() []ConfigLine {
	box := w.ConfigurationBox()
	lines := w.configKeys()
	head := len(w.ConfigHead())
	keys, froms := 0, 0
	for _, line := range lines {
		keys, froms = max(keys, monoWide(line.Key, small)), max(froms, monoWide(line.From, small))
	}
	for i := range lines {
		row := Inside(ListRowAt(box, head+i))
		lines[i].KeyAt = Box{X: row.X, Y: row.Y, W: keys, H: row.H}
		lines[i].FromAt = Box{X: row.X + row.W - froms, Y: row.Y, W: froms, H: row.H}
		from := row.X + keys + configGap + copyRoom
		lines[i].ValueAt = Box{X: from, Y: row.Y, W: lines[i].FromAt.X - configGap - from, H: row.H}
	}
	return lines
}

// ConfigHead is what the settings of a workspace say above their keys: the settings read, over the global file; nothing for a store no workspace names.
func (w Window) ConfigHead() []string {
	if w.configured.Settings == "" || w.switcher.place.Workspace == "" {
		return nil
	}
	return []string{w.configured.Settings, "over " + w.configured.File}
}

// TimeSegment is one of the three ways the time is kept, as the settings draw them: on, hidden, off, the one in force filled.
type TimeSegment struct {
	Said string
	On   bool
	At   Box
}

// TimeSegments is the time as the panel draws it, three segments over the value of pages.time, and whether the panel shows that key.
func (w Window) TimeSegments() ([]TimeSegment, bool) {
	for _, line := range w.ConfigLines() {
		if line.Key != "pages.time" {
			continue
		}
		modes := []string{"on", "hidden", "off"}
		wide := min(segmentWide, line.ValueAt.W/len(modes))
		segments := make([]TimeSegment, len(modes))
		for i, mode := range modes {
			segments[i] = TimeSegment{Said: mode, On: mode == line.Value, At: Box{X: line.ValueAt.X + i*wide, Y: line.ValueAt.Y, W: wide, H: line.ValueAt.H}}
		}
		return segments, true
	}
	return nil, false
}

const segmentWide = 64

// ConfigFoot is what the panel says under the keys: the file it reads, and when what is changed in it is read.
func (w Window) ConfigFoot() []string {
	file := w.configured.File
	if !w.configHere {
		file += "  (not written yet)"
	}
	return []string{file, ConfigurationFoot}
}

// ConfigurationBox is the panel of the configuration: across the middle of the window, a line a key and two for its foot.
func (w Window) ConfigurationBox() Box {
	lines := w.configKeys()
	said := append([]string{}, w.ConfigHead()...)
	for _, line := range lines {
		said = append(said, line.Key+"    "+line.Value+"        "+line.From)
	}
	return w.listBox(append(said, w.ConfigFoot()...))
}

const configGap = 24

// OpenAsked is the file the window wants the system to open, if any.
func (w Window) OpenAsked() (string, bool) { return w.opens, w.opens != "" }

// OpenTaken says the loop has handed the file asked for to the system.
func (w Window) OpenTaken() Window {
	w.opens = ""
	return w
}

// OpenCommand is the command that hands a file to the system, to open in the editor it keeps for that kind of file.
func OpenCommand(goos, path string) (string, []string) {
	switch goos {
	case "darwin":
		return "open", []string{"-t", path}
	case "windows":
		return "cmd", []string{"/c", "start", "", path}
	}
	return "xdg-open", []string{path}
}

// Opened takes what came of handing the file to the system: a command that could not start is said.
func (w Window) Opened(err error) Window {
	if err != nil {
		return w.failing(err)
	}
	return w
}

// HandOver starts the command that has the system open a file, and does not wait for it.
func HandOver(path string) error {
	name, args := OpenCommand(runtime.GOOS, path)
	command := exec.Command(name, args...)
	if err := command.Start(); err != nil {
		return err
	}
	go func() { _ = command.Wait() }()
	return nil
}
