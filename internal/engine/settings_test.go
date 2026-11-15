package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/store/files"
)

func settingsIn(t *testing.T, folder, said string) {
	t.Helper()
	if err := os.MkdirAll(folder, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(folder, files.SettingsFile), []byte(said), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestKeepsSizes_readsTheSettingsBesideTheStore(t *testing.T) {
	h := homeHolding(t, "", map[string]string{files.StoreVariable: "~/s/cairn.db"})
	settingsIn(t, filepath.Join(h.Dir, "s"), "[window]\nkeep_sizes = true\n")
	if !KeepsSizes(h) {
		t.Error("keep_sizes = true beside the store does not keep the window's sizes")
	}
}

func TestLocate_refusesSettingsThatSayWhatBelongsToTheMachine(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	settingsIn(t, filepath.Join(h.Dir, "w"), "store = \"~/x/cairn.db\"\n")
	_, err := Locate(h)
	if !errors.Is(err, files.ErrMachineKey) || RefusalOf(err).Kind != KindRefused {
		t.Errorf("Locate = %v, want ErrMachineKey, refused rather than a store unusable", err)
	}
	if RefusalOf(err).Reason != "config-machine-key" {
		t.Errorf("the refusal is %+v", RefusalOf(err))
	}
	settingsIn(t, filepath.Join(h.Dir, "w"), "[focus]\nsession = \"\"\n")
	if _, err := Locate(h); RefusalOf(err).Reason != "config-empty" {
		t.Errorf("Locate = %v, want an empty string refused", err)
	}
}

func TestLocate_settingsThatCannotBeReadLeaveTheStoreUnusable(t *testing.T) {
	h := homeHolding(t, twoWorkspaces, nil, "w", "h")
	if err := os.MkdirAll(filepath.Join(h.Dir, "w", files.SettingsFile), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := Locate(h); RefusalOf(err).Kind != KindStore {
		t.Errorf("Locate = %v, want the store unusable", err)
	}
}

func TestSessionLength_isTheOneTheSettingsOfTheStoreSay(t *testing.T) {
	path := storeAt(t, "one")
	settingsIn(t, filepath.Dir(path), "[focus]\nsession = \"25m\"\n")
	e, err := OpenAt(Places{Store: path, Home: homeSaying(t, path, "[focus]\nsession = \"1h\"\n")})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = e.Close() }()
	if length, err := e.sessionLength(); err != nil || length.String() != "25m0s" {
		t.Errorf("sessionLength = %v, %v; want 25m", length, err)
	}
	settingsIn(t, filepath.Dir(path), "[pages\n")
	if length, err := e.sessionLength(); err != nil || length != 0 {
		t.Errorf("with settings it cannot read, sessionLength = %v, %v; want none, said where the places are read", length, err)
	}
}

func TestIgnored_saysWhatTheGlobalFileAndTheSettingsPassOver(t *testing.T) {
	h := homeHolding(t, "colour = 1\n", map[string]string{files.StoreVariable: "~/s/cairn.db"})
	settingsIn(t, filepath.Join(h.Dir, "s"), "[pages]\ncalendar = false\n")
	said := Ignored(h)
	if len(said) != 2 || !strings.Contains(said[1], filepath.Join(h.Dir, "s", files.SettingsFile)) {
		t.Errorf("Ignored = %q, want the global key and the settings' one, naming the settings", said)
	}
	if Ignored(Home{}) != nil {
		t.Error("with no home, something is said to be ignored")
	}
	settingsIn(t, filepath.Join(h.Dir, "s"), "store = \"~/x\"\n")
	if Ignored(h) != nil {
		t.Error("settings refused are said to pass keys over")
	}
}

func TestWorkspaceAdd_leavesSettingsAlreadyThere(t *testing.T) {
	h := homeHolding(t, "", nil)
	settingsIn(t, filepath.Join(h.Dir, "p"), "[pages]\ntime = false\n")
	if _, err := WorkspaceAdd.Run(workedOn{Home: h}, adding("play", "~/p")); err != nil {
		t.Fatal(err)
	}
	if read, _ := os.ReadFile(filepath.Join(h.Dir, "p", files.SettingsFile)); string(read) != "[pages]\ntime = false\n" {
		t.Errorf("the settings read %q", read)
	}
}
