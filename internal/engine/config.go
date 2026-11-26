package engine

import (
	"cairn/internal/engine/app"
	"cairn/internal/engine/workspace"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"cairn/internal/store/files"
)

type readingGate struct {
	home   Home
	config workspace.Config
}

func (g *readingGate) ReadFrom(path string) {
	g.config.File = path
}

func (g *readingGate) Resolved(p files.Places) {
	places, err := g.home.answering(p)
	if err != nil {
		g.config.Problem = err.Error()
		return
	}
	global := workspace.ValuesOf(p)
	p, err = p.WithSettings(files.SettingsBeside(places.Store.At))
	if err != nil {
		g.config.Problem = err.Error()
		return
	}
	for _, key := range p.WrittenHere() {
		if value, ok := global[key]; ok {
			if g.config.Global == nil {
				g.config.Global = map[string]string{}
			}
			g.config.Global[key] = value
		}
	}
	own, _ := files.NameIn(files.SettingsBeside(places.Store.At))
	if colour := workspace.ColourOf(places.Workspace, places.Store.At); colour != "" || own != "" {
		g.config.Workspace = &workspace.Colouring{Colour: colour, Name: own}
	}
	g.config.Store = places.Store
	g.config.Notes = places.Notes
	if move, left, _ := notesLeft(places); left && places.Folder != "" {
		g.config.NotesLeft = &move
	}
	g.config.Workspaces = workspace.WorkspacesIn(p)
	g.config.Plugin = workspace.Emplacement{At: app.PluginBeside(places.Store.At), By: "nothing"}
	if told := p.Plugin(); told != "" {
		g.config.Plugin = workspace.Emplacement{At: told, By: "the file"}
	}
	pages := p.Pages()
	g.config.Pages = &workspace.PagesConfig{Tasks: pages[0], Objectives: pages[1], Time: p.TimeMode(), Agents: pages[3]}
	start, days, _ := p.Archive()
	g.config.Archive = &workspace.Archive{Start: start, Days: days}
	g.config.Time = &workspace.TimeConfig{Deleted: p.TimeDeleted(), Weekends: p.TimeWeekends()}
	g.config.Focus = &workspace.FocusConfig{Session: p.FocusSession()}
	g.config.Approach = &workspace.ApproachConfig{Label: p.ApproachLabel()}
	g.config.Recurrence = &workspace.RecurrenceConfig{On: p.Recurrence()}
	said := p.Agents()
	g.config.Agents = &workspace.AgentsConfig{Notify: said.Notify, Count: said.Count, Permissions: said.Permissions, Silent: said.Silent}
	g.themes(p)
	g.config.Written = p.Written()
	g.config.Settings = p.Settings()
	g.config.WrittenHere = p.WrittenHere()
	g.config.Ignored = p.Ignored()
}

func (g *readingGate) NotUnderstood(_ string, err error) {
	g.config.Problem = err.Error()
}

// ConfigShow shows which places cairn will use, and where it got them.
var ConfigShow = Operation{
	Thing:   "config",
	Verb:    "show",
	Short:   "Show the whole configuration cairn uses, and where it got its places",
	Does:    "Show the whole configuration cairn uses, written as the file is: every key under its table with the value in force — what the file says, or what is used when it says nothing — the file read, and what named each place: the file, the environment, or nothing.",
	Example: "cairn config show",
	Level:   Read,
	Needs:   aConfiguration,
	Answers: workspace.Config{},
	Run: func(on workedOn, _ Input) (Answer, error) {
		reading, err := files.NewReading(on.Home.Dir, on.Home.lookup())
		if err != nil {
			return Answer{}, err
		}
		gate := &readingGate{home: on.Home}
		reading.Tell(gate)
		return Answer{Config: &gate.config}, nil
	},
}

// ConfigInit writes the whole configuration file, every key set to what is used when nothing is said.
var ConfigInit = Operation{
	Thing:   "config",
	Verb:    "init",
	Short:   "Write the configuration file, every key in it",
	Does:    "Write the configuration file where cairn reads it, every key commented and set to what cairn uses when nothing is said, to change from there, with one workspace, main, on ~/.cairn, the default, its folder made. A configuration already there is never written over.",
	Example: "cairn config init",
	Level:   write,
	Needs:   aConfiguration,
	Again:   ChangesNothingMore,
	Answers: workspace.ConfigWritten{},
	Run: func(on workedOn, _ Input) (Answer, error) {
		path, err := on.Home.file()
		if err != nil {
			return Answer{}, err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return Answer{}, err
		}
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if errors.Is(err, fs.ErrExist) {
			return Answer{}, fmt.Errorf("%s: %w", path, workspace.ErrConfigAlreadyThere)
		}
		if err != nil {
			return Answer{}, err
		}
		if _, err := file.WriteString(files.Example); err != nil {
			_ = file.Close()
			return Answer{}, err
		}
		if err := file.Close(); err != nil {
			return Answer{}, err
		}
		read, err := on.Home.configuration()
		if err != nil {
			return Answer{}, err
		}
		if one, err := workspace.Named(read, read.DefaultWorkspace()); err == nil {
			if err := workspace.MakeFolder(one); err != nil {
				return Answer{}, err
			}
		}
		return Answer{Written: &path}, nil
	},
}
