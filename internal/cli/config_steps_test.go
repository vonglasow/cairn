package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

	sc.Step(`^a configuration that archives every (\d+) days from (\d{4}-\d{2}-\d{2})$`, func(length int, day string) error {
		return w.WriteConfig(fmt.Sprintf("[archive]\nstart = %s\ndays = %d\n", day, length))
	})
	sc.Step(`^a configuration that archives counting from "([^"]*)"$`, func(day string) error {
		return w.WriteConfig(fmt.Sprintf("[archive]\nstart = %q\n", day))
	})
	sc.Step(`^it says it archives every (\d+) days from (\d{4}-\d{2}-\d{2})$`, func(length int, day string) error {
		if err := w.saysKey("start", day); err != nil {
			return err
		}
		return w.saysKey("days", fmt.Sprint(length))
	})
	sc.Step(`^it says nothing is archived on its own$`, func() error {
		return test.WantContains(w.Out.Stdout, "nothing is archived on its own", "the archiving")
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

	sc.Step("^`cairn config init` wrote the configuration$", func() error {
		w.Run("cairn config init")
		if w.Out.Code != exitOK {
			return fmt.Errorf("config init: exit %d (%s)", w.Out.Code, w.Out.Stderr)
		}
		return nil
	})
	sc.Step(`^I take the comment off the key "([^"]*)" in it$`, func(key string) error {
		return w.rewriteConfig(`(?m)^# (`+regexp.QuoteMeta(key)+` = .*)$`, "$1")
	})
	sc.Step(`^I change "([^"]*)" to "([^"]*)" in it$`, func(key, value string) error {
		return w.rewriteConfig(`(?m)^`+regexp.QuoteMeta(key)+` = .*$`, key+" = "+value)
	})
	sc.Step(`^it says "([^"]*)" is "([^"]*)"$`, w.saysKey)
	sc.Step(`^a configuration that says "([^"]*)" is "([^"]*)" under "(\[[^"]*\])"$`, func(key, value, table string) error {
		return w.WriteConfig(fmt.Sprintf("%s\n%s = %s\n", table, key, value))
	})
	sc.Step(`^that same configuration keeps the window's sizes$`, func() error {
		return w.appendConfig("[window]\nkeep_sizes = true\n")
	})
	sc.Step(`^it says every key of the configuration, under its table:$`, w.saysEveryKey)
	sc.Step(`^it says a focus session is never asked to be renewed$`, func() error {
		return test.WantContains(w.Out.Stdout, "never asked to be renewed", "the focus")
	})
	sc.Step(`^I write what it said as the configuration$`, func() error {
		w.said = w.Out.Stdout
		return w.WriteConfig(w.Out.Stdout)
	})
	sc.Step(`^it says the same keys and values as before$`, func() error {
		if was, now := keysAndValues(w.said), keysAndValues(w.Out.Stdout); was != now {
			return fmt.Errorf("it said:\n%s\nand now says:\n%s", was, now)
		}
		return nil
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

func (w *world) rewriteConfig(pattern, with string) error {
	read, err := os.ReadFile(w.configPath())
	if err != nil {
		return err
	}
	re := regexp.MustCompile(pattern)
	if !re.Match(read) {
		return fmt.Errorf("the configuration has no line %s:\n%s", pattern, read)
	}
	done := false
	changed := re.ReplaceAllFunc(read, func(line []byte) []byte {
		if done {
			return line
		}
		done = true
		return re.ReplaceAll(line, []byte(with))
	})
	return os.WriteFile(w.configPath(), changed, 0o600)
}

func (w *world) saysKey(key, value string) error {
	line := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + ` = "?` + regexp.QuoteMeta(w.expand(value)) + `"?\s*(#.*)?$`)
	if !line.MatchString(w.Out.Stdout) {
		return fmt.Errorf("the answer does not say %s is %s:\n%s", key, value, w.Out.Stdout)
	}
	return nil
}

func (w *world) saysEveryKey(table *godog.Table) error {
	under := map[string][]string{}
	current := ""
	for _, l := range strings.Split(w.Out.Stdout, "\n") {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "[") && strings.HasSuffix(l, "]") {
			current = strings.Trim(l, "[]")
			continue
		}
		under[current] = append(under[current], l)
	}
	for _, row := range table.Rows[1:] {
		at, want := row.Cells[0].Value, row.Cells[1].Value
		found := false
		for _, l := range under[at] {
			found = found || strings.HasPrefix(l, want)
		}
		if !found {
			return fmt.Errorf("the answer does not say %q under [%s]:\n%s", want, at, w.Out.Stdout)
		}
	}
	return nil
}

func keysAndValues(said string) string {
	var kept []string
	for _, l := range strings.Split(said, "\n") {
		l, _, _ = strings.Cut(l, "#")
		if l = strings.TrimSpace(l); l != "" {
			kept = append(kept, l)
		}
	}
	return strings.Join(kept, "\n")
}
