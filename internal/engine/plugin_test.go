package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/build"
	claudeplugin "cairn/plugin/claude"
)

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

func TestPluginWrite_refusesAFolderItDidNotWrite(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mine"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := PluginWrite.Run(On{}, Input{Given: map[string]string{"dir": dir}}); err == nil {
		t.Error("a folder of somebody else's was written into")
	}
}

func TestPluginWrite_saysTheVersionOfThisCairn(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "plugin")
	if _, err := Do(PluginWrite, filepath.Join(t.TempDir(), "cairn.db"), Input{Given: map[string]string{"dir": dir}}); err != nil {
		t.Fatal(err)
	}
	held, err := os.ReadFile(filepath.Join(dir, "cairn", ".claude-plugin", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	if want := `"version": "` + build.Read().Version + `"`; !strings.Contains(string(held), want) {
		t.Errorf("the plugin says:\n%s\nwant %s", held, want)
	}
}

func TestKeepPluginCurrent_rewritesOnlyAPluginAnotherCairnWrote(t *testing.T) {
	store := filepath.Join(t.TempDir(), "cairn.db")
	at := Places{Store: store}
	if rewrote, err := KeepPluginCurrent(at, "v2.0.0"); err != nil || rewrote {
		t.Fatalf("with no plugin written, it rewrote: %v, %v", rewrote, err)
	}
	if _, err := os.Stat(PluginBeside(store)); err == nil {
		t.Fatal("a plugin nobody wrote was written")
	}
	if err := claudeplugin.Write(PluginBeside(store), "v1.0.0"); err != nil {
		t.Fatal(err)
	}
	if rewrote, err := KeepPluginCurrent(at, "v2.0.0"); err != nil || !rewrote {
		t.Fatalf("a plugin of v1.0.0 under cairn v2.0.0 was left: %v, %v", rewrote, err)
	}
	if rewrote, err := KeepPluginCurrent(at, "v2.0.0"); err != nil || rewrote {
		t.Fatalf("a plugin already current was written again: %v, %v", rewrote, err)
	}
}
