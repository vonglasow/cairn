package cli

import (
	"bytes"
	"cairn/internal/gui"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"

	"cairn/internal/config"
	"cairn/internal/debug"
	"cairn/internal/engine"
	"cairn/test"
)

func initFixtures(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerFixtures(sc)
	}
}

func TestAcceptance_Fixtures(t *testing.T) {
	runAcceptance(t, "fixtures.feature", initFixtures(t), "")
}

func (w *world) registerFixtures(sc *godog.ScenarioContext) {
	var folder, from string
	var held []byte
	var openedOn string
	was := openGui
	made := func(extra string) {
		w.Run(fmt.Sprintf("cairn fixtures %s%s", folder, extra))
	}
	inTheStore := func(line string) error {
		mine := w.StorePath
		w.Setenv(config.StoreVariable, filepath.Join(folder, "cairn.db"))
		defer w.Setenv(config.StoreVariable, mine)
		w.Run(line)
		if w.Out.Code != exitOK {
			return fmt.Errorf("%s: exit %d (%s)", line, w.Out.Code, w.Out.Stderr)
		}
		return nil
	}
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		folder, from, held, openedOn = filepath.Join(w.T.TempDir(), "fixtures"), "", nil, ""
		return ctx, nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		openGui = was
		return ctx, err
	})
	sc.Step(`^I make the fixtures in a folder of my own$`, func() { made("") })
	sc.Step(`^I made the fixtures in a folder of my own$`, func() error {
		made("")
		return w.ItSucceeded()
	})
	sc.Step(`^I make the fixtures in that folder(?: again)?$`, func() { made("") })
	sc.Step(`^I make the fixtures in that folder again with --new$`, func() { made(" --new") })
	sc.Step(`^I make the fixtures in a folder of my own from that file$`, func() { made(" --from " + from) })
	sc.Step(`^I make the fixtures in a folder of my own with --gui$`, func() {
		openGui = func(work engine.Work, _ engine.Pages, _, _ string, _ *debug.Log, _, _ int, _ gui.Here) error {
			openedOn = "opened"
			_, err := work.Version()
			return err
		}
		made(" --gui")
	})
	sc.Step(`^it says where the store is, and the line to open it with$`, func() error {
		store := filepath.Join(folder, "cairn.db")
		if err := test.WantContains(w.Out.Stdout, store, "the answer"); err != nil {
			return err
		}
		return test.WantContains(w.Out.Stdout, "CAIRN_STORE="+store, "the answer")
	})
	sc.Step(`^the store it made holds tasks$`, func() error {
		if err := inTheStore("cairn task list"); err != nil {
			return err
		}
		if bytes.Count([]byte(w.Out.Stdout), []byte("\n#")) < 5 {
			return fmt.Errorf("the store holds hardly anything:\n%s", w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^I added a task titled "([^"]*)" to the store it made$`, func(title string) error {
		return inTheStore(fmt.Sprintf("cairn task new %q", title))
	})
	sc.Step(`^the store it made holds a task titled "([^"]*)"$`, func(title string) error {
		if err := inTheStore("cairn task search " + title); err != nil {
			return err
		}
		return test.WantContains(w.Out.Stdout, title, "the store")
	})
	sc.Step(`^the store it made holds no task titled "([^"]*)"$`, func(title string) error {
		if err := inTheStore("cairn task search " + title); err != nil {
			return err
		}
		if bytes.Contains([]byte(w.Out.Stdout), []byte(title)) {
			return fmt.Errorf("the store still holds %q:\n%s", title, w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^a folder of my own holding a file$`, func() error {
		held = []byte("mine\n")
		if err := os.MkdirAll(folder, 0o700); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(folder, "mine.txt"), held, 0o600)
	})
	sc.Step(`^it says the folder was not made by cairn fixtures$`, func() error {
		return test.WantContains(w.Out.Stderr, "not made by cairn fixtures", "the refusal")
	})
	sc.Step(`^the file in that folder is as it was$`, func() error {
		read, err := os.ReadFile(filepath.Join(folder, "mine.txt"))
		if err != nil || !bytes.Equal(read, held) {
			return fmt.Errorf("the file reads %q: %w", read, err)
		}
		return nil
	})
	sc.Step(`^a file of my own holding a task titled "([^"]*)"$`, func(title string) error {
		from = filepath.Join(w.T.TempDir(), "mine.json")
		return os.WriteFile(from, []byte(fmt.Sprintf(`{"cairn": "tasks", "version": 4, "objectives": [],
			"tasks": [{"number": 1, "title": %q, "state": "todo", "priority": "normal"}]}`, title)), 0o600)
	})
	sc.Step(`^the store it made has a task in every column$`, func() error {
		for _, column := range []string{"backlog", "todo", "current", "blocked", "watching", "paused", "done"} {
			if err := inTheStore("cairn task list --state " + column); err != nil {
				return err
			}
			if !bytes.Contains([]byte(w.Out.Stdout), []byte("#")) {
				return fmt.Errorf("no task in %s", column)
			}
		}
		return nil
	})
	sc.Step(`^the store it made has tasks in focus and tasks archived$`, func() error {
		if err := inTheStore("cairn task focus"); err != nil {
			return err
		}
		if bytes.Count([]byte(w.Out.Stdout), []byte("(in focus)")) < 2 {
			return fmt.Errorf("fewer than two tasks in focus:\n%s", w.Out.Stdout)
		}
		if err := inTheStore("cairn task list --archived"); err != nil {
			return err
		}
		if !bytes.Contains([]byte(w.Out.Stdout), []byte("#")) {
			return fmt.Errorf("no task archived")
		}
		return nil
	})
	sc.Step(`^the window was opened on the store it made$`, func() error {
		if openedOn == "" {
			return fmt.Errorf("the window was never opened")
		}
		return nil
	})
}
