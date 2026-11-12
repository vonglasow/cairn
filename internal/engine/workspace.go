package engine

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"cairn/internal/config"
	"cairn/internal/domain"
)

const (
	storeFile         = "cairn.db"
	notesFolder       = "notes"
	byTheEnvironment  = "the environment"
	workspaceFlagName = "--workspace"
	storeFlagName     = "--store"
	rungDefault       = "the default"
	rungFile          = "the file"
	rungNothing       = "where cairn keeps a store by default"
)

const (
	rungWorkspaceFlag = workspaceFlagName
	rungStoreFlag     = storeFlagName
	rungEnvironment   = config.StoreVariable
)

// WorkspaceVariable is the environment variable naming a workspace: under CAIRN_STORE and CAIRN_NOTES, which name places, and over the default workspace.
const WorkspaceVariable = "CAIRN_WORKSPACE"

var rungs = []string{rungWorkspaceFlag, rungEnvironment, WorkspaceVariable, rungDefault, rungFile, rungNothing}

// ErrNoSuchWorkspace is returned when a workspace is named that the configuration does not name.
var ErrNoSuchWorkspace = errors.New("no workspace by that name")

// ErrWorkspaceNotThere is returned when the folder of the workspace asked for is not there.
var ErrWorkspaceNotThere = errors.New("the folder of this workspace is not there")

// ErrWorkspaceKnown is returned when a workspace is added under a name the configuration already gives one.
var ErrWorkspaceKnown = errors.New("a workspace by that name is already there")

// ErrRemovingTheDefault is returned when the workspace asked to be removed is the default, which every command would then be refused for naming nobody.
var ErrRemovingTheDefault = errors.New("is the default: make another the default first, with cairn workspace use")

// ErrNotAWorkspaceName is returned when a workspace is added under a name that is not lowercase letters, digits and dashes, or is the one name kept for the default.
var ErrNotAWorkspaceName = errors.New("a workspace is named in lowercase letters, digits and dashes, and is not called default")

var plainName = regexp.MustCompile(`^[a-z0-9-]+$`)

// Workspace is one workspace the configuration names: its name, its folder, whether it is the default, the colour it wears, and whether its folder is gone.
type Workspace struct {
	Name     string `json:"name"`
	Folder   string `json:"folder"`
	Default  bool   `json:"default,omitempty"`
	Colour   string `json:"colour" jsonschema:"one of the label colours: the one its settings say, or one read off its name"`
	NotFound bool   `json:"not_found,omitempty"`
	Time     string `json:"time,omitempty" jsonschema:"hidden or off, as its own settings over the global configuration say; nothing when the time is on"`
}

// Rung is one of the places that may name the store, passed over, and what it named there: a workspace, or a path.
type Rung struct {
	Rung  string `json:"rung"`
	Names string `json:"names"`
}

// Why is why a store answers: the rung that named it, the rungs under it that named another and were passed over, those over it that said nothing, and whether it was settled when a server started.
type Why struct {
	Rung    string   `json:"rung" jsonschema:"--workspace, --store, CAIRN_STORE, CAIRN_WORKSPACE, the default (the default workspace), the file (its store key), or where cairn keeps a store by default"`
	Over    []Rung   `json:"over,omitempty" jsonschema:"the rungs under it that named a store too, and were passed over"`
	Unsaid  []string `json:"unsaid,omitempty" jsonschema:"the rungs over it that named nothing"`
	Started bool     `json:"started,omitempty" jsonschema:"settled when this server started, and kept since"`
}

// Workspaces is how workspace list answers: every workspace the configuration names, by name, and the stores a board opened by path that none names, the newest first.
type Workspaces struct {
	Workspaces []Workspace    `json:"workspaces"`
	Recent     []RecentFolder `json:"recent,omitempty"`
}

