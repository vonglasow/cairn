package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"

	"cairn/internal/config"
)

func settled(t *testing.T, global, settings string) (config.Places, string, error) {
	t.Helper()
	home := writtenConfig(t, global)
	read, err := config.Read(home, silentEnv)
	if err != nil {
		t.Fatalf("the global configuration is refused: %v", err)
	}
	path := filepath.Join(home, "work", config.SettingsFile)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(settings), 0o600); err != nil {
		t.Fatal(err)
	}
	said, err := read.WithSettings(path)
	return said, path, err
}

func TestSettingsBeside_isTheFileInTheFolderOfTheStore(t *testing.T) {
	if got, want := config.SettingsBeside("/w/cairn.db"), filepath.Join("/w", "settings.toml"); got != want {
		t.Errorf("SettingsBeside = %q, want %q", got, want)
	}
}

func TestWithSettings_aKeyTheyWriteWinsOverTheGlobalOne(t *testing.T) {
	read, _, err := settled(t, "[focus]\nsession = \"1h\"\n[approach]\nlabel = \"discuss\"\n", "[focus]\nsession = \"25m\"\n[approach]\nlabel = \"agree\"\n")
	if err != nil {
		t.Fatal(err)
	}
	if read.FocusSession() != "25m" || read.ApproachLabel() != "agree" {
		t.Errorf("session %q, label %q; want 25m and agree", read.FocusSession(), read.ApproachLabel())
	}
}

func TestWithSettings_aKeyTheyDoNotWriteIsTheGlobalOne(t *testing.T) {
	read, _, err := settled(t, "[time]\ndeleted = \"taken\"\n[archive]\nstart = 2026-10-21\ndays = 7\n", "[time]\nweekends = \"hidden\"\n[archive]\ndays = 3\n")
	if err != nil {
		t.Fatal(err)
	}
	if read.TimeDeleted() != config.TimeTaken || read.TimeWeekends() != config.WeekendsHidden {
		t.Errorf("deleted %q, weekends %q; want taken and hidden", read.TimeDeleted(), read.TimeWeekends())
	}
	if start, days, on := read.Archive(); start != "2026-10-21" || days != 3 || !on {
		t.Errorf("Archive = %q, %d, %v; want 2026-10-21, 3, true", start, days, on)
	}
}

func TestWithSettings_eachPageIsSaidOnItsOwn(t *testing.T) {
	read, _, err := settled(t, "[pages]\nobjectives = false\ntime = false\n", "[pages]\nobjectives = true\ntasks = false\n")
	if err != nil {
		t.Fatal(err)
	}
	if got := read.Pages(); got != [3]bool{false, true, false} {
		t.Errorf("Pages = %v, want tasks off, objectives on, time off", got)
	}
}

func TestWithSettings_theWindowKeepsItsSizesWhenTheySaySo(t *testing.T) {
	read, _, err := settled(t, "", "[window]\nkeep_sizes = true\n")
	if err != nil {
		t.Fatal(err)
	}
	if !read.KeepSizes() {
		t.Error("keep_sizes = true in the settings does not keep the window's sizes")
	}
}

func TestWithSettings_theSchemeTheySayWinsOverTheGlobalOne(t *testing.T) {
	read, _, err := settled(t, "[window]\ntheme = \"nord\"\nkeep_sizes = true\n", "[window]\ntheme = \"dracula\"\n")
	if err != nil {
		t.Fatal(err)
	}
	if read.Theme() != "dracula" || !read.KeepSizes() {
		t.Errorf("theme %q, keep_sizes %t; want dracula, and the global true", read.Theme(), read.KeepSizes())
	}
}

func TestWithSettings_leavesTheGlobalReadingAsItWas(t *testing.T) {
	home := writtenConfig(t, "[focus]\nsession = \"1h\"\n")
	read, err := config.Read(home, silentEnv)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, config.SettingsFile)
	if err := os.WriteFile(path, []byte("[focus]\nsession = \"25m\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := read.WithSettings(path); err != nil {
		t.Fatal(err)
	}
	if read.FocusSession() != "1h" {
		t.Errorf("the global reading's session became %q", read.FocusSession())
	}
	again, err := read.WithSettings(filepath.Join(home, "nowhere", config.SettingsFile))
	if err != nil || again.FocusSession() != "1h" {
		t.Errorf("after the settings were read once, none beside another store says %q, %v; want 1h", again.FocusSession(), err)
	}
}

func TestWithSettings_noFileIsTheGlobalConfigurationWhole(t *testing.T) {
	home := writtenConfig(t, "[focus]\nsession = \"1h\"\n")
	read, err := config.Read(home, silentEnv)
	if err != nil {
		t.Fatal(err)
	}
	said, err := read.WithSettings(filepath.Join(home, "nowhere", config.SettingsFile))
	if err != nil {
		t.Fatal(err)
	}
	if said.FocusSession() != "1h" || said.Settings() != "" || said.WrittenHere() != nil {
		t.Errorf("session %q, settings %q, written here %v; want 1h, none, none", said.FocusSession(), said.Settings(), said.WrittenHere())
	}
}

func TestWithSettings_aFolderInPlaceOfTheFileIsNotRead(t *testing.T) {
	home := writtenConfig(t, "")
	read, err := config.Read(home, silentEnv)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, config.SettingsFile)
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := read.WithSettings(path); err == nil {
		t.Error("a folder where the settings should be is read as settings")
	}
}

