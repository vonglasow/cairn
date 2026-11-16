package engine

import (
	"cairn/internal/engine/workspace"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"cairn/internal/store/files"
)

// Keep settles where h answers from, once, so that a server started on it answers from there whatever the configuration names afterwards; what cannot be settled yet is left to be found at each call.
func Keep(h Home) Home {
	if _, err := Locate(h); err != nil {
		return h
	}
	_, a, _ := h.settled()
	a.Why.Started = true
	h.Kept = &workspace.Started{Workspace: a.Workspace, Implicit: a.Unnamed, Folder: a.Folder, Why: a.Why, Store: a.Store, Notes: a.Notes}
	return h
}

func (h Home) file() (string, error) {
	if h.Dir == "" {
		return "", fmt.Errorf("%w: %w", errStore, h.Missing)
	}
	return files.FileIn(h.Dir), nil
}

func (h Home) directoryWorkspace(p files.Places) (files.Workspace, bool) {
	if h.Here == "" {
		return files.Workspace{}, false
	}
	var chosen files.Workspace
	deepest := 0
	for _, one := range p.Workspaces() {
		if _, err := os.Stat(one.Folder); err != nil {
			continue
		}
		for _, root := range h.rootsOf(one) {
			if workspace.Within(h.Here, root) && len(root) > deepest {
				chosen, deepest = one, len(root)
			}
		}
	}
	return chosen, deepest > 0
}

func (h Home) rootsOf(one files.Workspace) []string {
	roots := []string{one.Folder}
	if project := filepath.Dir(one.Folder); filepath.Base(one.Folder) == files.DefaultFolder && project != h.Dir {
		roots = append(roots, project)
	}
	return roots
}

func (h Home) answering(p files.Places) (workspace.Answering, error) {
	if k := h.Kept; k != nil {
		return workspace.Answering{Store: k.Store, Notes: k.Notes, Workspace: k.Workspace, Folder: k.Folder, Why: k.Why, Unnamed: k.Implicit}, nil
	}
	if h.Workspace != "" && h.Store != "" {
		return workspace.Answering{}, Misread(fmt.Errorf("%s and %s both name the store: give one or the other", workspace.WorkspaceFlagName, workspace.StoreFlagName))
	}
	gate := &workspace.PlacesGate{}
	p.Tell(gate)
	a := workspace.Answering{Store: gate.Store, Notes: gate.Notes}
	if h.Workspace != "" {
		one, err := workspace.There(p, h.Workspace)
		if err != nil {
			return a, err
		}
		a.Store, a.Notes = workspace.InWorkspace(one, workspace.WorkspaceFlagName)
		a.Workspace, a.Why = one.Name, h.why(p, 0)
		return a.Named(p), nil
	}
	if h.Store != "" {
		at, err := files.Absolute(h.Store, h.Dir)
		if err != nil {
			return a, fmt.Errorf("%s %w", workspace.StoreFlagName, err)
		}
		a.Store, a.Why = workspace.Emplacement{At: at, By: workspace.StoreFlagName}, h.why(p, 0)
		a.Why.Rung = workspace.RungStoreFlag
		return a.OwnNotes(p), nil
	}
	if a.Store.By == workspace.ByTheEnvironment {
		a.Why = h.why(p, 1)
		return a.OwnNotes(p), nil
	}
	if h.environmentWorkspace() != "" {
		name := h.environmentWorkspace()
		one, err := workspace.There(p, name)
		if err != nil {
			return a, fmt.Errorf("%s: %w", workspace.WorkspaceVariable, err)
		}
		a.Why = h.why(p, 2)
		return a.Inside(p, one, fmt.Sprintf("%s, %q", workspace.WorkspaceVariable, name)), nil
	}
	if one, found := h.directoryWorkspace(p); found {
		a.Why = h.why(p, 3)
		return a.Inside(p, one, fmt.Sprintf("the current directory, under the workspace %q", one.Name)), nil
	}
	a.Why = h.why(p, 6)
	if a.Store.By == "the file" {
		a.Why = h.why(p, 5)
	}
	if name := p.DefaultWorkspace(); name != "" {
		one, err := workspace.There(p, name)
		if err != nil {
			return a, err
		}
		a.Why = h.why(p, 4)
		return a.Inside(p, one, fmt.Sprintf("the default workspace, %q", name)), nil
	}
	return a.Named(p), nil
}

func (h Home) environmentWorkspace() string {
	said, _ := h.lookup()(workspace.WorkspaceVariable)
	return said
}