// WorkspaceRemoved is how workspace forget answers: the workspace the configuration no longer names, its folder, left as it was, and the file it was taken out of.
type WorkspaceRemoved struct {
	Name    string `json:"name"`
	Folder  string `json:"folder"`
	Written string `json:"written"`
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
	Workspace string   `json:"workspace,omitempty"`
	Folder    string   `json:"folder,omitempty"`
	Why       Why      `json:"why"`
	Store     string   `json:"store"`
	Notes     string   `json:"notes"`
	Own       string   `json:"own,omitempty" jsonschema:"the name the store's settings give it, shown where no workspace names it"`
	Colour    string   `json:"colour,omitempty" jsonschema:"one of the label colours: the one the store's settings say, or one read off the workspace's name; none for a store no workspace names and whose settings say none"`
	Time      string   `json:"time" jsonschema:"on, recorded and shown; hidden, recorded and not shown; off, neither"`
	TimeFrom  string   `json:"time_from" jsonschema:"where the time is said: set in the store's settings, from the global configuration, or by default"`
	Pages     []string `json:"pages" jsonschema:"the pages that are on: tasks, objectives, time"`
	PagesFrom string   `json:"pages_from" jsonschema:"where the pages are said: set in the store's settings, from the global configuration, or by default"`
	Add       string   `json:"add,omitempty"`
	Switch    string   `json:"switch,omitempty" jsonschema:"said by a server, which keeps the store it started on: switching is not done through it"`
}

// Started is where a server answers from, settled when it started and kept whatever the configuration names afterwards.
type Started struct {
	Workspace string      `json:"workspace,omitempty"`
	Folder    string      `json:"folder,omitempty"`
	Why       Why         `json:"why"`
	Store     Emplacement `json:"store"`
	Notes     Emplacement `json:"notes"`
}

const switchingIsNotHere = "not from here: this server keeps its store; another workspace is another server, which the person starts"

type answering struct {
	store, notes Emplacement
	workspace    string
	folder       string
	why          Why
}

// Keep settles where h answers from, once, so that a server started on it answers from there whatever the configuration names afterwards; what cannot be settled yet is left to be found at each call.
func Keep(h Home) Home {
	if _, err := Locate(h); err != nil {
		return h
	}
	_, a, _ := h.settled()
	a.why.Started = true
	h.Kept = &Started{Workspace: a.workspace, Folder: a.folder, Why: a.why, Store: a.store, Notes: a.notes}
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
		store := filepath.Join(one.Folder, storeFile)
		w := Workspace{Name: one.Name, Folder: one.Folder, Default: one.Name == p.DefaultWorkspace(), Colour: colourOf(one.Name, store)}
		w.NotFound = !w.Found()
		if own, err := p.WithSettings(config.SettingsBeside(store)); err == nil && own.TimeMode() != config.TimeOn {
			w.Time = own.TimeMode()
		}
		listed = append(listed, w)
	}
	return listed
}

func colourOf(name, store string) string {
	said, err := config.ColourIn(config.SettingsBeside(store))
	if colour, wrong := domain.ParseColour(said); err == nil && wrong == nil {
		return colour
	}
	if name == "" {
		return ""
	}
	return domain.ColourOf(name)
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
		return answering{store: k.Store, notes: k.Notes, workspace: k.Workspace, folder: k.Folder, why: k.Why}, nil
	}
	if h.Workspace != "" && h.Store != "" {
		return answering{}, Misread(fmt.Errorf("%s and %s both name the store: give one or the other", workspaceFlagName, storeFlagName))
	}
	gate := &placesGate{}
	p.Tell(gate)
	a := answering{store: gate.store, notes: gate.notes}
	if h.Workspace != "" {
		one, err := workspaceThere(p, h.Workspace)
		if err != nil {
			return a, err
		}
		a.store, a.notes = inWorkspace(one, workspaceFlagName)
		a.workspace, a.why = one.Name, h.why(p, 0)
		return a.named(p), nil
	}
	if h.Store != "" {
		at, err := config.Absolute(h.Store, h.Dir)
		if err != nil {
			return a, fmt.Errorf("%s %w", storeFlagName, err)
		}
		a.store, a.why = Emplacement{At: at, By: storeFlagName}, h.why(p, 0)
		a.why.Rung = rungStoreFlag
		return a.ownNotes(p), nil
	}
	if a.store.By == byTheEnvironment {
		a.why = h.why(p, 1)
		return a.ownNotes(p), nil
	}
	if h.environmentWorkspace() != "" {
		name := h.environmentWorkspace()
		one, err := workspaceThere(p, name)
		if err != nil {
			return a, fmt.Errorf("%s: %w", WorkspaceVariable, err)
		}
		a.why = h.why(p, 2)
		return a.inside(p, one, fmt.Sprintf("%s, %q", WorkspaceVariable, name)), nil
	}
	a.why = h.why(p, 5)
	if a.store.By == "the file" {
		a.why = h.why(p, 4)
	}
	if name := p.DefaultWorkspace(); name != "" {
		one, err := workspaceThere(p, name)
		if err != nil {
			return a, err
		}
		a.why = h.why(p, 3)
		return a.inside(p, one, fmt.Sprintf("the default workspace, %q", name)), nil
	}
	return a.named(p), nil
}

