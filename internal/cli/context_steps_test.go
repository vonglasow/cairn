package cli

// Steps observed only by bdd/context.feature.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/config"
)

// contextLine is the line of a card that says where a task's outside context
// lives, and nothing else.
func (w *world) contextLine() (string, bool) {
	for _, l := range strings.Split(w.Out.Stdout, "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(l), "context"); ok {
			return strings.TrimSpace(rest), true
		}
	}
	return "", false
}

// registerContext declares the steps only bdd/context.feature uses.
func (w *world) registerContext(sc *godog.ScenarioContext) {
	// Every place a show said, in order: two shows in a row each overwrite
	// what the last command answered, and the comparison needs both.
	var said []string
	sc.StepContext().After(func(ctx context.Context, step *godog.Step, _ godog.StepResultStatus, err error) (context.Context, error) {
		// The step running a show against each store keeps its own two.
		if strings.Contains(step.Text, "cairn task show") && !strings.HasSuffix(step.Text, "against each store") {
			if place, ok := w.contextLine(); ok {
				said = append(said, place)
			}
		}
		return ctx, err
	})

	sc.Step(`^it says the context lives at "([^"]*)"$`, func(place string) error {
		got, ok := w.contextLine()
		if !ok {
			return fmt.Errorf("it says nothing about where the context lives:\n%s", w.Out.Stdout)
		}
		if want := w.expand(place); got != want {
			return fmt.Errorf("it says the context lives at %q, want %q", got, want)
		}
		return nil
	})

	sc.Step(`^no line says where a context lives$`, func() error {
		if place, ok := w.contextLine(); ok {
			return fmt.Errorf("a line says a context lives at %q", place)
		}
		return nil
	})

	sc.Step(`^I moved task (\d+) under task (\d+)$`, func(n, parent int) error {
		return w.Seed(fmt.Sprintf(`cairn task edit %d --parent %d`, n, parent))
	})

	sc.Step(`^the two places are not the same$`, func() error {
		if len(said) != 2 {
			return fmt.Errorf("%d places were said, want 2: %v", len(said), said)
		}
		if said[0] == said[1] {
			return fmt.Errorf("both tasks are sent to %q", said[0])
		}
		return nil
	})

	// Another store, the same configuration: the notes are found the way the
	// store is, and only the store moves.
	var stores []string
	sc.Step(`^another store of my own, somewhere else again$`, func() error {
		stores = append(stores, w.StorePath)
		w.StorePath = filepath.Join(w.T.TempDir(), "other", "cairn.db")
		stores = append(stores, w.StorePath)
		w.Setenv(config.StoreVariable, w.StorePath)
		return nil
	})

	sc.Step("^I run `([^`]*)` against each store$", func(line string) error {
		for _, store := range stores {
			w.Setenv(config.StoreVariable, store)
			w.Run(line)
			if place, ok := w.contextLine(); ok {
				said = append(said, place)
			}
		}
		return nil
	})

	sc.Step(`^both say the context lives at the same place$`, func() error {
		if len(said) != 2 {
			return fmt.Errorf("%d places were said, want 2: %v", len(said), said)
		}
		if said[0] != said[1] {
			return fmt.Errorf("one store says %q and the other %q", said[0], said[1])
		}
		return nil
	})

	sc.Step(`^nothing was ever written under "([^"]*)"$`, func(place string) error {
		if _, err := os.Stat(w.expand(place)); err == nil {
			return fmt.Errorf("something is already under %s", w.expand(place))
		}
		return nil
	})

	// The place and nothing beside it: no word on whether it is there, empty,
	// or full — a claim about the disk made now and read later.
	sc.Step(`^it says nothing about whether anything is there$`, func() error {
		got, _ := w.contextLine()
		if got != w.expand("~/vault/1") {
			return fmt.Errorf("the context line reads %q, more than the place", got)
		}
		for _, word := range []string{"exist", "empty", "missing", "nothing there", "not found"} {
			if strings.Contains(strings.ToLower(w.Out.Stdout), word) {
				return fmt.Errorf("it says %q about the place:\n%s", word, w.Out.Stdout)
			}
		}
		return nil
	})

	sc.Step(`^nothing was made under "([^"]*)"$`, func(place string) error {
		if _, err := os.Stat(w.expand(place)); err == nil {
			return fmt.Errorf("%s was made", w.expand(place))
		}
		return nil
	})

	// Written by hand, the way a harness or a person would — never by cairn.
	sc.Step(`^notes were written under "([^"]*)"$`, func(place string) error {
		dir := w.expand(place)
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "2026-08-11-meeting.md"), []byte("what was decided\n"), 0o600)
	})

	sc.Step(`^what was written under "([^"]*)" is still there$`, func(place string) error {
		path := filepath.Join(w.expand(place), "2026-08-11-meeting.md")
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("the notes are gone: %w", err)
		}
		return nil
	})
}