func (h Home) why(p files.Places, from int) workspace.Why {
	why := workspace.Why{Rung: workspace.Rungs[from]}
	if from >= 4 {
		why.Unsaid = slices.DeleteFunc(slices.Clone(workspace.Rungs[:from]), func(rung string) bool { return rung == workspace.WorkspaceVariable || rung == workspace.RungDirectory })
	}
	if said, ok := h.lookup()(files.StoreVariable); from < 1 && ok && said != "" {
		at, err := files.Absolute(said, h.Dir)
		if err != nil {
			at = said
		}
		named := workspace.NameOf(p, at)
		if named == "" {
			named = at
		}
		why.Over = append(why.Over, workspace.Rung{Rung: workspace.RungEnvironment, Names: named})
	}
	if name := h.environmentWorkspace(); from < 2 && name != "" {
		why.Over = append(why.Over, workspace.Rung{Rung: workspace.WorkspaceVariable, Names: name})
	}
	if one, found := h.directoryWorkspace(p); from < 3 && found {
		why.Over = append(why.Over, workspace.Rung{Rung: workspace.RungDirectory, Names: one.Name})
	}
	if name := p.DefaultWorkspace(); from < 4 && name != "" {
		why.Over = append(why.Over, workspace.Rung{Rung: workspace.RungDefault, Names: name})
	}
	if store := p.FileStore(); from < 5 && store != "" {
		why.Over = append(why.Over, workspace.Rung{Rung: workspace.RungFile, Names: store})
	}
	return why
}