func (h Home) environmentWorkspace() string {
	said, _ := h.lookup()(WorkspaceVariable)
	return said
}

func (a answering) inside(p config.Places, one config.Workspace, by string) answering {
	store, notes := inWorkspace(one, by)
	a.store, a.workspace = store, one.Name
	if a.notes.By != byTheEnvironment {
		a.notes = notes
	}
	return a.named(p)
}

func (a answering) ownNotes(p config.Places) answering {
	a = a.named(p)
	if a.workspace != "" && a.notes.By != byTheEnvironment {
		_, a.notes = inWorkspace(config.Workspace{Name: a.workspace, Folder: a.folder}, fmt.Sprintf("the workspace %q", a.workspace))
	}
	return a
}

func (a answering) named(p config.Places) answering {
	if a.workspace == "" {
		a.workspace = nameOf(p, a.store.At)
	}
	for _, one := range p.Workspaces() {
		if one.Name == a.workspace {
			a.folder = one.Folder
		}
	}
	return a
}

func nameOf(p config.Places, store string) string {
	for _, one := range p.Workspaces() {
		if filepath.Join(one.Folder, storeFile) == store {
			return one.Name
		}
	}
	return ""
}

func (h Home) why(p config.Places, from int) Why {
	why := Why{Rung: rungs[from]}
	if from >= 3 {
		why.Unsaid = slices.DeleteFunc(slices.Clone(rungs[:from]), func(rung string) bool { return rung == WorkspaceVariable })
	}
	if said, ok := h.lookup()(config.StoreVariable); from < 1 && ok && said != "" {
		at, err := config.Absolute(said, h.Dir)
		if err != nil {
			at = said
		}
		named := nameOf(p, at)
		if named == "" {
			named = at
		}
		why.Over = append(why.Over, Rung{Rung: rungEnvironment, Names: named})
	}
	if name := h.environmentWorkspace(); from < 2 && name != "" {
		why.Over = append(why.Over, Rung{Rung: WorkspaceVariable, Names: name})
	}
	if name := p.DefaultWorkspace(); from < 3 && name != "" {
		why.Over = append(why.Over, Rung{Rung: rungDefault, Names: name})
	}
	if store := p.FileStore(); from < 4 && store != "" {
		why.Over = append(why.Over, Rung{Rung: rungFile, Names: store})
	}
	return why
}

func (w Why) said() string {
	said := w.Rung
	if len(w.Unsaid) > 0 {
		no := make([]string, len(w.Unsaid))
		for i, rung := range w.Unsaid {
			no[i] = "no " + unsaidAs(rung)
		}
		said += " (" + strings.Join(no, ", ") + ")"
	}
	if len(w.Over) == 1 {
		said += " — over " + w.Over[0].Rung + ", " + w.Over[0].Names
	}
	if len(w.Over) > 1 {
		over := make([]string, len(w.Over))
		for i, one := range w.Over {
			over[i] = one.Rung + " (" + one.Names + ")"
		}
		said += " — over " + InWords(over)
	}
	if w.Started {
		said = "when this server started, " + said
	}
	return said
}

func unsaidAs(rung string) string {
	switch rung {
	case rungDefault:
		return "default"
	case rungFile:
		return "store in the file"
	}
	return rung
}

// InWords joins the items given as a sentence lists them: commas, and a last "and".
func InWords(items []string) string {
	if len(items) < 2 {
		return strings.Join(items, "")
	}
	return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
}

