package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	"github.com/pelletier/go-toml/v2"

	"cairn/internal/engine"
	"cairn/test"
)

func (w *workspaces) registerPlaces(sc *godog.ScenarioContext) {
	var files []string
	sc.StepContext().Before(func(ctx context.Context, step *godog.Step) (context.Context, error) {
		if strings.HasPrefix(step.Text, "I run") || strings.HasPrefix(step.Text, "I add") || strings.HasPrefix(step.Text, "the MCP server") {
			files = w.filesUnderHome()
		}
		if strings.HasPrefix(step.Text, "I added") {
			_, err := engine.UpgradeConfiguration(w.Home())
			return ctx, err
		}
		return ctx, nil
	})
	sc.Step(`^I am in "([^"]*)"$`, func(dir string) error {
		w.Here = w.expand(dir)
		return os.MkdirAll(w.Here, 0o700)
	})
	sc.Step(`^the workspace "([^"]*)" at "([^"]*)", with no default$`, func(name, folder string) error {
		if err := os.MkdirAll(w.expand(folder), 0o700); err != nil {
			return err
		}
		return w.WriteConfig(fmt.Sprintf("[workspaces]\n%s = %q\n", name, folder))
	})
	sc.Step(`^the workspace "([^"]*)" at "([^"]*)", its notes at "([^"]*)", the default being "([^"]*)"$`, func(name, folder, notes, chosen string) error {
		if err := os.MkdirAll(w.expand(folder), 0o700); err != nil {
			return err
		}
		return w.WriteConfig(withNotes(name, folder, notes, chosen))
	})
	sc.Step(`^it says the workspace is "([^"]*)", at "([^"]*)"$`, func(name, folder string) error {
		if err := test.WantContains(w.Out.Stdout, "workspace  "+name+"\n", "the answer"); err != nil {
			return err
		}
		return test.WantContains(w.Out.Stdout, "folder     "+w.expand(folder)+"\n", "the answer")
	})
	sc.Step(`^it says the workspace is "([^"]*)", because the current directory is under its folder$`, func(name string) error {
		if err := test.WantContains(w.Out.Stdout, "workspace  "+name+"\n", "the answer"); err != nil {
			return err
		}
		return test.WantContains(w.Out.Stdout, "why        the current directory", "the answer")
	})
	sc.Step(`^a configuration that says the store lives at "([^"]*)" and the notes at "([^"]*)"$`, func(store, notes string) error {
		return w.WriteConfig(fmt.Sprintf("store = %q\nnotes = %q\n", store, notes))
	})
	sc.Step(`^the configuration names the workspace "([^"]*)" at "([^"]*)"(?:, its notes at "([^"]*)")?, the default$`, func(name, folder, notes string) error {
		said, err := w.workspacesSaid()
		if err != nil {
			return err
		}
		if said.chosen != name {
			return fmt.Errorf("the default is %q, want %q:\n%s", said.chosen, name, said.text)
		}
		if got := said.folders[name]; w.expand(got) != w.expand(folder) {
			return fmt.Errorf("%q is at %q, want %q:\n%s", name, got, folder, said.text)
		}
		if notes != "" && w.expand(said.notes[name]) != w.expand(notes) {
			return fmt.Errorf("the notes of %q are at %q, want %q:\n%s", name, said.notes[name], notes, said.text)
		}
		return nil
	})
	sc.Step(`^the configuration names the workspace "([^"]*)" the default, and no workspace "([^"]*)"$`, func(name, none string) error {
		said, err := w.workspacesSaid()
		if err != nil {
			return err
		}
		if _, there := said.folders[none]; there || said.chosen != name {
			return fmt.Errorf("the configuration says the default %q, with the workspaces %v; want %q and no %q:\n%s", said.chosen, said.folders, name, none, said.text)
		}
		return nil
	})
	sc.Step(`^the configuration names no top-level store or notes$`, func() error {
		var top struct {
			Store *string `toml:"store"`
			Notes *string `toml:"notes"`
		}
		read, err := os.ReadFile(w.configPath())
		if err != nil {
			return err
		}
		if err := toml.Unmarshal(read, &top); err != nil {
			return err
		}
		if top.Store != nil || top.Notes != nil {
			return fmt.Errorf("the configuration still names a top-level place:\n%s", read)
		}
		return nil
	})
	sc.Step(`^the configuration as it was is kept in "([^"]*)"$`, func(name string) error {
		kept, err := os.ReadFile(filepath.Join(filepath.Dir(w.configPath()), name))
		if err != nil {
			return err
		}
		if !bytes.Contains(kept, []byte("store = ")) {
			return fmt.Errorf("%s does not hold the configuration as it was:\n%s", name, kept)
		}
		return nil
	})
	sc.Step(`^it says on standard error what it changed, and where the old configuration is$`, func() error {
		if err := test.WantContains(w.Out.Stderr, `"main"`, "standard error"); err != nil {
			return err
		}
		return test.WantContains(w.Out.Stderr, w.configPath()+".v0.11.0", "standard error")
	})
	sc.Step(`^nothing was moved(?: on disk)?$`, func() error {
		for _, file := range files {
			if _, err := os.Stat(file); err != nil {
				return fmt.Errorf("%s was there before and is not now: %w", file, err)
			}
		}
		return nil
	})
	sc.Step(`^the MCP server starts$`, func() error {
		w.Out = runWith(w.context(), w.Home(), "mcp", "serve", "--access", "full")
		return nil
	})
	sc.Step(`^the workspace "([^"]*)" whose notes were at "([^"]*)", holding task 1's folder$`, func(name, notes string) error {
		folder := filepath.Dir(notes)
		if err := os.MkdirAll(filepath.Join(w.expand(notes), "1"), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(w.expand(notes), "1", "summary.md"), []byte("where it stands\n"), 0o600); err != nil {
			return err
		}
		if err := w.WriteConfig(fmt.Sprintf("[workspaces]\ndefault = %q\n%s = %q\n", name, name, folder)); err != nil {
			return err
		}
		if r := runWith(w.context(), w.Home(), "task", "list"); r.Code != 0 {
			return fmt.Errorf("cairn task list: exit %d (%s)", r.Code, r.Stderr)
		}
		w.held = []byte(name + "\n" + folder)
		return nil
	})
	sc.Step(`^the configuration now names its notes at "([^"]*)"(?:, which holds "([^"]*)")?$`, func(notes, holding string) error {
		name, folder, _ := strings.Cut(string(w.held), "\n")
		if holding != "" {
			if err := os.MkdirAll(w.expand(notes), 0o700); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(w.expand(notes), holding), []byte("mine\n"), 0o600); err != nil {
				return err
			}
		}
		return w.WriteConfig(withNotes(name, folder, notes, name))
	})
	sc.Step("^I run `cairn ([^`]*)` in a terminal(?:, and answer \"([^\"]*)\")?$", func(line, answer string) error {
		var out, errs bytes.Buffer
		ctx := withAsking(w.context(), true)
		code := MainIn(ctx, w.Home(), test.SplitArgs(line), strings.NewReader(answer+"\n"), &out, &errs)
		w.Line, w.Out = "cairn "+line, test.Result{Stdout: out.String(), Stderr: errs.String(), Code: code}
		return nil
	})
	sc.Step(`^"([^"]*)" is there$`, func(place string) error {
		_, err := os.Stat(w.expand(place))
		return err
	})
	sc.Step(`^"([^"]*)" is gone$`, func(place string) error {
		if _, err := os.Stat(w.expand(place)); !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s is still there: %w", place, err)
		}
		return nil
	})
	sc.Step(`^it asks whether to move the notes$`, func() error {
		return test.WantContains(w.Out.Stderr, "Move them there? [y/N]", "standard error")
	})
	sc.Step(`^it says on standard error that a run of cairn in a terminal will offer to move them$`, func() error {
		return test.WantContains(w.Out.Stderr, "a run of cairn in a terminal will offer to move them", "standard error")
	})
	sc.Step(`^it says it cannot move them there, naming "([^"]*)"$`, func(place string) error {
		if err := test.WantContains(w.Out.Stderr, "cannot move them", "standard error"); err != nil {
			return err
		}
		return test.WantContains(w.Out.Stderr, w.expand(place), "standard error")
	})
}

