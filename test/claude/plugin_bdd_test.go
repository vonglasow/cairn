package claude

// Executable acceptance scenarios for bdd/claude-plugin.feature: the plugin
// the binary carries, written out by the real binary and held against the
// plugin in this repository, file by file.

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"cairn/test"
)

// pluginWorld is the session world, and what the last command said.
type pluginWorld struct {
	*sessionWorld
	out    string
	ran    error
	folder string
}

// beside is where the plugin goes when no folder is named: next to the store.
func (w *pluginWorld) beside() string {
	for _, kv := range w.env {
		if store, ok := strings.CutPrefix(kv, "CAIRN_STORE="); ok {
			return filepath.Join(filepath.Dir(store), "claude-plugin")
		}
	}
	return ""
}

// files is every file under a folder, by its path inside it, with what it holds.
func files(root string) (map[string]string, error) {
	found := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		said, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		inside, _ := filepath.Rel(root, path)
		found[filepath.ToSlash(inside)] = string(said)
		return nil
	})
	return found, err
}

// offers says a folder is a marketplace offering the plugin named, and where
// that plugin lies.
func offers(folder, name string) (string, error) {
	var market struct {
		Name    string `json:"name"`
		Plugins []struct {
			Name   string `json:"name"`
			Source string `json:"source"`
		} `json:"plugins"`
	}
	said, err := os.ReadFile(filepath.Join(folder, ".claude-plugin", "marketplace.json"))
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(said, &market); err != nil {
		return "", err
	}
	for _, p := range market.Plugins {
		if p.Name == name {
			return filepath.Join(folder, p.Source), nil
		}
	}
	return "", fmt.Errorf("the marketplace in %s offers no plugin %q", folder, name)
}

func initPlugin(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := &pluginWorld{sessionWorld: &sessionWorld{t: t}}
		sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
			w.reset()
			w.out, w.ran, w.folder = "", nil, ""
			return ctx, nil
		})
		run := func(args ...string) {
			w.out, w.ran = w.cairn(args...)
		}
		sc.Step("^I run `cairn plugin write`$", func() { run("plugin", "write") })
		sc.Step(`^I write the plugin into a folder of my own$`, func() {
			w.folder = filepath.Join(t.TempDir(), "my plugins")
			run("plugin", "write", "--dir", w.folder)
		})
		sc.Step(`^the plugin was written beside the store with a file it no longer carries$`, func() error {
			if _, err := w.cairn("plugin", "write"); err != nil {
				return err
			}
			plugin, err := offers(w.beside(), "cairn")
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(plugin, "skills", "gone.md"), []byte("an older skill"), 0o600)
		})
		sc.Step(`^it succeeds$`, func() error { return w.ran })
		sc.Step(`^beside the store is a marketplace offering the plugin "([^"]*)"$`, func(name string) error {
			_, err := offers(w.beside(), name)
			return err
		})
		sc.Step(`^that folder is a marketplace offering the plugin "([^"]*)"$`, func(name string) error {
			_, err := offers(w.folder, name)
			return err
		})
		sc.Step(`^the plugin written there is the plugin in this repository$`, func() error {
			written, err := offers(w.beside(), "cairn")
			if err != nil {
				return err
			}
			got, err := files(written)
			if err != nil {
				return err
			}
			want, err := files(plugin)
			if err != nil {
				return err
			}
			// The Go that carries the plugin inside the binary is not part
			// of it: Claude Code reads the rest.
			for path := range want {
				if strings.HasSuffix(path, ".go") {
					delete(want, path)
				}
			}
			for path, held := range want {
				if path == manifest {
					// The manifest is written with the version of the cairn
					// that wrote it, and is otherwise the repository's.
					if !sameButVersion(got[path], held) {
						return fmt.Errorf("%s says other than the repository's, its version aside", path)
					}
					continue
				}
				if got[path] != held {
					return fmt.Errorf("%s is not written as the repository holds it", path)
				}
			}
			for path := range got {
				if _, ok := want[path]; !ok {
					return fmt.Errorf("%s is written, and the plugin carries no such file", path)
				}
			}
			return nil
		})
		sc.Step(`^the plugin written there says the version of the cairn that wrote it$`, func() error {
			written, err := offers(w.beside(), "cairn")
			if err != nil {
				return err
			}
			var said struct {
				Version string `json:"version"`
			}
			held, err := os.ReadFile(filepath.Join(written, manifest))
			if err != nil {
				return err
			}
			if err := json.Unmarshal(held, &said); err != nil {
				return err
			}
			version, _ := w.cairn("--version")
			if said.Version == "" || !strings.Contains(version, said.Version) {
				return fmt.Errorf("the plugin says version %q, and the cairn that wrote it says %q", said.Version, version)
			}
			return nil
		})
		sc.Step(`^the hook written there can be run$`, func() error {
			written, err := offers(w.beside(), "cairn")
			if err != nil {
				return err
			}
			info, err := os.Stat(filepath.Join(written, "hooks", "cairn-session.sh"))
			if err != nil {
				return err
			}
			if info.Mode()&0o111 == 0 {
				return fmt.Errorf("the hook is written %v, and cannot be run", info.Mode())
			}
			return nil
		})
		sc.Step(`^it says to run "([^"]*)" and "([^"]*)"$`, func(first, second string) error {
			for _, said := range []string{first + " " + w.beside(), second} {
				if !strings.Contains(w.out, said) {
					return fmt.Errorf("it does not say %q:\n%s", said, w.out)
				}
			}
			return nil
		})
	}
}

func TestAcceptance_ClaudePlugin(t *testing.T) {
	if test.Suite(t, "claude-plugin.feature", initPlugin(t), "").Run() != 0 {
		t.Fatal("the Claude plugin's scenarios did not all pass")
	}
}

// manifest is where the plugin's own manifest lies, inside the plugin.
const manifest = ".claude-plugin/plugin.json"

// sameButVersion says two manifests say the same, whatever version either
// carries.
func sameButVersion(a, b string) bool {
	var left, right map[string]any
	if json.Unmarshal([]byte(a), &left) != nil || json.Unmarshal([]byte(b), &right) != nil {
		return false
	}
	delete(left, "version")
	delete(right, "version")
	l, _ := json.Marshal(left)
	r, _ := json.Marshal(right)
	return string(l) == string(r)
}
