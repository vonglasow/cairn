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
		return os.WriteFile(filepath.Join(at, "settings.toml"), []byte(said.Content+"\n"), 0o600)
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
		line := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + ` = "?` + regexp.QuoteMeta(value) + `"?\s+# ` + regexp.QuoteMeta(from) + `$`)
		if !line.MatchString(w.Out.Stdout) {
			return fmt.Errorf("the answer does not say %s is %s, %s:\n%s", key, value, from, w.Out.Stdout)
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
