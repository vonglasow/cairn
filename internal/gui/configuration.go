package gui

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"cairn/internal/engine"
)

// ConfigurationFoot is what the panel of the configuration says last: what the window reads of the file, it reads when it opens.
const ConfigurationFoot = "Changes take effect when cairn is opened again."

const configurationKeyed = 18

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

func namedByWords(by string) string {
	if by == "nothing" {
		return "by default"
	}
	return by
}

func emplaced(at engine.Emplacement) string { return at.At + "  (" + namedByWords(at.By) + ")" }

func notSaid(said string) string {
	if said == "" {
		return "not said"
	}
	return said
}

// Configured is what the panel of the configuration says, line by line: the file read, then each key with its value and what named it.
func (w Window) Configured() []Said {
	c := w.configured
	file := c.File
	if !w.configHere {
		file += "  (not written yet)"
	}
	lines := []Said{{Key: "file", Means: file}}
	if c.Problem != "" {
		return append(lines, Said{Key: "problem", Means: c.Problem}, Said{Means: ConfigurationFoot})
	}
	return append(lines,
		Said{Key: "store", Means: emplaced(c.Store)},
		Said{Key: "notes", Means: emplaced(c.Notes)},
		Said{Key: "plugin", Means: emplaced(c.Plugin)},
		Said{Key: "pages.tasks", Means: strconv.FormatBool(c.Pages.Tasks)},
		Said{Key: "pages.objectives", Means: strconv.FormatBool(c.Pages.Objectives)},
		Said{Key: "pages.time", Means: strconv.FormatBool(c.Pages.Time)},
		Said{Key: "archive.start", Means: notSaid(c.Archive.Start)},
		Said{Key: "archive.days", Means: strconv.Itoa(c.Archive.Days)},
		Said{Key: "focus.session", Means: notSaid(c.Focus.Session)},
		Said{Key: "time.deleted", Means: c.Time.Deleted},
		Said{Key: "time.weekends", Means: c.Time.Weekends},
		Said{Key: "window.keep_sizes", Means: strconv.FormatBool(c.Window.KeepSizes)},
		Said{Means: ConfigurationFoot},
	)
}

// ConfigurationSays is one line of the panel: the key, and what it holds.
func ConfigurationSays(line Said) string {
	if line.Key == "" {
		return line.Means
	}
	return line.Key + strings.Repeat(" ", max(1, configurationKeyed-len(line.Key))) + line.Means
}

// ConfigurationBox is the panel of the configuration: across the middle of the window, a line a key.
func (w Window) ConfigurationBox() Box {
	lines := w.Configured()
	said := make([]string, 0, len(lines))
	for _, line := range lines {
		said = append(said, ConfigurationSays(line))
	}
	return w.listBox(said)
}

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
		return "open", []string{path}
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
