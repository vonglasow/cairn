package engine

import (
	"cairn/internal/engine/app"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"cairn/internal/build"
	claudeplugin "cairn/plugin/claude"
)

func TestPluginWrite_writesThePluginBesideTheStoreOrWhereTheConfigurationSays(t *testing.T) {
	home := t.TempDir()
	store := filepath.Join(home, "cairn.db")
	answer, err := pluginWrite.Run(workedOn{Path: store}, Input{})
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
	answer, err = pluginWrite.Run(workedOn{Path: store, Plugin: told}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if *answer.Plugin != told {
		t.Errorf("the plugin is said to be at %s, want %s", *answer.Plugin, told)
	}
	if shaped, ok := Of(pluginWrite, answer).(app.PluginWritten); !ok || shaped.Plugin != told {
		t.Errorf("a machine is told %#v", Of(pluginWrite, answer))
	}
	if len(pluginWrite.Takes) != 0 {
		t.Errorf("plugin write takes %v: the folder is the configuration's to say, not a command's", pluginWrite.Takes)
	}
}

func TestPluginWrite_refusesAFolderItDidNotWrite(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mine"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := pluginWrite.Run(workedOn{Plugin: dir}, Input{}); err == nil {
		t.Error("a folder of somebody else's was written into")
	}
}

func TestPluginWrite_saysTheVersionOfThisCairn(t *testing.T) {
	store := filepath.Join(t.TempDir(), "cairn.db")
	if _, err := Do(pluginWrite, store, Input{}); err != nil {
		t.Fatal(err)
	}
	held, err := os.ReadFile(filepath.Join(app.PluginBeside(store), "cairn", ".claude-plugin", "plugin.json"))
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
	if _, err := os.Stat(app.PluginBeside(store)); err == nil {
		t.Fatal("a plugin nobody wrote was written")
	}
	if err := claudeplugin.Write(app.PluginBeside(store), "v1.0.0"); err != nil {
		t.Fatal(err)
	}
	if rewrote, err := KeepPluginCurrent(at, "v2.0.0"); err != nil || !rewrote {
		t.Fatalf("a plugin of v1.0.0 under cairn v2.0.0 was left: %v, %v", rewrote, err)
	}
	if rewrote, err := KeepPluginCurrent(at, "v2.0.0"); err != nil || rewrote {
		t.Fatalf("a plugin already current was written again: %v, %v", rewrote, err)
	}
}

func TestPluginShow_aReleaseSaysHowToInstallThePluginFromItsTag(t *testing.T) {
	answer, err := pluginShow.Run(workedOn{Home: Home{Version: "v0.12.0"}}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	install := []string{
		"/plugin marketplace remove cairn",
		"/plugin marketplace add vonglasow/cairn@v0.12.0",
		"/plugin install cairn@cairn",
	}
	if shaped, ok := Of(pluginShow, answer).(app.PluginShown); !ok || shaped.Version != "v0.12.0" || !slices.Equal(shaped.Install, install) {
		t.Errorf("a machine is told %#v", Of(pluginShow, answer))
	}
}

func TestPluginShow_aBuildThatIsNoVersionSaysToWriteThePluginItCarries(t *testing.T) {
	answer, err := pluginShow.Run(workedOn{Home: Home{Version: "v0.11.0-79-g616ccfc"}}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	if shaped := Of(pluginShow, answer).(app.PluginShown); shaped.Version != "v0.11.0-79-g616ccfc" || len(shaped.Install) != 0 {
		t.Errorf("a machine is told %#v", shaped)
	}
}
