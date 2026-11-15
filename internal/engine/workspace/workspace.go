// Package workspace holds a workspace: its places, its configuration and settings, the recent ones.
package workspace

import (
	"cairn/internal/engine/label"
	"cairn/internal/store/files"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// The names of a workspace's places, of its flags, and of the rungs that can name the store.
const (
	StoreFile         = "cairn.db"
	notesFolder       = "notes"
	ByTheEnvironment  = "the environment"
	WorkspaceFlagName = "--workspace"
	StoreFlagName     = "--store"
	RungDirectory     = "the current directory"
	RungDefault       = "the default"
	RungFile          = "the file"
	RungNothing       = "where cairn keeps a store by default"
)

// The rungs a flag or the environment names.
const (
	RungWorkspaceFlag = WorkspaceFlagName
	RungStoreFlag     = StoreFlagName
	RungEnvironment   = files.StoreVariable
)

// WorkspaceVariable is the environment variable naming a workspace: under CAIRN_STORE and CAIRN_NOTES, which name places, and over the default workspace.
const WorkspaceVariable = "CAIRN_WORKSPACE"

// Rungs is every rung that can name the store, the first one first.
var Rungs = []string{RungWorkspaceFlag, RungEnvironment, WorkspaceVariable, RungDirectory, RungDefault, RungFile, RungNothing}

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

// PlainName is what a workspace name holds: a to z, 0 to 9 and dashes.
var PlainName = regexp.MustCompile(`^[a-z0-9-]+$`)

// Workspace is one workspace the configuration names: its name, its folder, whether it is the default, the colour it wears, and whether its folder is gone.
type Workspace struct {
	Name     string `json:"name"`
	Folder   string `json:"folder"`
	Notes    string `json:"notes,omitempty" jsonschema:"the notes the configuration names for it; <folder>/notes when it names none"`
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
	Rung    string   `json:"rung" jsonschema:"--workspace, --store, CAIRN_STORE, CAIRN_WORKSPACE, the current directory (under a workspace's folder, or the project its .cairn folder sits in), the default (the default workspace), the file (its store key), or where cairn keeps a store by default (the workspace main)"`
	Over    []Rung   `json:"over,omitempty" jsonschema:"the rungs under it that named a store too, and were passed over"`
	Unsaid  []string `json:"unsaid,omitempty" jsonschema:"the rungs over it that named nothing"`
	Started bool     `json:"started,omitempty" jsonschema:"settled when this server started, and kept since"`
}

// Workspaces is how workspace list answers: every workspace the configuration names, by name, and the stores a board opened by path that none names, the newest first.
type Workspaces struct {
	Workspaces []Workspace    `json:"workspaces"`
	Recent     []RecentFolder `json:"recent,omitempty"`
}

// Removed is how workspace forget answers: the workspace the configuration no longer names, its folder, left as it was, and the file it was taken out of.
type Removed struct {
	Name    string `json:"name"`
	Folder  string `json:"folder"`
	Written string `json:"written"`
}

// Added is how workspace add answers: the workspace written, its two places, and the file it was written in.
type Added struct {
	Name    string `json:"name"`
	Folder  string `json:"folder"`
	Store   string `json:"store"`
	Notes   string `json:"notes"`
	Written string `json:"written"`
}

// Used is how workspace use answers: the default now, and the one it was, if any.
type Used struct {
	Default string `json:"default"`
	Was     string `json:"was,omitempty"`
}

// Shown is how workspace show answers: which workspace answers, if one does, why, the places this command uses, and how to name them as a workspace when none does.
type Shown struct {
	Workspace string   `json:"workspace,omitempty"`
	Implicit  bool     `json:"implicit,omitempty" jsonschema:"main, in ~/.cairn, answering because nothing else did and no configuration names it"`
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
	Implicit  bool        `json:"implicit,omitempty"`
	Folder    string      `json:"folder,omitempty"`
	Why       Why         `json:"why"`
	Store     Emplacement `json:"store"`
	Notes     Emplacement `json:"notes"`
}

// SwitchingIsNotHere is why a server does not switch workspaces.
const SwitchingIsNotHere = "not from here: this server keeps its store; another workspace is another server, which the person starts"

// Answering is the workspace that answers, its places and why.
type Answering struct {
	Store, Notes Emplacement
	Workspace    string
	Folder       string
	Why          Why
	Unnamed      bool
}

// WorkspacesIn is every workspace the configuration names.
func WorkspacesIn(p files.Places) []Workspace {
	var listed []Workspace
	for _, one := range p.Workspaces() {
		store := filepath.Join(one.Folder, StoreFile)
		w := Workspace{Name: one.Name, Folder: one.Folder, Notes: one.Notes, Default: one.Name == p.DefaultWorkspace(), Colour: ColourOf(one.Name, store)}
		w.NotFound = !w.Found()
		if own, err := p.WithSettings(files.SettingsBeside(store)); err == nil && own.TimeMode() != files.TimeOn {
			w.Time = own.TimeMode()
		}
		listed = append(listed, w)
	}
	return listed
}

// ColourOf is the colour a workspace wears.
func ColourOf(name, store string) string {
	said, err := files.ColourIn(files.SettingsBeside(store))
	if colour, wrong := label.ParseColour(said); err == nil && wrong == nil {
		return colour
	}
	if name == "" {
		return ""
	}
	return label.ColourOf(name)
}

// Named is the workspace the configuration names so.
func Named(p files.Places, name string) (files.Workspace, error) {
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
	return files.Workspace{}, fmt.Errorf("%q: %w; %s", name, ErrNoSuchWorkspace, there)
}

// There is the workspace named so, refused when its folder is gone.
func There(p files.Places, name string) (files.Workspace, error) {
	one, err := Named(p, name)
	if err != nil {
		return one, err
	}
	if _, err := os.Stat(one.Folder); err != nil {
		return one, fmt.Errorf("%s: %w", one.Folder, ErrWorkspaceNotThere)
	}
	return one, nil
}

// InWorkspace is the store and the notes of a workspace.
func InWorkspace(one files.Workspace, by string) (Emplacement, Emplacement) {
	notes := one.Notes
	if notes == "" {
		notes = filepath.Join(one.Folder, notesFolder)
	}
	return Emplacement{At: filepath.Join(one.Folder, StoreFile), By: by}, Emplacement{At: notes, By: by}
}

// Within says whether a folder is inside another.
func Within(dir, root string) bool {
	for _, pair := range [][2]string{{dir, root}, {resolved(dir), resolved(root)}} {
		rel, err := filepath.Rel(pair[1], pair[0])
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func resolved(path string) string {
	if real, err := filepath.EvalSymlinks(path); err == nil {
		return real
	}
	return path
}

// Inside is the answer with the places of the workspace given.
func (a Answering) Inside(p files.Places, one files.Workspace, by string) Answering {
	store, notes := InWorkspace(one, by)
	a.Store, a.Workspace = store, one.Name
	if a.Notes.By != ByTheEnvironment {
		a.Notes = notes
	}
	return a.Named(p)
}

// OwnNotes is the answer with the notes of its own workspace.
func (a Answering) OwnNotes(p files.Places) Answering {
	a = a.Named(p)
	if a.Notes.By == ByTheEnvironment {
		return a
	}
	if one, err := Named(p, a.Workspace); err == nil {
		_, a.Notes = InWorkspace(one, fmt.Sprintf("the workspace %q", a.Workspace))
	} else if main, err := Named(p, files.UpgradedMain); a.Workspace == "" && err == nil {
		_, a.Notes = InWorkspace(main, fmt.Sprintf("the workspace %q, which no other names", files.UpgradedMain))
	}
	return a
}

// Implicit says whether no workspace was named.
func (a Answering) Implicit() bool { return a.Unnamed }

// Configured is the workspace answering when the configuration names it; nothing for a store no workspace names, and for main answering because nothing named any.
func (s Shown) Configured() string {
	if s.Implicit {
		return ""
	}
	return s.Workspace
}

// Named is the answer with the name and the folder of its workspace.
func (a Answering) Named(p files.Places) Answering {
	if a.Workspace == "" {
		a.Workspace = NameOf(p, a.Store.At)
	}
	for _, one := range p.Workspaces() {
		if one.Name == a.Workspace {
			a.Folder = one.Folder
		}
	}
	if a.Workspace == "" && a.Why.Rung == RungNothing {
		a.Workspace, a.Folder, a.Unnamed = files.UpgradedMain, filepath.Dir(a.Store.At), true
	}
	return a
}

// NameOf is the name of the workspace holding the store.
func NameOf(p files.Places, store string) string {
	for _, one := range p.Workspaces() {
		if filepath.Join(one.Folder, StoreFile) == store {
			return one.Name
		}
	}
	return ""
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
	case RungDefault:
		return "default"
	case RungFile:
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

// Shown is the answer as workspace show says it.
func (a Answering) Shown() Shown {
	own, _ := files.NameIn(files.SettingsBeside(a.Store.At))
	shown := Shown{Workspace: a.Workspace, Implicit: a.Implicit(), Folder: a.Folder, Why: a.Why, Store: a.Store.At, Notes: a.Notes.At, Own: own, Colour: ColourOf(a.Workspace, a.Store.At)}
	if a.Workspace == "" {
		shown.Add = "cairn workspace add <name> <folder>"
		if filepath.Base(a.Store.At) == StoreFile {
			shown.Add = "cairn workspace add <name> " + filepath.Dir(a.Store.At)
		}
	}
	return shown
}

// MakeFolder makes the folder of a new workspace.
func MakeFolder(at string) error {
	if err := os.MkdirAll(filepath.Join(at, notesFolder), 0o700); err != nil {
		return err
	}
	return settingsSayingNothing(at)
}

func settingsSayingNothing(folder string) error {
	file, err := os.OpenFile(filepath.Join(folder, files.SettingsFile), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, fs.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = file.WriteString(files.SettingsHeader)
	return errors.Join(err, file.Close())
}

// WorkspacesText is workspace list for a person.
func WorkspacesText(listed []Workspace) string {
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

// AddedText is workspace add for a person.
func AddedText(added Added) string {
	return fmt.Sprintf("workspace %s is %s, written in %s\nits store is %s, its notes %s\ncairn workspace use %s makes it the default\n",
		added.Name, added.Folder, added.Written, added.Store, added.Notes, added.Name)
}

// RemovedText is workspace forget for a person.
func RemovedText(removed Removed) string {
	return fmt.Sprintf("workspace %s is no longer named, in %s\n%s is left as it was; cairn workspace add names it again\n", removed.Name, removed.Written, removed.Folder)
}

// UsedText is workspace use for a person.
func UsedText(used Used) string {
	was := "there was none"
	if used.Was != "" {
		was = fmt.Sprintf("it was %q", used.Was)
	}
	return fmt.Sprintf("the default workspace is now %q; %s\n", used.Default, was)
}

// ShownText is workspace show for a person.
func ShownText(shown Shown) string {
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
		if one.Notes == "" {
			fmt.Fprintf(&said, "%s = %q\n", one.Name, one.Folder)
		}
	}
	for _, one := range listed {
		if one.Notes != "" {
			fmt.Fprintf(&said, "\n[workspaces.%s]\nfolder = %q\nnotes = %q\n", one.Name, one.Folder, one.Notes)
		}
	}
	return said.String()
}

// Name is how the place a board shows is named to a person, short, never a whole path: the workspace that names its store, the name the store's settings give it, or the folder holding the store.
func (s Shown) Name() string {
	if s.Workspace != "" {
		return s.Workspace
	}
	if s.Own != "" {
		return s.Own
	}
	if s.Store == "" {
		return ""
	}
	return filepath.Base(filepath.Dir(s.Store))
}
