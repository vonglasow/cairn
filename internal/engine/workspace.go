package engine

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"cairn/internal/config"
)

const (
	storeFile         = "cairn.db"
	notesFolder       = "notes"
	byTheEnvironment  = "the environment"
	workspaceFlagName = "--workspace"
	storeFlagName     = "--store"
)

// ErrNoSuchWorkspace is returned when a workspace is named that the configuration does not name.
var ErrNoSuchWorkspace = errors.New("no workspace by that name")

// ErrWorkspaceNotThere is returned when the folder of the workspace asked for is not there.
var ErrWorkspaceNotThere = errors.New("the folder of this workspace is not there")

// ErrWorkspaceKnown is returned when a workspace is added under a name the configuration already gives one.
var ErrWorkspaceKnown = errors.New("a workspace by that name is already there")

// ErrNotAWorkspaceName is returned when a workspace is added under a name that is not lowercase letters, digits and dashes, or is the one name kept for the default.
var ErrNotAWorkspaceName = errors.New("a workspace is named in lowercase letters, digits and dashes, and is not called default")

var plainName = regexp.MustCompile(`^[a-z0-9-]+$`)

// Workspace is one workspace the configuration names: its name, its folder, and whether it is the default.
type Workspace struct {
	Name    string `json:"name"`
	Folder  string `json:"folder"`
	Default bool   `json:"default,omitempty"`
}

// Workspaces is how workspace list answers: every workspace the configuration names, by name.
type Workspaces struct {
	Workspaces []Workspace `json:"workspaces"`
}

// WorkspaceAdded is how workspace add answers: the workspace written, its two places, and the file it was written in.
type WorkspaceAdded struct {
	Name    string `json:"name"`
	Folder  string `json:"folder"`
	Store   string `json:"store"`
	Notes   string `json:"notes"`
	Written string `json:"written"`
}

// WorkspaceUsed is how workspace use answers: the default now, and the one it was, if any.
type WorkspaceUsed struct {
	Default string `json:"default"`
	Was     string `json:"was,omitempty"`
}

// WorkspaceShown is how workspace show answers: which workspace answers, if one does, why, the places this command uses, and how to name them as a workspace when none does.
type WorkspaceShown struct {
	Workspace string `json:"workspace,omitempty"`
	Folder    string `json:"folder,omitempty"`
	Because   string `json:"because"`
	Store     string `json:"store"`
	Notes     string `json:"notes"`
	Time      string `json:"time" jsonschema:"on, recorded and shown; hidden, recorded and not shown; off, neither"`
	TimeFrom  string `json:"time_from" jsonschema:"where the time is said: set in the store's settings, from the global configuration, or by default"`
	Add       string `json:"add,omitempty"`
	Switch    string `json:"switch,omitempty" jsonschema:"said by a server, which keeps the store it started on: switching is not done through it"`
}

// Started is where a server answers from, settled when it started and kept whatever the configuration names afterwards.
type Started struct {
	Workspace string      `json:"workspace,omitempty"`
	Folder    string      `json:"folder,omitempty"`
	Because   string      `json:"because"`
	Store     Emplacement `json:"store"`
	Notes     Emplacement `json:"notes"`
}

const switchingIsNotHere = "not from here: this server keeps its store; another workspace is another server, which the person starts"

type answering struct {
	store, notes Emplacement
	workspace    string
	folder       string
	because      string
	then         string
}

// Keep settles where h answers from, once, so that a server started on it answers from there whatever the configuration names afterwards; what cannot be settled yet is left to be found at each call.
func Keep(h Home) Home {
	if _, err := Locate(h); err != nil {
		return h
	}
	_, a, _ := h.settled()
	h.Kept = &Started{Workspace: a.workspace, Folder: a.folder, Because: a.then, Store: a.store, Notes: a.notes}
	return h
}

func (h Home) file() (string, error) {
	if h.Dir == "" {
		return "", fmt.Errorf("%w: %w", ErrStore, h.Missing)
	}
	return config.FileIn(h.Dir), nil
}

func workspacesIn(p config.Places) []Workspace {
	var listed []Workspace
	for _, one := range p.Workspaces() {
		listed = append(listed, Workspace{Name: one.Name, Folder: one.Folder, Default: one.Name == p.DefaultWorkspace()})
	}
	return listed
}

