package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/pelletier/go-toml/v2"

	"cairn/test"
)

func initSettings(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := &workspaces{world: newWorld(t)}
		w.register(sc)
		w.registerConfig(sc)
		w.registerContext(sc)
		w.registerWorkspace(sc)
		w.registerSettings(sc)
	}
}

func TestAcceptance_Settings(t *testing.T) {
	runAcceptance(t, "settings.feature", initSettings(t), "~@later")
}

func (w *workspaces) registerSettings(sc *godog.ScenarioContext) {
	sc.Step(`^the global configuration also says:$`, func(said *godog.DocString) error {
		return w.appendConfig("\n" + said.Content + "\n")
	})
	sc.Step(`^the settings in "([^"]*)" say:$`, func(folder string, said *godog.DocString) error {
		at := w.expand(folder)
		if err := os.MkdirAll(at, 0o700); err != nil {
			return err
		}
		if w.settings == nil {
			w.settings = map[string]string{}
		}
		w.settings[folder] = said.Content + "\n"
		return os.WriteFile(filepath.Join(at, "settings.toml"), []byte(said.Content+"\n"), 0o600)
	})
	sc.Step(`^the settings in "([^"]*)" read:$`, func(folder string, said *godog.DocString) error {
		return w.fileReads(filepath.Join(w.expand(folder), "settings.toml"), said.Content+"\n")
	})
	sc.Step(`^it says it wrote the settings in "([^"]*)"$`, func(folder string) error {
		return test.WantContains(w.Out.Stdout, "written in "+filepath.Join(w.expand(folder), "settings.toml"), "the answer")
	})
	sc.Step(`^the settings in "([^"]*)" are as they were$`, func(folder string) error {
		return w.fileReads(filepath.Join(w.expand(folder), "settings.toml"), w.settings[folder])
	})
	sc.Step(`^the global configuration reads:$`, func(said *godog.DocString) error {
		return w.fileReads(w.configPath(), said.Content+"\n")
	})
	sc.Step("^the global configuration holds `([^`]*)` right under `([^`]*)`$", func(line, above string) error {
		return test.WantContains(w.readFile(w.configPath()), above+"\n"+line+"\n", "the global configuration")
	})
	sc.Step("^the global configuration holds `([^`]*)`$", func(line string) error {
		return test.WantContains(w.readFile(w.configPath()), "\n"+line+"\n", "the global configuration")
	})
	sc.Step(`^it names the global configuration$`, func() error {
		return test.WantContains(w.Out.Stderr, w.configPath(), "the reason")
	})
	sc.Step(`^it names on standard error the file "([^"]*)"$`, func(path string) error {
		return test.WantContains(w.Out.Stderr, w.expand(path), "standard error")
	})
	sc.Step(`^it names the store's settings "([^"]*)"$`, func(path string) error {
		return test.WantContains(w.Out.Stdout, "# the store's settings: "+w.expand(path)+"\n", "config show")
	})
	sc.Step(`^it says "([^"]*)" is "([^"]*)", (set in the store's settings|from the global configuration|by default)$`, func(key, value, from string) error {
		line := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + ` = "?` + regexp.QuoteMeta(value) + `"?\s+# ` + regexp.QuoteMeta(from) + `(; global: .*)?$`)
		if !line.MatchString(w.Out.Stdout) {
			return fmt.Errorf("the answer does not say %s is %s, %s:\n%s", key, value, from, w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^it says "([^"]*)" is "([^"]*)", set in the store's settings, the global configuration giving "([^"]*)"$`, func(key, value, global string) error {
		line := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + ` = "?` + regexp.QuoteMeta(value) + `"?\s+# set in the store's settings; global: "?` + regexp.QuoteMeta(global) + `"?$`)
		if !line.MatchString(w.Out.Stdout) {
			return fmt.Errorf("the answer does not say %s is %s here and %s globally:\n%s", key, value, global, w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^the structured answer says the global configuration gives "([^"]*)" as "([^"]*)"$`, func(key, value string) error {
		var doc struct {
			Global map[string]string `json:"global"`
		}
		if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
			return fmt.Errorf("the answer is not JSON (%w):\n%s", err, w.Out.Stdout)
		}
		if doc.Global[key] != value {
			return fmt.Errorf("the answer says the global configuration gives %q as %q, want %q:\n%s", key, doc.Global[key], value, w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^it says the colour is read off the workspace's name$`, func() error {
		line := regexp.MustCompile(`(?m)^colour = "[a-z]+"\s+# read off the workspace's name$`)
		if !line.MatchString(w.Out.Stdout) {
			return fmt.Errorf("the answer does not say the colour is read off the name:\n%s", w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^the structured answer says the workspace wears "([^"]*)"$`, func(colour string) error {
		var doc struct {
			Workspace struct {
				Colour string `json:"colour"`
			} `json:"workspace"`
		}
		if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
			return fmt.Errorf("the answer is not JSON (%w):\n%s", err, w.Out.Stdout)
		}
		if doc.Workspace.Colour != colour {
			return fmt.Errorf("the answer says the workspace wears %q, want %q:\n%s", doc.Workspace.Colour, colour, w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^the structured answer says "([^"]*)" is written in (the store's settings|the global configuration)$`, func(key, where string) error {
		var doc struct {
			Written     []string `json:"written"`
			WrittenHere []string `json:"written_here"`
		}
		if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
			return fmt.Errorf("the answer is not JSON (%w):\n%s", err, w.Out.Stdout)
		}
		here, global := slices.Contains(doc.WrittenHere, key), slices.Contains(doc.Written, key)
		if want := where == "the store's settings"; here != want || global == want {
			return fmt.Errorf("%q written in the store's settings %v, in the global configuration %v; want it in %s:\n%s", key, here, global, where, w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^the settings in "([^"]*)" say nothing, and that everything comes from the global configuration until they say otherwise$`, func(folder string) error {
		read, err := os.ReadFile(filepath.Join(w.expand(folder), "settings.toml"))
		if err != nil {
			return err
		}
		var said map[string]any
		if err := toml.Unmarshal(read, &said); err != nil {
			return err
		}
		if len(said) > 0 {
			return fmt.Errorf("the settings say %v, want nothing:\n%s", said, read)
		}
		for _, words := range []string{"global configuration", "until", "otherwise"} {
			if !strings.Contains(string(read), words) {
				return fmt.Errorf("the settings do not say %q:\n%s", words, read)
			}
		}
		return nil
	})
}

func (w *workspaces) readFile(path string) string {
	data, _ := os.ReadFile(path)
	return "\n" + string(data)
}

func (w *workspaces) fileReads(path, want string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if string(data) != want {
		return fmt.Errorf("%s reads\n%s\nnot\n%s", path, data, want)
	}
	return nil
}