func withConfiguration(h Home, change func([]byte, files.Places) ([]byte, error)) error {
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

// WorkspaceList lists every workspace the configuration names, the default marked.
var WorkspaceList = Operation{
	Thing:   "workspace",
	Verb:    "list",
	Short:   "List the workspaces, the default marked",
	Does:    "List every workspace the configuration names: its name and the folder holding its store and its notes, the default marked.",
	Example: "cairn workspace list",
	Level:   Read,
	Needs:   aConfiguration,
	Answers: workspace.Workspaces{},
	Run: func(on workedOn, _ Input) (Answer, error) {
		read, err := on.Home.configuration()
		if err != nil {
			return Answer{}, err
		}
		return Answer{Workspaces: &workspace.Workspaces{Workspaces: workspace.WorkspacesIn(read), Recent: recentIn(on.Home, read)}}, nil
	},
}

// WorkspaceAdd names a workspace in the configuration, and makes its folder.
var WorkspaceAdd = Operation{
	Thing:   "workspace",
	Verb:    "add",
	Short:   "Name a workspace, a folder holding a store and its notes",
	Does:    "Name a workspace in the configuration file, writing the file when there is none and leaving the rest of it as it was, and make its folder with an empty notes folder in it and settings.toml saying nothing yet. Its store is cairn.db in the folder, its notes the notes folder; cairn workspace use makes it the default.",
	Example: "cairn workspace add work ~/cairn/work",
	Level:   write,
	Needs:   aConfiguration,
	Again:   ChangesNothingMore,
	Answers: workspace.Added{},
	Takes: []Argument{
		{Name: "name", Kind: Text, Positional: true, Required: true, Means: "the workspace's name: lowercase letters, digits and dashes"},
		{Name: "folder", Kind: Text, Positional: true, Required: true, Means: "the folder holding its store and its notes: absolute, or under ~"},
		{Name: "colour", Kind: Text, Means: "the colour it wears, one of the label colours; without it, one read off its name"},
		{Name: "notes", Kind: Text, Means: "the folder its notes are written in, a vault's for one: absolute, or under ~; without it, notes in its folder"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		name, folder, notes := in.Given["name"], in.Given["folder"], in.Given["notes"]
		if !workspace.PlainName.MatchString(name) || name == files.DefaultWorkspaceKey {
			return Answer{}, fmt.Errorf("%q: %w", name, workspace.ErrNotAWorkspaceName)
		}
		at, err := files.Absolute(folder, on.Home.Dir)
		if err != nil {
			return Answer{}, err
		}
		one := files.Workspace{Name: name, Folder: at}
		if notes != "" {
			if one.Notes, err = notesOf(at, notes, on.Home.Dir); err != nil {
				return Answer{}, err
			}
		}
		colour, _ := files.KeyNamed("workspace.colour")
		literal, wears := "", in.Given["colour"] != ""
		if wears {
			if literal, err = on.Home.literal(colour, in.Given["colour"], false); err != nil {
				return Answer{}, err
			}
		}
		err = withConfiguration(on.Home, func(was []byte, read files.Places) ([]byte, error) {
			if _, err := workspace.Named(read, name); err == nil {
				return nil, fmt.Errorf("%q: %w", name, workspace.ErrWorkspaceKnown)
			}
			said, err := files.WithWorkspace(was, name, folder)
			if err == nil && notes != "" {
				said, err = files.WithWorkspaceNotes(said, name, notes)
			}
			if err != nil {
				return nil, err
			}
			return said, workspace.MakeFolder(one)
		})
		if err != nil {
			return Answer{}, err
		}
		if wears {
			said := func(was []byte) ([]byte, error) { return files.WithKey(was, colour.Name, literal) }
			if err := workspace.Rewritten(filepath.Join(at, files.SettingsFile), on.Home.Dir, true, said); err != nil {
				return Answer{}, err
			}
		}
		store, written := workspace.InWorkspace(one, "")
		path, _ := on.Home.file()
		return Answer{WorkspaceAdded: &workspace.Added{Name: name, Folder: at, Store: store.At, Notes: written.At, Written: path}}, nil
	},
}

// WorkspaceUse makes a workspace the default.
var WorkspaceUse = Operation{
	Thing:   "workspace",
	Verb:    "use",
	Short:   "Make a workspace the default",
	Does:    "Make a workspace the default, the one used when neither a flag nor the environment names another, and say which one it was.",
	Example: "cairn workspace use home",
	Level:   write,
	Needs:   aConfiguration,
	Again:   ChangesNothingMore,
	Answers: workspace.Used{},
	Takes: []Argument{
		{Name: "name", Kind: Text, Positional: true, Required: true, Means: "the workspace's name"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		used := workspace.Used{Default: in.Given["name"]}
		err := withConfiguration(on.Home, func(was []byte, read files.Places) ([]byte, error) {
			if _, err := workspace.There(read, used.Default); err != nil {
				return nil, err
			}
			used.Was = read.DefaultWorkspace()
			return files.WithDefault(was, used.Default)
		})
		if err != nil {
			return Answer{}, err
		}
		return Answer{WorkspaceUsed: &used}, nil
	},
}

// workspaceRemove takes a workspace out of the configuration, and leaves its folder as it is.
var workspaceRemove = Operation{
	Thing:   "workspace",
	Verb:    "forget",
	Short:   "Forget a workspace, leaving its folder as it is",
	Does:    "Take a workspace out of the configuration file, leaving the rest of the file as it was and its folder, its store and its notes untouched; the default is refused until another is made the default. A folder moved is named again with cairn workspace add.",
	Example: "cairn workspace forget play",
	Level:   write,
	Needs:   aConfiguration,
	Answers: workspace.Removed{},
	Takes: []Argument{
		{Name: "name", Kind: Text, Positional: true, Required: true, Means: "the workspace's name"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		removed := workspace.Removed{Name: in.Given["name"]}
		err := withConfiguration(on.Home, func(was []byte, read files.Places) ([]byte, error) {
			one, err := workspace.Named(read, removed.Name)
			if err != nil {
				return nil, err
			}
			if read.DefaultWorkspace() == removed.Name {
				return nil, fmt.Errorf("%q %w", removed.Name, workspace.ErrRemovingTheDefault)
			}
			removed.Folder = one.Folder
			return files.WithoutWorkspace(was, removed.Name)
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
	Needs:   aConfiguration,
	Answers: workspace.Shown{},
	Run: func(on workedOn, _ Input) (Answer, error) {
		read, answers, err := on.Home.settled()
		if err != nil {
			return Answer{}, err
		}
		shown := answers.Shown()
		shown.Time, shown.TimeFrom = read.TimeMode(), workspace.WhereSaid(read, "pages.time")
		shown.Pages, shown.PagesFrom = []string{}, workspace.WhereSaid(read, "pages.tasks", "pages.objectives", "pages.time")
		for i, on := range read.Pages() {
			if on {
				shown.Pages = append(shown.Pages, []string{PageTasks, PageObjectives, PageTime}[i])
			}
		}
		if on.Home.Kept != nil {
			shown.Switch = workspace.SwitchingIsNotHere
		}
		return Answer{WorkspaceShown: &shown}, nil
	},
}

func notesOf(folder, said, home string) (string, error) {
	at, err := files.Absolute(said, home)
	if err != nil {
		return "", err
	}
	if workspace.Within(folder, at) {
		return "", fmt.Errorf("%s %w", said, workspace.ErrNotesHoldTheBoard)
	}
	return at, nil
}