func workspaceNamed(p config.Places, name string) (config.Workspace, error) {
	var names []string
	for _, one := range p.Workspaces() {
		if one.Name == name {
			return one, nil
		}
		names = append(names, one.Name)
	}
	there := "none is named"
	if len(names) > 0 {
		there = "the workspaces are " + strings.Join(names, ", ")
	}
	return config.Workspace{}, fmt.Errorf("%q: %w; %s", name, ErrNoSuchWorkspace, there)
}

func workspaceThere(p config.Places, name string) (config.Workspace, error) {
	one, err := workspaceNamed(p, name)
	if err != nil {
		return one, err
	}
	if _, err := os.Stat(one.Folder); err != nil {
		return one, fmt.Errorf("%s: %w", one.Folder, ErrWorkspaceNotThere)
	}
	return one, nil
}

func inWorkspace(one config.Workspace, by string) (Emplacement, Emplacement) {
	return Emplacement{At: filepath.Join(one.Folder, storeFile), By: by}, Emplacement{At: filepath.Join(one.Folder, notesFolder), By: by}
}

func (h Home) answering(p config.Places) (answering, error) {
	if k := h.Kept; k != nil {
		return answering{store: k.Store, notes: k.Notes, workspace: k.Workspace, folder: k.Folder, because: k.Because, then: k.Because}, nil
	}
	if h.Workspace != "" && h.Store != "" {
		return answering{}, Misread(fmt.Errorf("%s and %s both name the store: give one or the other", workspaceFlagName, storeFlagName))
	}
	gate := &placesGate{}
	p.Tell(gate)
	a := answering{store: gate.store, notes: gate.notes, because: "nothing names it: it is where cairn keeps a store by default", then: "nothing named it when this server started: it is where cairn keeps a store by default"}
	switch {
	case h.Workspace != "":
		one, err := workspaceThere(p, h.Workspace)
		if err != nil {
			return a, err
		}
		a.store, a.notes = inWorkspace(one, workspaceFlagName)
		a.workspace, a.because, a.then = one.Name, workspaceFlagName+" names it", workspaceFlagName+", given when this server started"
		return a.named(p), nil
	case h.Store != "":
		at, err := config.Absolute(h.Store, h.Dir)
		if err != nil {
			return a, fmt.Errorf("%s %w", storeFlagName, err)
		}
		a.store, a.because, a.then = Emplacement{At: at, By: storeFlagName}, storeFlagName+" names it", storeFlagName+", given when this server started"
		return a.ownNotes(p), nil
	case a.store.By == byTheEnvironment:
		a.because, a.then = config.StoreVariable+" names it", config.StoreVariable+", set when this server started"
		return a.ownNotes(p), nil
	case a.store.By == "the file":
		a.because, a.then = "the file names it", "the file named it when this server started"
	}
	if name := p.DefaultWorkspace(); name != "" {
		one, err := workspaceThere(p, name)
		if err != nil {
			return a, err
		}
		store, notes := inWorkspace(one, fmt.Sprintf("the default workspace, %q", name))
		a.store, a.workspace, a.because, a.then = store, one.Name, "it is the default of [workspaces]", "it was the default of [workspaces] when this server started"
		if a.notes.By != byTheEnvironment {
			a.notes = notes
		}
	}
	return a.named(p), nil
}

func (a answering) ownNotes(p config.Places) answering {
	a = a.named(p)
	if a.workspace != "" && a.notes.By != byTheEnvironment {
		_, a.notes = inWorkspace(config.Workspace{Name: a.workspace, Folder: a.folder}, fmt.Sprintf("the workspace %q", a.workspace))
	}
	return a
}

func (a answering) named(p config.Places) answering {
	for _, one := range p.Workspaces() {
		if a.workspace == "" && filepath.Join(one.Folder, storeFile) == a.store.At {
			a.workspace = one.Name
		}
		if one.Name == a.workspace {
			a.folder = one.Folder
		}
	}
	return a
}

