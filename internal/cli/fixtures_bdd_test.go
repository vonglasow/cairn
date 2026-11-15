package cli

import (
	"bytes"
	"cairn/internal/board"
	"cairn/internal/gui"
	"cairn/internal/gui/design"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"

	"cairn/internal/debug"
	"cairn/internal/engine"
	"cairn/internal/store/files"
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
	var folder string
	var held []byte
	var openedOn string
	made := func(extra string) {
		w.Run(fmt.Sprintf("cairn fixtures %s%s", folder, extra))
	}
	inTheStore := func(line string) error {
		mine := w.StorePath
		w.Setenv(files.StoreVariable, filepath.Join(folder, "cairn.db"))
		defer w.Setenv(files.StoreVariable, mine)
		w.Run(line)
		if w.Out.Code != exitOK {
			return fmt.Errorf("%s: exit %d (%s)", line, w.Out.Code, w.Out.Stderr)
		}
		return nil
	}
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		folder, held, openedOn = filepath.Join(w.T.TempDir(), "fixtures"), nil, ""
		return ctx, nil
	})
	sc.Step(`^I make the fixtures in a folder of my own$`, func() { made("") })
	sc.Step(`^I make the fixtures in that folder(?: again)?$`, func() { made("") })
	sc.Step(`^I make the fixtures in a folder of my own with --gui$`, func() {
		w.window = func(work board.Work, _ engine.Pages, _ design.Skin, _ *debug.Log, _, _ int, _ gui.Here) error {
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
	w.registerFixturesCases(sc, inTheStore)
	sc.Step(`^the window was opened on the store it made$`, func() error {
		if openedOn == "" {
			return fmt.Errorf("the window was never opened")
		}
		return nil
	})
}
