package cli

import (
	"cairn/internal/engine/workspace"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/store/files"
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
		w.Setenv(files.StoreVariable, "")
		if err := os.Unsetenv(files.StoreVariable); err != nil {
			return err
		}
		w.Setenv("XDG_CONFIG_HOME", "")
		return os.Unsetenv("XDG_CONFIG_HOME")
	})

	sc.Step(`^a configuration that says nothing about the store$`, func() error {
		return w.WriteConfig("notes = \"~/vault\"\n")
	})

	sc.Step(`^a configuration whose default workspace is "([^"]*)" at "([^"]*)"(?:, its notes at "([^"]*)")?$`, func(name, folder, notes string) error {
		if filepath.IsAbs(w.expand(folder)) {
			if err := os.MkdirAll(w.expand(folder), 0o700); err != nil {
				return err
			}
		}
		if notes == "" {
			return w.WriteConfig(fmt.Sprintf("[workspaces]\ndefault = %q\n%s = %q\n", name, name, folder))
		}
		return w.WriteConfig(withNotes(name, folder, notes, name))
	})

	sc.Step(`^the environment says the store lives at "([^"]*)"$`, func(place string) error {
		w.Setenv(files.StoreVariable, place)
		return nil
	})

	sc.Step(`^the environment names the workspace "([^"]*)"$`, func(name string) error {
		w.Setenv(workspace.WorkspaceVariable, name)
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
	sc.Step(`^the file it wrote is the configuration the documentation shows$`, func() error {
		written, err := os.ReadFile(w.ConfigPath())
		if err != nil {
			return err
		}
		if string(written) != files.Example {
			return fmt.Errorf("it wrote:\n%s\nand the configuration cairn shows is:\n%s", written, files.Example)
		}
		return nil
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

	sc.Step(`^it refuses, and names "([^"]*)"$`, func(name string) error {
		if err := w.itRefuses(); err != nil {
			return err
		}
		return test.WantContains(w.Out.Stderr, name, "the reason")
	})

	sc.Step(`^it says on standard error that "([^"]*)" is ignored$`, func(key string) error {
		return test.WantContains(w.Out.Stderr, fmt.Sprintf("%q is not a key this cairn knows", key), "standard error")
	})
	sc.Step(`^it says on standard error "([^"]*)"$`, w.itSaysOnStandardError)
	sc.Step(`^it says "([^"]*)" is ignored$`, func(key string) error {
		return test.WantContains(w.Out.Stdout, fmt.Sprintf("# ignored: %q", key), "config show")
	})

	sc.Step(`^it refuses, and says the place must be absolute$`, func() error {
		if err := w.itRefuses(); err != nil {
			return err
		}
		return test.WantContains(w.Out.Stderr, "absolute", "the reason")
	})

	sc.Step(`^a configuration that archives every (\d+) days from (\d{4}-\d{2}-\d{2})$`, func(length int, day string) error {
		return w.WriteConfig(fmt.Sprintf("[archive]\nstart = %s\ndays = %d\n", day, length))
	})
	sc.Step(`^it says it archives every (\d+) days from (\d{4}-\d{2}-\d{2})$`, func(length int, day string) error {
		if err := w.saysKey("start", day); err != nil {
			return err
		}
		return w.saysKey("days", fmt.Sprint(length))
	})

	sc.Step(`^it names "([^"]*)"$`, func(place string) error {
		return test.WantContains(w.Out.Stderr, place, "the reason")
	})

	sc.Step(`^it says the store is at "([^"]*)", named by (.+)$`, func(place, by string) error {
		return w.namedBy(place, by)
	})

	sc.Step(`^it says "([^"]*)" is "([^"]*)"$`, w.saysKey)
	sc.Step(`^a configuration that says "([^"]*)" is "([^"]*)" under "(\[[^"]*\])"$`, func(key, value, table string) error {
		return w.WriteConfig(fmt.Sprintf("%s\n%s = %s\n", table, key, value))
	})
	sc.Step(`^a configuration whose approach label is "([^"]*)"$`, func(label string) error {
		return w.WriteConfig(fmt.Sprintf("[approach]\nlabel = %q\n", label))
	})
	sc.Step(`^it says no label asks for the approach to be agreed first$`, func() error {
		return test.WantContains(w.Out.Stdout, "no label asks for the approach to be agreed first", "the approach")
	})

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

}

func (w *world) namedBy(place, by string) error {
	if !mentionsPath(w.Out.Stdout, w.expand(place)) {
		return fmt.Errorf("the card does not name %q:\n%s", place, w.Out.Stdout)
	}
	return test.WantContains(w.Out.Stdout, "named by "+by, "the card")
}

func (w *world) saysKey(key, value string) error {
	line := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + ` = "?` + regexp.QuoteMeta(w.expand(value)) + `"?\s*(#.*)?$`)
	if !line.MatchString(w.Out.Stdout) {
		return fmt.Errorf("the answer does not say %s is %s:\n%s", key, value, w.Out.Stdout)
	}
	return nil
}

func (w *world) itSaysOnStandardError(said string) error {
	return test.WantContains(w.Out.Stderr, said, "standard error")
}
