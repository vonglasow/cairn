package cli

// Steps observed only by bdd/config.feature: the file cairn reads its two
// places from, and how the environment and the default sit around it.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/config"
	"cairn/test"
)

// configPath is where the configuration lives for this scenario's home:
// ~/.config/cairn/config.toml, the same on every platform this runs on.
// Spelled once, in test.World — the world every surface shares, and the one
// Seed itself sets aside — rather than a second time here.
func (w *world) configPath() string {
	return w.ConfigPath()
}

// writeConfig poses a configuration file by hand, the way a person would have
// — never through cairn, so a scenario asserting what cairn does with it is
// not also asserting that cairn wrote it.
func (w *world) writeConfig(content string) error {
	path := w.configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o600)
}

// appendConfig adds to a configuration already posed, rather than replacing
// it — a scenario building up a file line by line, the way "that same
// configuration says something under..." reads.
func (w *world) appendConfig(content string) error {
	f, err := os.OpenFile(w.configPath(), os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(content)
	return err
}

// mentionsPath says whether a rendered reason names a path, ignoring every
// blank on both sides. The renderer wraps a long line wherever it must,
// including in the middle of a path, and test.Flatten turns that wrap into a
// space — which a path never carries. Dropping blanks altogether is safe for
// a path and would not be in general: a path is one token, so nothing can be
// glued to it that was not already beside it. Measured on this machine: the
// default TMPDIR is long enough that the path wraps, and the scenario failed
// on the comparison while the reason cairn wrote held the path intact.
func mentionsPath(where, what string) bool {
	strip := func(s string) string {
		return strings.Join(strings.Fields(strings.ToLower(s)), "")
	}
	return strings.Contains(strip(where), strip(what))
}

// itRefuses fails unless the last command line was understood and turned
// down — exit 1, not any other code a failure could carry. exitStore (3) and
// exitUsage (2) both look like "it failed" to a step that only checked the
// code was not exitOK, and the distinction internal/cli/exit.go draws
// between the three is exactly what these scenarios are pinning down.
func (w *world) itRefuses() error {
	if w.Out.Code != exitRefused {
		return fmt.Errorf("exit %d, want %d (%s)", w.Out.Code, exitRefused, w.Out.Stderr)
	}
	return nil
}

// expand turns a scenario's "~/..." into a real path under this run's own
// home, the way the file itself is read against it rather than against
// whoever runs the suite.
func (w *world) expand(place string) string {
	if rest, ok := strings.CutPrefix(place, "~"); ok {
		return filepath.Join(w.Defaults, rest)
	}
	return place
}

// registerConfig declares the steps only bdd/config.feature uses.
func (w *world) registerConfig(sc *godog.ScenarioContext) {
	sc.Step(`^nothing configured anywhere yet$`, func() error {
		w.Setenv(config.StoreVariable, "")
		if err := os.Unsetenv(config.StoreVariable); err != nil {
			return err
		}
		w.Setenv("XDG_CONFIG_HOME", "")
		return os.Unsetenv("XDG_CONFIG_HOME")
	})

	// A file that names the other place, not an empty one. "Says nothing about
	// the store" has to be told apart from there being no file at all, which
	// is what @config-that-is-not-there-is-the-ordinary-case covers on its own:
	// an empty file decodes to the same nothing as an absent one, and the two
	// scenarios would guard one behaviour between them instead of two.
	sc.Step(`^a configuration that says nothing about the store$`, func() error {
		return w.writeConfig("notes = \"~/vault\"\n")
	})

	sc.Step(`^a configuration that says the store lives at "([^"]*)"$`, func(place string) error {
		return w.writeConfig(fmt.Sprintf("store = %q\n", place))
	})

	// The variable carries what a person would have typed, "~" and all: it is
	// cairn that expands it, so a step that expanded it first would be testing
	// itself.
	sc.Step(`^the environment says the store lives at "([^"]*)"$`, func(place string) error {
		w.Setenv(config.StoreVariable, place)
		return nil
	})

	// The notes' twin of the step above.
	sc.Step(`^the environment says the notes live at "([^"]*)"$`, func(place string) error {
		w.Setenv(config.NotesVariable, place)
		return nil
	})

	// The behaviour under test, not a precondition: it goes through the real
	// command, the same as `I run`.
	sc.Step(`^I add a task titled "([^"]*)"$`, func(title string) error {
		w.Run(fmt.Sprintf(`cairn task add "%s"`, title))
		return nil
	})

	sc.Step(`^the task was written to "([^"]*)"$`, func(place string) error {
		path := w.expand(place)
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("the task was not written to %s: %w", path, err)
		}
		return nil
	})

	sc.Step(`^a configuration that says the notes live at "([^"]*)"$`, func(place string) error {
		return w.writeConfig(fmt.Sprintf("notes = %q\n", place))
	})

	// Not TOML at all, rather than TOML that says something cairn does not
	// know: @config-an-unknown-name-is-refused covers the second on its own,
	// and the two would guard one behaviour between them instead of two.
	sc.Step(`^a configuration that is not TOML$`, func() error {
		return w.writeConfig("this is not toml: [[[\n")
	})

	// Appended, not written afresh: the scenario reads as one configuration
	// that says two things, not two configurations replacing each other.
	sc.Step(`^that same configuration says something under "([^"]*)"$`, func(name string) error {
		return w.appendConfig(fmt.Sprintf("%s = \"something\"\n", name))
	})

	sc.Step(`^it refuses, and names the file it could not make sense of$`, func() error {
		if err := w.itRefuses(); err != nil {
			return err
		}
		if !mentionsPath(w.Out.Stderr, w.configPath()) {
			return fmt.Errorf("the reason does not name the file %q:\n%s", w.configPath(), w.Out.Stderr)
		}
		return nil
	})

	sc.Step(`^it refuses, and names "([^"]*)"$`, func(name string) error {
		if err := w.itRefuses(); err != nil {
			return err
		}
		return test.WantContains(w.Out.Stderr, name, "the reason")
	})

	// What it accepts instead of a name it does not know: the two places a
	// configuration is allowed to speak of.
	sc.Step(`^it names what it accepts instead$`, func() error {
		if err := test.WantContains(w.Out.Stderr, "store", "the reason"); err != nil {
			return err
		}
		return test.WantContains(w.Out.Stderr, "notes", "the reason")
	})

	sc.Step(`^it refuses, and says the place must be absolute$`, func() error {
		if err := w.itRefuses(); err != nil {
			return err
		}
		return test.WantContains(w.Out.Stderr, "absolute", "the reason")
	})

	sc.Step(`^it names "([^"]*)"$`, func(place string) error {
		return test.WantContains(w.Out.Stderr, place, "the reason")
	})

	sc.Step(`^it names the file it read$`, func() error {
		if !mentionsPath(w.Out.Stdout, w.configPath()) {
			return fmt.Errorf("the card does not name the file %q:\n%s", w.configPath(), w.Out.Stdout)
		}
		return nil
	})

	sc.Step(`^it says the store is at "([^"]*)", named by (.+)$`, func(place, by string) error {
		return w.namedBy(place, by)
	})

	sc.Step(`^it says the notes are at "([^"]*)", named by (.+)$`, func(place, by string) error {
		return w.namedBy(place, by)
	})

	sc.Step(`^the structured answer carries the file it read, each place, and what named it$`,
		w.jsonCarriesEachPlace)

	sc.Step(`^it says the file cannot be made sense of$`, func() error {
		return test.WantContains(w.Out.Stdout, "cannot be made sense of", "the card")
	})

	sc.Step(`^it says the other commands will refuse until it is$`, func() error {
		return test.WantContains(w.Out.Stdout, "other commands will refuse until it is", "the card")
	})

	// A real invocation, not a simulation: this measures the previous step's
	// claim instead of taking it on faith.
	sc.Step("^`cairn task list` on that same configuration does refuse$", func() error {
		w.Run("cairn task list")
		return w.itRefuses()
	})

	sc.Step(`^the structured answer carries the file it read, and what it could not make sense of$`,
		w.jsonCarriesWhatItCouldNotMakeSenseOf)
}

