package claude

// The plugin's wiring, which no scenario reaches: Claude Code reads these files
// and calls the hook, and a file that names the wrong event, the wrong script
// or a server cairn does not answer to fails silently — on somebody's machine,
// with a session that never ties itself to anything.

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var plugin = filepath.Join("..", "..", "plugin", "claude")

func readJSON(t *testing.T, path string, into any) {
	t.Helper()
	said, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(said, into); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
}

type hookFile struct {
	Hooks map[string][]struct {
		Matcher string `json:"matcher"`
		Hooks   []struct {
			Type    string `json:"type"`
			Command string `json:"command"`
		} `json:"hooks"`
	} `json:"hooks"`
}

func TestPlugin_wiresEveryEventToTheHook(t *testing.T) {
	var file hookFile
	readJSON(t, filepath.Join(plugin, "hooks", "hooks.json"), &file)
	for _, event := range []string{"PostToolUse", "UserPromptSubmit", "Stop", "SessionEnd"} {
		groups := file.Hooks[event]
		if len(groups) != 1 || len(groups[0].Hooks) != 1 {
			t.Errorf("%s calls %d groups, want one hook", event, len(groups))
			continue
		}
		command := strings.ReplaceAll(groups[0].Hooks[0].Command, `"${CLAUDE_PLUGIN_ROOT}"`, plugin)
		info, err := os.Stat(command)
		if err != nil {
			t.Errorf("%s calls %q, which is not there: %v", event, groups[0].Hooks[0].Command, err)
			continue
		}
		if info.Mode()&0o111 == 0 {
			t.Errorf("%s calls %s, which cannot be run", event, command)
		}
	}
}

// The cairn server is named one way when the person declared it and another
// when the plugin did, and the hook must hear both — and nothing else.
func TestPlugin_hearsTheFocusFromEitherServer(t *testing.T) {
	var file hookFile
	readJSON(t, filepath.Join(plugin, "hooks", "hooks.json"), &file)
	matcher := regexp.MustCompile(file.Hooks["PostToolUse"][0].Matcher)
	for tool, want := range map[string]bool{
		"mcp__cairn__task_focus":                true,
		"mcp__cairn__task_unfocus":              true,
		"mcp__plugin_cairn_cairn__task_focus":   true,
		"mcp__plugin_cairn_cairn__task_unfocus": true,
		"mcp__cairn__task_list":                 false,
		"mcp__cairn__task_focus_list":           false,
		"Bash":                                  false,
	} {
		if matcher.MatchString(tool) != want {
			t.Errorf("the matcher hears %s: %v, want %v", tool, !want, want)
		}
	}
}

func TestPlugin_declaresAServerCairnAnswersTo(t *testing.T) {
	var file struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}
	readJSON(t, filepath.Join(plugin, ".mcp.json"), &file)
	server, ok := file.MCPServers["cairn"]
	if !ok || server.Command != "cairn" {
		t.Fatalf("the plugin declares %+v, want a server named cairn running cairn", file.MCPServers)
	}
	run := exec.Command(filepath.Join(cairnOnPath(t), "cairn"), append(server.Args, "--help")...)
	if out, err := run.CombinedOutput(); err != nil {
		t.Errorf("cairn refuses %v: %v\n%s", server.Args, err, out)
	}
}

func TestPlugin_isWhereTheMarketplaceSaysItIs(t *testing.T) {
	var market struct {
		Plugins []struct {
			Name   string `json:"name"`
			Source string `json:"source"`
		} `json:"plugins"`
	}
	readJSON(t, filepath.Join("..", "..", ".claude-plugin", "marketplace.json"), &market)
	if len(market.Plugins) != 1 {
		t.Fatalf("the marketplace lists %d plugins, want one", len(market.Plugins))
	}
	listed := market.Plugins[0]
	var manifest struct {
		Name string `json:"name"`
	}
	readJSON(t, filepath.Join("..", "..", listed.Source, ".claude-plugin", "plugin.json"), &manifest)
	if manifest.Name != listed.Name {
		t.Errorf("the marketplace lists %q, the plugin at %s is %q", listed.Name, listed.Source, manifest.Name)
	}
	if _, err := os.Stat(filepath.Join("..", "..", listed.Source, "skills", "cairn", "SKILL.md")); err != nil {
		t.Errorf("the plugin holds no cairn skill: %v", err)
	}
}