func (a answering) shown() WorkspaceShown {
	own, _ := config.NameIn(config.SettingsBeside(a.store.At))
	shown := WorkspaceShown{Workspace: a.workspace, Folder: a.folder, Why: a.why, Store: a.store.At, Notes: a.notes.At, Own: own, Colour: colourOf(a.workspace, a.store.At)}
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

func workspaceFolder(at string) error {
	if err := os.MkdirAll(filepath.Join(at, notesFolder), 0o700); err != nil {
		return err
	}
	return settingsSayingNothing(at)
}

func settingsSayingNothing(folder string) error {
	file, err := os.OpenFile(filepath.Join(folder, config.SettingsFile), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, fs.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = file.WriteString(config.SettingsHeader)
	return errors.Join(err, file.Close())
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
		return Answer{Workspaces: &Workspaces{Workspaces: workspacesIn(read), Recent: recentIn(on.Home, read)}}, nil
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
			return said, workspaceFolder(at)
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

// WorkspaceRemove takes a workspace out of the configuration, and leaves its folder as it is.
var WorkspaceRemove = Operation{
	Thing:   "workspace",
	Verb:    "forget",
	Short:   "Forget a workspace, leaving its folder as it is",
	Does:    "Take a workspace out of the configuration file, leaving the rest of the file as it was and its folder, its store and its notes untouched; the default is refused until another is made the default. A folder moved is named again with cairn workspace add.",
	Example: "cairn workspace forget play",
	Level:   Write,
	Needs:   AConfiguration,
	Answers: WorkspaceRemoved{},
	Takes: []Argument{
		{Name: "name", Kind: Text, Positional: true, Required: true, Means: "the workspace's name"},
	},
	Run: func(on On, in Input) (Answer, error) {
		removed := WorkspaceRemoved{Name: in.Given["name"]}
		err := withConfiguration(on.Home, func(was []byte, read config.Places) ([]byte, error) {
			one, err := workspaceNamed(read, removed.Name)
			if err != nil {
				return nil, err
			}
			if read.DefaultWorkspace() == removed.Name {
				return nil, fmt.Errorf("%q %w", removed.Name, ErrRemovingTheDefault)
			}
			removed.Folder = one.Folder
			return config.WithoutWorkspace(was, removed.Name)
		})
		if err != nil {
			return Answer{}, err
		}
		removed.Written, _ = on.Home.file()
		return Answer{WorkspaceRemoved: &removed}, nil
	},
}

// WorkspaceShow says which workspace answers, why, and where its store and notes are.
var WorkspaceShow = Operation{
	Thing:   "workspace",
	Verb:    "show",
	Short:   "Say which workspace answers, and why",
	Does:    "Say which workspace answers this command and why — the rung that named its store, a flag, the environment, the default workspace, the file or cairn's own default, and the rungs under it it passed over — the store and notes it uses, its time and the pages that are on, each with where it is said; a store no workspace names is said, with how to name it.",
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
		shown.Pages, shown.PagesFrom = []string{}, whereSaid(read, "pages.tasks", "pages.objectives", "pages.time")
		for i, on := range read.Pages() {
			if on {
				shown.Pages = append(shown.Pages, []string{PageTasks, PageObjectives, PageTime}[i])
			}
		}
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
		if one.NotFound {
			line += "  not found"
		}
		if one.Time != "" {
			line += "  time " + one.Time
		}
		said.WriteString(line + "  " + one.Colour + "\n")
	}
	return said.String()
}

func workspaceAddedText(added WorkspaceAdded) string {
	return fmt.Sprintf("workspace %s is %s, written in %s\nits store is %s, its notes %s\ncairn workspace use %s makes it the default\n",
		added.Name, added.Folder, added.Written, added.Store, added.Notes, added.Name)
}

func workspaceRemovedText(removed WorkspaceRemoved) string {
	return fmt.Sprintf("workspace %s is no longer named, in %s\n%s is left as it was; cairn workspace add names it again\n", removed.Name, removed.Written, removed.Folder)
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
	fmt.Fprintf(&said, "why        %s\nstore      %s\nnotes      %s\n", shown.Why.said(), shown.Store, shown.Notes)
	if shown.Colour != "" {
		fmt.Fprintf(&said, "colour     %s\n", shown.Colour)
	}
	if shown.Time != "" {
		fmt.Fprintf(&said, "time       %s, %s\n", shown.Time, shown.TimeFrom)
	}
	if shown.PagesFrom != "" {
		pages := InWords(shown.Pages)
		if pages == "" {
			pages = "none"
		}
		fmt.Fprintf(&said, "pages      %s, %s\n", pages, shown.PagesFrom)
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
