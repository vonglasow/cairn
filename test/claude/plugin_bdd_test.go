package claude

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

type pluginWorld struct {
	*sessionWorld
	out    string
	ran    error
	folder string
}

func (w *pluginWorld) beside() string {
	for _, kv := range w.env {
		if store, ok := strings.CutPrefix(kv, "CAIRN_STORE="); ok {
			return filepath.Join(filepath.Dir(store), "claude-plugin")
		}
	}
	return ""
}

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

func (w *pluginWorld) olderPluginIn(folder string) error {
	if _, err := w.cairn("plugin", "write"); err != nil {
		return err
	}
	plugin, err := offers(folder, "cairn")
	if err != nil {
		return err
	}
	older := []byte(`{"name": "cairn", "version": "v0.0.1"}`)
	if err := os.WriteFile(filepath.Join(plugin, manifest), older, 0o600); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(plugin, "skills", "gone.md"), []byte("an older skill"), 0o600)
}

func (w *pluginWorld) saysItsCairn(folder string) error {
	written, err := offers(folder, "cairn")
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
		sc.Step("^I run `cairn plugin write --dir ([^`]*)`$", func(dir string) { run("plugin", "write", "--dir", filepath.Join(t.TempDir(), dir)) })
		sc.Step(`^it fails, saying the flag is unknown$`, func() error {
			if w.ran == nil || !strings.Contains(w.out, "unknown flag") {
				return fmt.Errorf("it said %q, %w", w.out, w.ran)
			}
			return nil
		})
		sc.Step(`^a configuration that says the plugin lives in a folder of my own$`, func() error {
			w.folder = filepath.Join(t.TempDir(), "my plugins")
			home := ""
			for _, kv := range w.env {
				if at, ok := strings.CutPrefix(kv, "HOME="); ok {
					home = at
				}
			}
			dir := filepath.Join(home, ".config", "cairn")
			if err := os.MkdirAll(dir, 0o700); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(dir, "config.toml"), []byte(fmt.Sprintf("plugin = %q\n", w.folder)), 0o600)
		})
		sc.Step(`^the plugin was written there by an older cairn$`, func() error {
			return w.olderPluginIn(w.folder)
		})
		sc.Step(`^the plugin in that folder says the version of the cairn that wrote it$`, func() error {
			return w.saysItsCairn(w.folder)
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
		sc.Step(`^the plugin was written beside the store by an older cairn$`, func() error {
			return w.olderPluginIn(w.beside())
		})
		sc.Step(`^the MCP server starts$`, func() { run("mcp", "serve") })
		sc.Step(`^nothing is written beside the store$`, func() error {
			if _, err := os.Stat(w.beside()); err == nil {
				return fmt.Errorf("%s was written, and nobody asked for it", w.beside())
			}
			return nil
		})
		sc.Step(`^I read the cairn skill$`, func() error {
			if _, err := w.cairn("plugin", "write"); err != nil {
				return err
			}
			written, err := offers(w.beside(), "cairn")
			if err != nil {
				return err
			}
			said, err := os.ReadFile(filepath.Join(written, "skills", "cairn", "SKILL.md"))
			w.out = string(said)
			return err
		})
		sc.Step(`^the skill says "([^"]*)"$`, func(want string) error {
			return test.WantContains(w.out, want, "the skill")
		})
		sc.Step(`^it succeeds$`, func() error { return w.ran })
		sc.Step(`^beside the store is a marketplace offering the plugin "([^"]*)"$`, func(name string) error {
			_, err := offers(w.beside(), name)
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
			for path := range want {
				if strings.HasSuffix(path, ".go") {
					delete(want, path)
				}
			}
			for path, held := range want {
				if path == manifest {
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
			return w.saysItsCairn(w.beside())
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

const manifest = ".claude-plugin/plugin.json"

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