func withNotes(name, folder, notes, chosen string) string {
	return fmt.Sprintf("[workspaces]\ndefault = %q\n\n[workspaces.%s]\nfolder = %q\nnotes = %q\n", chosen, name, folder, notes)
}

type workspacesRead struct {
	chosen  string
	folders map[string]string
	notes   map[string]string
	text    string
}

func (w *world) workspacesSaid() (workspacesRead, error) {
	read, err := os.ReadFile(w.configPath())
	if err != nil {
		return workspacesRead{}, err
	}
	var doc struct {
		Workspaces map[string]any `toml:"workspaces"`
	}
	if err := toml.Unmarshal(read, &doc); err != nil {
		return workspacesRead{}, fmt.Errorf("%w:\n%s", err, read)
	}
	said := workspacesRead{folders: map[string]string{}, notes: map[string]string{}, text: string(read)}
	for name, value := range doc.Workspaces {
		switch one := value.(type) {
		case string:
			if name == "default" {
				said.chosen = one
				continue
			}
			said.folders[name], said.notes[name] = one, filepath.Join(one, "notes")
		case map[string]any:
			folder, _ := one["folder"].(string)
			notes, _ := one["notes"].(string)
			if notes == "" {
				notes = filepath.Join(folder, "notes")
			}
			said.folders[name], said.notes[name] = folder, notes
		}
	}
	return said, nil
}

func (w *world) filesUnderHome() []string {
	var files []string
	_ = filepath.WalkDir(w.Defaults, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && !strings.Contains(path, string(filepath.Separator)+".config"+string(filepath.Separator)) && !strings.Contains(d.Name(), "cairn.db") {
			files = append(files, path)
		}
		return nil
	})
	return files
}