func (a answering) shown() WorkspaceShown {
	shown := WorkspaceShown{Workspace: a.workspace, Folder: a.folder, Because: a.because, Store: a.store.At, Notes: a.notes.At}
	if a.workspace == "" {
		shown.Add = "cairn workspace add <name> <folder>"
		if filepath.Base(a.store.At) == storeFile {
			shown.Add = "cairn workspace add <name> " + filepath.Dir(a.store.At)
		}
	}
	return shown
}

func withConfiguration(h Home, change func([]byte, config.Places) ([]byte, error)) error {
	path, err := h.file()
	if err != nil {
		return err
	}
	read, err := h.configuration()
	if err != nil {
		return err
	}
	was, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	said, err := change(was, read)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, said, 0o600)
}

func settingsSayingNothing(folder string) error {
	file, err := os.OpenFile(filepath.Join(folder, config.SettingsFile), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, fs.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err := file.WriteString(config.SettingsHeader); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

// WorkspaceList lists every workspace the configuration names, the default marked.
var WorkspaceList = Operation{
	Thing:   "workspace",
	Verb:    "list",
	Short:   "List the workspaces, the default marked",
	Does:    "List every workspace the configuration names: its name and the folder holding its store and its notes, the default marked.",
	Example: "cairn workspace list",
	Level:   Read,
	Needs:   AConfiguration,
	Answers: Workspaces{},
	Run: func(on On, _ Input) (Answer, error) {
		read, err := on.Home.configuration()
		if err != nil {
			return Answer{}, err
		}
		return Answer{Workspaces: &Workspaces{Workspaces: workspacesIn(read)}}, nil
	},
}

// WorkspaceAdd names a workspace in the configuration, and makes its folder.
var WorkspaceAdd = Operation{
	Thing:   "workspace",
	Verb:    "add",
	Short:   "Name a workspace, a folder holding a store and its notes",
	Does:    "Name a workspace in the configuration file, writing the file when there is none and leaving the rest of it as it was, and make its folder with an empty notes folder in it and settings.toml saying nothing yet. Its store is cairn.db in the folder, its notes the notes folder; cairn workspace use makes it the default.",
	Example: "cairn workspace add work ~/cairn/work",
	Level:   Write,
	Needs:   AConfiguration,
	Again:   ChangesNothingMore,
	Answers: WorkspaceAdded{},
	Takes: []Argument{
		{Name: "name", Kind: Text, Positional: true, Required: true, Means: "the workspace's name: lowercase letters, digits and dashes"},
		{Name: "folder", Kind: Text, Positional: true, Required: true, Means: "the folder holding its store and its notes: absolute, or under ~"},
	},
	Run: func(on On, in Input) (Answer, error) {
		name, folder := in.Given["name"], in.Given["folder"]
		if !plainName.MatchString(name) || name == config.DefaultWorkspaceKey {
			return Answer{}, fmt.Errorf("%q: %w", name, ErrNotAWorkspaceName)
		}
		at, err := config.Absolute(folder, on.Home.Dir)
		if err != nil {
			return Answer{}, err
		}
		err = withConfiguration(on.Home, func(was []byte, read config.Places) ([]byte, error) {
			if _, err := workspaceNamed(read, name); err == nil {
				return nil, fmt.Errorf("%q: %w", name, ErrWorkspaceKnown)
			}
			said, err := config.WithWorkspace(was, name, folder)
			if err != nil {
				return nil, err
			}
			if err := os.MkdirAll(filepath.Join(at, notesFolder), 0o700); err != nil {
				return nil, err
			}
			return said, settingsSayingNothing(at)
		})
		if err != nil {
			return Answer{}, err
		}
		store, notes := inWorkspace(config.Workspace{Name: name, Folder: at}, "")
		path, _ := on.Home.file()
		return Answer{WorkspaceAdded: &WorkspaceAdded{Name: name, Folder: at, Store: store.At, Notes: notes.At, Written: path}}, nil
	},
}

// WorkspaceUse makes a workspace the default.
var WorkspaceUse = Operation{
	Thing:   "workspace",
	Verb:    "use",
	Short:   "Make a workspace the default",
	Does:    "Make a workspace the default, the one used when neither a flag nor the environment names another, and say which one it was.",
	Example: "cairn workspace use home",
	Level:   Write,
	Needs:   AConfiguration,
	Again:   ChangesNothingMore,
	Answers: WorkspaceUsed{},
	Takes: []Argument{
		{Name: "name", Kind: Text, Positional: true, Required: true, Means: "the workspace's name"},
	},
	Run: func(on On, in Input) (Answer, error) {
		used := WorkspaceUsed{Default: in.Given["name"]}
		err := withConfiguration(on.Home, func(was []byte, read config.Places) ([]byte, error) {
			if _, err := workspaceThere(read, used.Default); err != nil {
				return nil, err
			}
			used.Was = read.DefaultWorkspace()
			return config.WithDefault(was, used.Default)
		})
		if err != nil {
			return Answer{}, err
		}
		return Answer{WorkspaceUsed: &used}, nil
	},
}

// WorkspaceShow says which workspace answers, why, and where its store and notes are.
var WorkspaceShow = Operation{
	Thing:   "workspace",
	Verb:    "show",
	Short:   "Say which workspace answers, and why",
	Does:    "Say which workspace answers this command and why — a flag, the environment, the default — and the store and notes it uses; a store no workspace names is said, with how to name it.",
	Example: "cairn workspace show\ncairn workspace show --workspace home",
	Level:   Read,
	Needs:   AConfiguration,
	Answers: WorkspaceShown{},
	Run: func(on On, _ Input) (Answer, error) {
		read, answers, err := on.Home.settled()
		if err != nil {
			return Answer{}, err
		}
		shown := answers.shown()
		shown.Time, shown.TimeFrom = read.TimeMode(), whereSaid(read, "pages.time")
		if on.Home.Kept != nil {
			shown.Switch = switchingIsNotHere
		}
		return Answer{WorkspaceShown: &shown}, nil
	},
}

func workspacesText(listed []Workspace) string {
	if len(listed) == 0 {
		return "no workspace is named\ncairn workspace add <name> <folder> names one\n"
	}
	wide := 0
	for _, one := range listed {
		wide = max(wide, len(one.Name))
	}
	var said strings.Builder
	for _, one := range listed {
		line := fmt.Sprintf("%-*s  %s", wide, one.Name, one.Folder)
		if one.Default {
			line += "  default"
		}
		said.WriteString(line + "\n")
	}
	return said.String()
}

func workspaceAddedText(added WorkspaceAdded) string {
	return fmt.Sprintf("workspace %s is %s, written in %s\nits store is %s, its notes %s\ncairn workspace use %s makes it the default\n",
		added.Name, added.Folder, added.Written, added.Store, added.Notes, added.Name)
}

func workspaceUsedText(used WorkspaceUsed) string {
	was := "there was none"
	if used.Was != "" {
		was = fmt.Sprintf("it was %q", used.Was)
	}
	return fmt.Sprintf("the default workspace is now %q; %s\n", used.Default, was)
}

func workspaceShownText(shown WorkspaceShown) string {
	var said strings.Builder
	if shown.Workspace == "" {
		said.WriteString("workspace  none: no workspace names this store\n")
	} else {
		fmt.Fprintf(&said, "workspace  %s\nfolder     %s\n", shown.Workspace, shown.Folder)
	}
	fmt.Fprintf(&said, "because    %s\nstore      %s\nnotes      %s\n", shown.Because, shown.Store, shown.Notes)
	if shown.Time != "" {
		fmt.Fprintf(&said, "time       %s, %s\n", shown.Time, shown.TimeFrom)
	}
	if shown.Switch != "" {
		fmt.Fprintf(&said, "switch     %s\n", shown.Switch)
	}
	if shown.Add != "" {
		fmt.Fprintf(&said, "\nto name it: %s\n", shown.Add)
	}
	return said.String()
}

func workspacesTable(listed []Workspace) string {
	if len(listed) == 0 {
		return "\n[workspaces]\n# none is named: cairn workspace add names one\n"
	}
	var said strings.Builder
	said.WriteString("\n[workspaces]\n")
	for _, one := range listed {
		if one.Default {
			fmt.Fprintf(&said, "default = %q\n", one.Name)
		}
	}
	for _, one := range listed {
		fmt.Fprintf(&said, "%s = %q\n", one.Name, one.Folder)
	}
	return said.String()
}
