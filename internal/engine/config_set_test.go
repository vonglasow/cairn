package engine

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"cairn/internal/config"
)

const homeIsDefault = "[workspaces]\ndefault = \"home\"\nhome = \"~/h\"\n"

func settingsOf(t *testing.T, h Home, folder, said string) string {
	t.Helper()
	path := filepath.Join(h.Dir, folder, config.SettingsFile)
	if err := os.WriteFile(path, []byte(said), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func fileRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	return string(data)
}

func setting(h Home, op Operation, given map[string]string) (Answer, error) {
	return DoFrom(op, h, Input{Given: given}, Acting{})
}

func TestConfigSet_writesTheSettingsOfTheWorkspaceInUse(t *testing.T) {
	h := homeHolding(t, homeIsDefault, nil, "h")
	path := settingsOf(t, h, "h", "[pages]\n# none at home\nobjectives = false\n")
	got, err := setting(h, ConfigSet, map[string]string{"key": "pages.time", "value": "hidden"})
	if err != nil {
		t.Fatal(err)
	}
	if want := "[pages]\n# none at home\nobjectives = false\ntime = \"hidden\"\n"; fileRead(t, path) != want {
		t.Errorf("the settings became\n%s", fileRead(t, path))
	}
	if c := got.ConfigChanged; c == nil || c.File != path || c.Global || c.Value != "hidden" || c.From != "set in the store's settings" {
		t.Errorf("the answer is %+v", got.ConfigChanged)
	}
	if text := RenderText(ConfigSet, got); !strings.Contains(text, `pages.time = "hidden"`) || !strings.Contains(text, path) {
		t.Errorf("it says %q", text)
	}
}

func TestConfigSet_globalWritesTheGlobalFile(t *testing.T) {
	h := homeHolding(t, homeIsDefault, nil, "h")
	got, err := setting(h, ConfigSet, map[string]string{"key": "focus.session", "value": "45m", "global": "true"})
	if err != nil {
		t.Fatal(err)
	}
	if want := homeIsDefault + "\n[focus]\nsession = \"45m\"\n"; fileRead(t, config.FileIn(h.Dir)) != want {
		t.Errorf("the file became\n%s", fileRead(t, config.FileIn(h.Dir)))
	}
	if !got.ConfigChanged.Global || got.ConfigChanged.From != "from the global configuration" {
		t.Errorf("the answer is %+v", got.ConfigChanged)
	}
}

func TestConfigSet_aLooseStoreWithNoSettingsIsSetInTheGlobalFile(t *testing.T) {
	h := homeHolding(t, "", nil)
	if _, err := setting(h, ConfigSet, map[string]string{"key": "pages.time", "value": "off"}); err != nil {
		t.Fatal(err)
	}
	if got := fileRead(t, config.FileIn(h.Dir)); got != "[pages]\ntime = \"off\"\n" {
		t.Errorf("the file became %q", got)
	}
}

func TestConfigSet_aLooseStoreWithSettingsBesideItIsSetThere(t *testing.T) {
	h := homeHolding(t, "", map[string]string{config.StoreVariable: "~/l/cairn.db"}, "l")
	path := settingsOf(t, h, "l", "")
	if _, err := setting(h, ConfigSet, map[string]string{"key": "time.weekends", "value": "hidden"}); err != nil {
		t.Fatal(err)
	}
	if got := fileRead(t, path); got != "[time]\nweekends = \"hidden\"\n" {
		t.Errorf("the settings became %q", got)
	}
}

func TestConfigSet_aKeyOfTheMachineGoesGlobalAndAKeyOfAStoreDoesNot(t *testing.T) {
	h := homeHolding(t, homeIsDefault, nil, "h")
	if _, err := setting(h, ConfigSet, map[string]string{"key": "plugin", "value": "~/p"}); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(fileRead(t, config.FileIn(h.Dir)), "plugin = \"~/p\"\n\n[workspaces]") {
		t.Errorf("the file became\n%s", fileRead(t, config.FileIn(h.Dir)))
	}
	_, err := setting(h, ConfigSet, map[string]string{"key": "workspace.colour", "value": "green", "global": "true"})
	if err == nil || !strings.Contains(err.Error(), "workspace.colour belongs to a store") || RefusalOf(err).Kind != KindRefused {
		t.Errorf("the colour is set globally: %v", err)
	}
	if _, err := setting(h, ConfigSet, map[string]string{"key": "workspace.colour", "value": "green"}); err != nil {
		t.Fatal(err)
	}
	if got := fileRead(t, filepath.Join(h.Dir, "h", config.SettingsFile)); got != "[workspace]\ncolour = \"green\"\n" {
		t.Errorf("the settings became %q", got)
	}
}

func TestConfigSet_refusesWhatTheKeyDoesNotTakeAndWritesNothing(t *testing.T) {
	h := homeHolding(t, homeIsDefault, nil, "h")
	path := settingsOf(t, h, "h", "[pages]\nobjectives = false\n")
	for _, c := range []struct{ key, value, said string }{
		{"pages.time", "hiden", `"hiden" is not a word pages.time takes (on, hidden, off)`},
		{"focus.session", "1 hour", `focus.session: "1 hour" is not a duration — write 1h, 45m, 1h30m`},
		{"focus.session", "-1h", `focus.session: "-1h" is not a duration`},
		{"window.theme", "neon", `"neon" is not a scheme cairn sees`},
		{"workspace.colour", "puce", "puce"},
		{"pages.tiem", "on", `"pages.tiem" is not a key cairn sets`},
		{"plugin", "p", "must be absolute"},
		{"pages.time", "", "pages.time"},
	} {
		_, err := setting(h, ConfigSet, map[string]string{"key": c.key, "value": c.value})
		if err == nil || !strings.Contains(err.Error(), c.said) || RefusalOf(err).Kind != KindRefused {
			t.Errorf("%s = %q is refused with %v (%v), want %q", c.key, c.value, err, RefusalOf(err), c.said)
		}
	}
	if fileRead(t, path) != "[pages]\nobjectives = false\n" {
		t.Errorf("the settings became %q", fileRead(t, path))
	}
}

func TestConfigSet_aLabelEmptyIsNoLabelHereOnlyInTheSettings(t *testing.T) {
	h := homeHolding(t, homeIsDefault, nil, "h")
	if _, err := setting(h, ConfigSet, map[string]string{"key": "approach.label", "value": ""}); err != nil {
		t.Fatal(err)
	}
	if got := fileRead(t, filepath.Join(h.Dir, "h", config.SettingsFile)); got != "[approach]\nlabel = \"\"\n" {
		t.Errorf("the settings became %q", got)
	}
	if _, err := setting(h, ConfigSet, map[string]string{"key": "approach.label", "value": "", "global": "true"}); err == nil {
		t.Error("the global label is said empty")
	}
}

func TestConfigSet_aFileThatDoesNotReadIsNotWrittenInto(t *testing.T) {
	h := homeHolding(t, homeIsDefault, nil, "h")
	path := settingsOf(t, h, "h", "[pages]\ntime = \"hiden\"\n")
	_, err := setting(h, ConfigSet, map[string]string{"key": "pages.objectives", "value": "false"})
	if err == nil || !strings.Contains(err.Error(), path) {
		t.Errorf("a file that does not read was written into: %v", err)
	}
	if fileRead(t, path) != "[pages]\ntime = \"hiden\"\n" {
		t.Errorf("the settings became %q", fileRead(t, path))
	}
	_, err = setting(h, ConfigSet, map[string]string{"key": "focus.session", "value": "1h", "global": "true"})
	if err != nil {
		t.Errorf("the global file is not written while the settings do not read: %v", err)
	}
}

func TestConfigSet_leavesNoFileBehindButTheOneWritten(t *testing.T) {
	h := homeHolding(t, homeIsDefault, nil, "h")
	if _, err := setting(h, ConfigSet, map[string]string{"key": "pages.time", "value": "off"}); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(filepath.Join(h.Dir, "h"))
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if !slices.Equal(names, []string{config.SettingsFile}) {
		t.Errorf("the folder holds %v", names)
	}
	if info, err := os.Stat(filepath.Join(h.Dir, "h", config.SettingsFile)); err != nil || info.Mode().Perm() != 0o600 {
		t.Errorf("the settings are %v, %v", info, err)
	}
}

func TestConfigUnset_takesTheLineOutAndSaysWhatAnswersNow(t *testing.T) {
	h := homeHolding(t, homeIsDefault+"\n[pages]\nobjectives = false\n", nil, "h")
	path := settingsOf(t, h, "h", "[pages]\n# no objectives at home\nobjectives = true\n")
	got, err := setting(h, ConfigUnset, map[string]string{"key": "pages.objectives"})
	if err != nil {
		t.Fatal(err)
	}
	if fileRead(t, path) != "[pages]\n# no objectives at home\n" {
		t.Errorf("the settings became %q", fileRead(t, path))
	}
	if c := got.ConfigChanged; c.Value != "false" || c.From != "from the global configuration" || c.Set {
		t.Errorf("the answer is %+v", c)
	}
	if text := RenderText(ConfigUnset, got); !strings.Contains(text, "pages.objectives is no longer said in "+path) || !strings.Contains(text, "false, from the global configuration") {
		t.Errorf("it says %q", text)
	}
	_, err = setting(h, ConfigUnset, map[string]string{"key": "pages.objectives"})
	if err == nil || !strings.Contains(err.Error(), "pages.objectives is not said in "+path) || RefusalOf(err).Kind != KindRefused {
		t.Errorf("a key not said is taken out: %v", err)
	}
}

func TestConfigSetAndUnsetAreAPersonsOnly(t *testing.T) {
	for _, op := range []Operation{ConfigSet, ConfigUnset} {
		if slices.ContainsFunc(All(), func(o Operation) bool { return o.Name() == op.Name() }) {
			t.Errorf("%s is offered to an agent", op.Name())
		}
		if !slices.ContainsFunc(ForAPersonOnly(), func(o Operation) bool { return o.Name() == op.Name() }) {
			t.Errorf("%s is not offered to a person", op.Name())
		}
	}
}
