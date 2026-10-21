package cli

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

func (w *world) configPath() string {
	return w.ConfigPath()
}

func (w *world) appendConfig(content string) error {
	f, err := os.OpenFile(w.configPath(), os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(content)
	return err
}

func mentionsPath(where, what string) bool {
	strip := func(s string) string {
		return strings.Join(strings.Fields(strings.ToLower(s)), "")
	}
	return strings.Contains(strip(where), strip(what))
}

func (w *world) itRefuses() error {
	if w.Out.Code != exitRefused {
		return fmt.Errorf("exit %d, want %d (%s)", w.Out.Code, exitRefused, w.Out.Stderr)
	}
	return nil
}

func (w *world) expand(place string) string {
	if rest, ok := strings.CutPrefix(place, "~"); ok {
		return filepath.Join(w.Defaults, rest)
	}
	return place
}

func (w *world) registerConfig(sc *godog.ScenarioContext) {
	sc.Step(`^nothing configured anywhere yet$`, func() error {
		w.Setenv(config.StoreVariable, "")
		if err := os.Unsetenv(config.StoreVariable); err != nil {
			return err
		}
		w.Setenv("XDG_CONFIG_HOME", "")
		return os.Unsetenv("XDG_CONFIG_HOME")
	})

	sc.Step(`^a configuration that says nothing about the store$`, func() error {
		return w.WriteConfig("notes = \"~/vault\"\n")
	})

	sc.Step(`^a configuration that says nothing about the notes$`, func() error {
		return w.WriteConfig("store = \"~/.cairn/cairn.db\"\n")
	})

	sc.Step(`^no configuration file$`, func() error {
		if err := os.Remove(w.configPath()); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	})

	sc.Step(`^a configuration file that cannot be read$`, func() error {
		if err := w.WriteConfig("notes = \"~/vault\"\n"); err != nil {
			return err
		}
		return os.Chmod(w.configPath(), 0o000)
	})

	sc.Step(`^it refuses, and names the file it could not read$`, func() error {
		if err := w.itRefuses(); err != nil {
			return err
		}
		if !mentionsPath(w.Out.Stderr, w.configPath()) {
			return fmt.Errorf("the reason does not name the file %q:\n%s", w.configPath(), w.Out.Stderr)
		}
		return nil
	})

	sc.Step(`^the configuration still says the notes live at "([^"]*)"$`, func(place string) error {
		read, err := os.ReadFile(w.configPath())
		if err != nil {
			return err
		}
		if want := fmt.Sprintf("notes = %q\n", place); string(read) != want {
			return fmt.Errorf("the configuration reads %q, want %q", read, want)
		}
		return nil
	})

	sc.Step(`^a configuration that says the store lives at "([^"]*)"$`, func(place string) error {
		return w.WriteConfig(fmt.Sprintf("store = %q\n", place))
	})

	sc.Step(`^the environment says the store lives at "([^"]*)"$`, func(place string) error {
		w.Setenv(config.StoreVariable, place)
		return nil
	})

	sc.Step(`^the environment says the notes live at "([^"]*)"$`, func(place string) error {
		w.Setenv(config.NotesVariable, place)
		return nil
	})

	sc.Step(`^I add a task titled "([^"]*)"$`, func(title string) error {
		w.Run(fmt.Sprintf(`cairn task new "%s"`, title))
		return nil
	})

	sc.Step(`^the task was written to "([^"]*)"$`, func(place string) error {
		path := w.expand(place)
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("the task was not written to %s: %w", path, err)
		}
		return nil
	})

	sc.Step(`^it names the file it wrote$`, func() error {
		return test.WantContains(w.Out.Stdout, w.ConfigPath(), "the answer")
	})
	sc.Step(`^the file it wrote is the configuration the README shows$`, func() error {
		written, err := os.ReadFile(w.ConfigPath())
		if err != nil {
			return err
		}
		if string(written) != config.Example {
			return fmt.Errorf("it wrote:\n%s\nand the configuration cairn shows is:\n%s", written, config.Example)
		}
		return nil
	})
	sc.Step(`^it says a configuration is already there$`, func() error {
		return test.WantContains(w.Out.Stderr, "already there", "the reason")
	})
	sc.Step(`^a configuration that says the plugin lives at "([^"]*)"$`, func(place string) error {
		return w.WriteConfig(fmt.Sprintf("plugin = %q\n", place))
	})
	sc.Step(`^"([^"]*)" holds the Claude Code plugin$`, func(place string) error {
		_, err := os.Stat(filepath.Join(w.expand(place), ".claude-plugin", "marketplace.json"))
		return err
	})
	sc.Step(`^it says the plugin is at "([^"]*)", named by (.+)$`, func(place, by string) error {
		return w.namedBy(place, by)
	})

	sc.Step(`^a configuration that says the notes live at "([^"]*)"$`, func(place string) error {
		return w.WriteConfig(fmt.Sprintf("notes = %q\n", place))
	})

	sc.Step(`^a configuration that is not TOML$`, func() error {
		return w.WriteConfig("this is not toml: [[[\n")
	})

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

	sc.Step(`^a configuration whose sprints begin on (\d{4}-\d{2}-\d{2}) and last (\d+) days$`, func(day string, length int) error {
		return w.WriteConfig(fmt.Sprintf("[sprint]\nstart = %s\ndays = %d\n", day, length))
	})
	sc.Step(`^a configuration whose sprints begin on "([^"]*)"$`, func(day string) error {
		return w.WriteConfig(fmt.Sprintf("[sprint]\nstart = %q\n", day))
	})
	sc.Step(`^it says the sprints begin on (\d{4}-\d{2}-\d{2}), every (\d+) days$`, func(day string, length int) error {
		return test.WantContains(w.Out.Stdout, fmt.Sprintf("%s, every %d days", day, length), "the sprints")
	})
	sc.Step(`^it says there are no sprints$`, func() error {
		return test.WantContains(w.Out.Stdout, "no sprints", "the sprints")
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

	sc.Step("^`cairn task list` on that same configuration does refuse$", func() error {
		w.Run("cairn task list")
		return w.itRefuses()
	})

	sc.Step(`^the structured answer carries the file it read, and what it could not make sense of$`,
		w.jsonCarriesWhatItCouldNotMakeSenseOf)
}

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

func (w *world) namedBy(place, by string) error {
	if !mentionsPath(w.Out.Stdout, w.expand(place)) {
		return fmt.Errorf("the card does not name %q:\n%s", place, w.Out.Stdout)
	}
	return test.WantContains(w.Out.Stdout, "named by "+by, "the card")
}
