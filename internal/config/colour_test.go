package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/config"
)

func TestWithSettings_theColourOfTheWorkspaceIsTheStoresOwn(t *testing.T) {
	read, _, err := settled(t, "", "[workspace]\ncolour = \"green\"\n")
	if err != nil || read.Colour() != "green" {
		t.Errorf("Colour = %q, %v; want green", read.Colour(), err)
	}
	read, _, err = settled(t, "", "[pages]\nobjectives = false\n")
	if err != nil || read.Colour() != "" {
		t.Errorf("Colour = %q, %v; want none when the settings say none", read.Colour(), err)
	}
}

func TestWithSettings_anEmptyColourIsRefused(t *testing.T) {
	if _, _, err := settled(t, "", "[workspace]\ncolour = \"\"\n"); !errors.Is(err, config.ErrEmpty) || !strings.Contains(err.Error(), "workspace.colour") {
		t.Errorf("WithSettings = %v; want %v naming workspace.colour", err, config.ErrEmpty)
	}
}

func TestRead_aColourInTheGlobalFileIsPassedOverSayingItBelongsToAStore(t *testing.T) {
	read, err := config.Read(writtenConfig(t, "[workspace]\ncolour = \"green\"\n"), silentEnv)
	if err != nil {
		t.Fatal(err)
	}
	said := strings.Join(read.Ignored(), "\n")
	if read.Colour() != "" || !strings.Contains(said, `"workspace.colour" is not a key this cairn knows`) || !strings.Contains(said, config.SettingsFile) {
		t.Errorf("Colour = %q, ignoring %q; want it passed over, saying it goes in %s", read.Colour(), said, config.SettingsFile)
	}
}

func TestColourIn_readsTheSettingsAloneWhateverElseTheySay(t *testing.T) {
	folder := t.TempDir()
	path := filepath.Join(folder, config.SettingsFile)
	if colour, err := config.ColourIn(path); err != nil || colour != "" {
		t.Errorf("ColourIn with no file = %q, %v; want none", colour, err)
	}
	if err := os.WriteFile(path, []byte("[pages]\ntime = \"off\"\n[workspace]\ncolour = \"peach\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if colour, err := config.ColourIn(path); err != nil || colour != "peach" {
		t.Errorf("ColourIn = %q, %v; want peach", colour, err)
	}
	if err := os.WriteFile(path, []byte("not toml ["), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.ColourIn(path); err == nil {
		t.Error("ColourIn of a file that is not TOML said nothing")
	}
}

func TestWithoutWorkspace_dropsItsLineAndNothingElse(t *testing.T) {
	was := "# mine\n[workspaces]\ndefault = \"work\"\nhome = \"~/h\"   # the house\nwork = \"~/w\"\n\n[pages]\nhome = true\n"
	said, err := config.WithoutWorkspace([]byte(was), "home")
	if want := "# mine\n[workspaces]\ndefault = \"work\"\nwork = \"~/w\"\n\n[pages]\nhome = true\n"; err != nil || string(said) != want {
		t.Errorf("WithoutWorkspace = %q, %v; want %q", said, err, want)
	}
}

func TestWithoutWorkspace_refusesWhatItCannotChange(t *testing.T) {
	for _, was := range []string{
		"",
		"[pages]\nobjectives = false\n",
		"workspaces = { home = \"~/h\" }\n",
		"[workspaces]\n\"home\" = \"~/h\"\n",
		"[workspaces]\nwork = \"~/w\"\n",
	} {
		if _, err := config.WithoutWorkspace([]byte(was), "home"); !errors.Is(err, config.ErrNotWritten) {
			t.Errorf("WithoutWorkspace(%q) = %v; want %v", was, err, config.ErrNotWritten)
		}
	}
}

func TestNameIn_readsTheNameAStoreGivesItself(t *testing.T) {
	path := filepath.Join(t.TempDir(), config.SettingsFile)
	if err := os.WriteFile(path, []byte("[workspace]\nname = \"fixtures\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if name, err := config.NameIn(path); err != nil || name != "fixtures" {
		t.Errorf("NameIn = %q, %v; want fixtures", name, err)
	}
}