// jsonCarriesEachPlace fails unless the JSON `config show --json` printed
// carries the file it read and, for each of its two places, where it
// resolved to and what named it.
func (w *world) jsonCarriesEachPlace() error {
	var doc map[string]any
	if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
		return fmt.Errorf("the answer is not JSON (%w):\n%s", err, w.Out.Stdout)
	}
	if !mentionsPath(fmt.Sprint(doc["file"]), w.configPath()) {
		return fmt.Errorf("the answer does not carry the file %q:\n%s", w.configPath(), w.Out.Stdout)
	}
	for _, key := range []string{"store", "notes"} {
		place, ok := doc[key].(map[string]any)
		if !ok {
			return fmt.Errorf("the answer carries no %q place:\n%s", key, w.Out.Stdout)
		}
		at, _ := place["at"].(string)
		if at == "" {
			return fmt.Errorf("the answer's %q names no place:\n%s", key, w.Out.Stdout)
		}
		by, _ := place["named_by"].(string)
		if by == "" {
			return fmt.Errorf("the answer's %q does not say what named it:\n%s", key, w.Out.Stdout)
		}
	}
	return nil
}

// jsonCarriesWhatItCouldNotMakeSenseOf fails unless the JSON `config show
// --json` printed carries the file it read and what stood in the way of
// making sense of it.
func (w *world) jsonCarriesWhatItCouldNotMakeSenseOf() error {
	var doc map[string]any
	if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
		return fmt.Errorf("the answer is not JSON (%w):\n%s", err, w.Out.Stdout)
	}
	if !mentionsPath(fmt.Sprint(doc["file"]), w.configPath()) {
		return fmt.Errorf("the answer does not carry the file %q:\n%s", w.configPath(), w.Out.Stdout)
	}
	problem, _ := doc["problem"].(string)
	if problem == "" {
		return fmt.Errorf("the answer does not carry what it could not make sense of:\n%s", w.Out.Stdout)
	}
	return nil
}

// namedBy fails unless the card names place among what it printed, alongside
// what named it — used for both the store and the notes, since a card says
// the same two things about either.
func (w *world) namedBy(place, by string) error {
	if !mentionsPath(w.Out.Stdout, w.expand(place)) {
		return fmt.Errorf("the card does not name %q:\n%s", place, w.Out.Stdout)
	}
	return test.WantContains(w.Out.Stdout, "named by "+by, "the card")
}
