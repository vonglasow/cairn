package claudeplugin

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWrite_makesAMarketplaceOfThePlugin(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "plugins")
	if err := Write(dir, "1.2.3"); err != nil {
		t.Fatal(err)
	}
	var market struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Owner       struct {
			Name string `json:"name"`
		} `json:"owner"`
		Plugins []struct {
			Name, Source, Description string
		} `json:"plugins"`
	}
	said, err := os.ReadFile(filepath.Join(dir, ".claude-plugin", "marketplace.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(said, &market); err != nil {
		t.Fatal(err)
	}
	if market.Name != "cairn" || market.Description == "" || market.Owner.Name == "" || len(market.Plugins) != 1 {
		t.Fatalf("the marketplace reads %+v", market)
	}
	p := market.Plugins[0]
	if p.Name != "cairn" || p.Source != "./cairn" || p.Description == "" {
		t.Errorf("the plugin is offered as %+v", p)
	}
	if _, err := os.Stat(filepath.Join(dir, "cairn", ".claude-plugin", "plugin.json")); err != nil {
		t.Errorf("the plugin is not where the marketplace says: %v", err)
	}
	info, err := os.Stat(filepath.Join(dir, "cairn", "hooks", "cairn-session.sh"))
	if err != nil || info.Mode().Perm() != 0o755 {
		t.Errorf("the hook is written %v, %v", info, err)
	}
	if info, err := os.Stat(filepath.Join(dir, "cairn", "hooks", "hooks.json")); err != nil || info.Mode().Perm() != 0o644 {
		t.Errorf("a file that is not run is written %v, %v", info, err)
	}
}

func TestWrite_refusesAFolderItDidNotWrite(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "precious.txt"), []byte("mine"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Write(dir, "1.2.3"); !errors.Is(err, ErrNotItsFolder) {
		t.Fatalf("writing into a folder of somebody else's = %v, want %v", err, ErrNotItsFolder)
	}
	if _, err := os.Stat(filepath.Join(dir, "precious.txt")); err != nil {
		t.Errorf("what the folder held is gone: %v", err)
	}
}

func TestWrite_writesAnEmptyFolderOrItsOwnAgain(t *testing.T) {
	dir := t.TempDir()
	if err := Write(dir, "1.2.3"); err != nil {
		t.Fatalf("an empty folder: %v", err)
	}
	stale := filepath.Join(dir, "cairn", "skills", "gone.md")
	if err := os.WriteFile(stale, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Write(dir, "1.2.3"); err != nil {
		t.Fatalf("its own folder again: %v", err)
	}
	if _, err := os.Stat(stale); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("what an older cairn wrote is still there: %v", err)
	}
}

func TestWrite_saysWhatCannotBeWritten(t *testing.T) {
	file := filepath.Join(t.TempDir(), "a file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Write(filepath.Join(file, "under"), "1.2.3"); err == nil {
		t.Error("writing under a file said nothing went wrong")
	}
}

func TestWrite_refusesAMarketplaceThatIsNotItsOwn(t *testing.T) {
	for _, held := range []string{`{"name": "another"}`, `not json`} {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, ".claude-plugin"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, ".claude-plugin", "marketplace.json"), []byte(held), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := Write(dir, "1.2.3"); !errors.Is(err, ErrNotItsFolder) {
			t.Errorf("a marketplace holding %s = %v, want %v", held, err, ErrNotItsFolder)
		}
	}
}

func TestWrite_saysWhichStepCouldNotBeWritten(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permissions do not stop the superuser")
	}
	locked := func(t *testing.T, path string) {
		t.Helper()
		if err := os.Chmod(path, 0o500); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(path, 0o755) })
	}

	clearing := t.TempDir()
	if err := Write(clearing, "1.2.3"); err != nil {
		t.Fatal(err)
	}
	locked(t, filepath.Join(clearing, Name, "hooks"))
	if err := Write(clearing, "1.2.3"); err == nil {
		t.Error("a plugin that could not be cleared said nothing went wrong")
	}

	market := t.TempDir()
	if err := Write(market, "1.2.3"); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(market, ".claude-plugin", "marketplace.json"), 0o400); err != nil {
		t.Fatal(err)
	}
	locked(t, filepath.Join(market, ".claude-plugin"))
	if err := Write(market, "1.2.3"); err == nil {
		t.Error("a marketplace that could not be written said nothing went wrong")
	}

	folders := t.TempDir()
	if err := Write(folders, "1.2.3"); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(folders, Name)); err != nil {
		t.Fatal(err)
	}
	locked(t, folders)
	if err := Write(folders, "1.2.3"); err == nil {
		t.Error("a plugin whose folders could not be made said nothing went wrong")
	}
}

func TestWrite_saysTheVersionOfTheCairnThatWroteIt(t *testing.T) {
	dir := t.TempDir()
	if err := Write(dir, "4.5.6"); err != nil {
		t.Fatal(err)
	}
	var plugin struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	read(t, filepath.Join(dir, Name, ".claude-plugin", "plugin.json"), &plugin)
	if plugin.Version != "4.5.6" || plugin.Name != Name {
		t.Errorf("the plugin written says %+v", plugin)
	}
	var market struct {
		Plugins []struct {
			Version string `json:"version"`
		} `json:"plugins"`
	}
	read(t, filepath.Join(dir, ".claude-plugin", "marketplace.json"), &market)
	if len(market.Plugins) != 1 || market.Plugins[0].Version != "4.5.6" {
		t.Errorf("the marketplace says %+v", market)
	}
}

func read(t *testing.T, path string, into any) {
	t.Helper()
	held, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(held, into); err != nil {
		t.Fatal(err)
	}
}
