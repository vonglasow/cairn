package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The plugin is written beside the store unless a folder is named, and the
// answer says where, so the person knows what to hand Claude Code.
func TestPluginWrite_writesThePluginBesideTheStoreOrWhereTold(t *testing.T) {
	home := t.TempDir()
	store := filepath.Join(home, "cairn.db")
	answer, err := PluginWrite.Run(On{Path: store}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	beside := filepath.Join(home, "claude-plugin")
	if answer.Plugin == nil || *answer.Plugin != beside {
		t.Fatalf("the plugin is said to be at %v, want %s", answer.Plugin, beside)
	}
	if _, err := os.Stat(filepath.Join(beside, ".claude-plugin", "marketplace.json")); err != nil {
		t.Errorf("no marketplace beside the store: %v", err)
	}
	told := filepath.Join(home, "elsewhere")
	answer, err = PluginWrite.Run(On{Path: store}, Input{Given: map[string]string{"dir": told}})
	if err != nil {
		t.Fatal(err)
	}
	if *answer.Plugin != told {
		t.Errorf("the plugin is said to be at %s, want %s", *answer.Plugin, told)
	}
	if said := RenderText(PluginWrite, answer); !strings.Contains(said, "/plugin marketplace add "+told) ||
		!strings.Contains(said, "/plugin install cairn@cairn") {
		t.Errorf("the answer reads %q", said)
	}
	if shaped, ok := Of(PluginWrite, answer).(PluginWritten); !ok || shaped.Plugin != told {
		t.Errorf("a machine is told %#v", Of(PluginWrite, answer))
	}
}

// A folder that is not the plugin's is refused, and the refusal is said.
func TestPluginWrite_refusesAFolderItDidNotWrite(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mine"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := PluginWrite.Run(On{}, Input{Given: map[string]string{"dir": dir}}); err == nil {
		t.Error("a folder of somebody else's was written into")
	}
}