func TestWithSettings_saysWhichKeysTheyWrite(t *testing.T) {
	read, path, err := settled(t, "[focus]\nsession = \"1h\"\n", "[time]\nweekends = \"hidden\"\n")
	if err != nil {
		t.Fatal(err)
	}
	if read.Settings() != path {
		t.Errorf("Settings = %q, want %q", read.Settings(), path)
	}
	if !slices.Equal(read.WrittenHere(), []string{"time.weekends"}) || !slices.Equal(read.Written(), []string{"focus.session"}) {
		t.Errorf("written here %v, written %v; want time.weekends here and focus.session in the global file", read.WrittenHere(), read.Written())
	}
}

func TestWithSettings_refusesAKeyOfTheMachineNamingWhereItGoes(t *testing.T) {
	for _, c := range []struct{ said, key string }{
		{"store = \"/elsewhere/cairn.db\"\n", "store"},
		{"notes = \"/vault\"\n", "notes"},
		{"plugin = \"/plugin\"\n", "plugin"},
		{"[workspaces]\nplay = \"/play\"\n", "workspaces"},
	} {
		home := writtenConfig(t, "")
		read, err := config.Read(home, silentEnv)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(home, config.SettingsFile)
		if err := os.WriteFile(path, []byte(c.said), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err = read.WithSettings(path)
		if !errors.Is(err, config.ErrMachineKey) {
			t.Errorf("%s: WithSettings = %v, want ErrMachineKey", c.key, err)
			continue
		}
		for _, named := range []string{`"` + c.key + `"`, path, config.FileIn(home)} {
			if !strings.Contains(err.Error(), named) {
				t.Errorf("%s: %q does not name %s", c.key, err, named)
			}
		}
	}
}

func TestWithSettings_anEmptyApproachLabelIsNoLabelHere(t *testing.T) {
	read, _, err := settled(t, "[approach]\nlabel = \"discuss\"\n", "[approach]\nlabel = \"\"\n")
	if err != nil {
		t.Fatal(err)
	}
	if read.ApproachLabel() != "" {
		t.Errorf("ApproachLabel = %q, want none", read.ApproachLabel())
	}
}

func TestWithSettings_refusesAnyOtherEmptyStringNamingItsKey(t *testing.T) {
	for _, c := range []struct{ said, key string }{
		{"[focus]\nsession = \"\"\n", "focus.session"},
		{"[time]\ndeleted = \"\"\n", "time.deleted"},
		{"[time]\nweekends = \"\"\n", "time.weekends"},
	} {
		_, path, err := settled(t, "", c.said)
		if !errors.Is(err, config.ErrEmpty) || !strings.Contains(err.Error(), c.key) || !strings.Contains(err.Error(), path) {
			t.Errorf("%s: WithSettings = %v, want ErrEmpty naming the key and the file", c.key, err)
		}
	}
}

func TestWithSettings_aKeyItDoesNotKnowIsPassedOverAndSaid(t *testing.T) {
	read, path, err := settled(t, "[pages]\ncolour = true\n", "[pages]\ncalendar = false\ntime = false\n")
	if err != nil {
		t.Fatal(err)
	}
	ignored := read.Ignored()
	if len(ignored) != 2 || !strings.Contains(ignored[1], `"pages.calendar" is not a key this cairn knows`) || !strings.HasSuffix(ignored[1], "in "+path) {
		t.Errorf("Ignored = %q, want the global key then pages.calendar, in %s", ignored, path)
	}
	if read.Pages()[2] {
		t.Error("the rest of the settings was not read")
	}
}

func TestWithSettings_refusesWhatItCannotRead(t *testing.T) {
	for _, c := range []struct {
		said string
		want error
	}{
		{"[focus\n", config.ErrNotTOML},
		{"[time]\ndeleted = \"gone\"\n", config.ErrTime},
		{"[archive]\ndays = 0\n", config.ErrArchive},
		{"[window]\nkeep_sizes = \"yes\"\n", config.ErrKeepSizes},
	} {
		_, path, err := settled(t, "", c.said)
		if !errors.Is(err, c.want) || !strings.Contains(err.Error(), path) {
			t.Errorf("%q: WithSettings = %v, want %v naming %s", c.said, err, c.want, path)
		}
	}
}

func TestSettingsHeader_saysNothingAndWhy(t *testing.T) {
	var said map[string]any
	if err := toml.Unmarshal([]byte(config.SettingsHeader), &said); err != nil || len(said) > 0 {
		t.Errorf("the header reads as %v, %v; want nothing", said, err)
	}
	if !strings.Contains(config.SettingsHeader, "until\n# this file says otherwise") {
		t.Errorf("the header does not say everything comes from the global configuration until it says otherwise:\n%s", config.SettingsHeader)
	}
}

func TestWithSettings_theStoreTurnsTheTimeOffWhileTheGlobalHasItOn(t *testing.T) {
	read, _, err := settled(t, "[pages]\ntime = \"on\"\n", "[pages]\ntime = \"off\"\n")
	if err != nil {
		t.Fatal(err)
	}
	if got := read.TimeMode(); got != config.TimeOff {
		t.Errorf("TimeMode = %q, want off", got)
	}
}
