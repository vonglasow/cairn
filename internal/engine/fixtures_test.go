package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/store/files"
)

func countTasks(t *testing.T, store string) int {
	t.Helper()
	a, err := Do(TaskList, store, Input{Given: map[string]string{"with-archived": "true"}})
	if err != nil {
		t.Fatal(err)
	}
	return len(a.Tasks)
}

func TestMakeFixtures_fillsAFolderItMakesAndKeepsItAfter(t *testing.T) {
	folder := filepath.Join(t.TempDir(), "fixtures")
	made, err := MakeFixtures(folder, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if made.Store != filepath.Join(folder, "cairn.db") || made.Notes != filepath.Join(folder, "notes") || !made.Made {
		t.Errorf("made %+v", made)
	}
	if settings, err := os.ReadFile(filepath.Join(folder, "settings.toml")); err != nil || !strings.Contains(string(settings), "[series]\non = true") {
		t.Errorf("the settings of the fixtures read %q (%v), want recurrence on", settings, err)
	}
	if colour, err := files.ColourIn(filepath.Join(folder, "settings.toml")); err != nil || colour == "" {
		t.Errorf("the settings of the fixtures say the colour %q (%v), want one", colour, err)
	}
	if name, err := files.NameIn(filepath.Join(folder, "settings.toml")); err != nil || name != "fixtures" {
		t.Errorf("the settings of the fixtures name them %q (%v), want fixtures", name, err)
	}
	first := countTasks(t, made.Store)
	if first < 10 {
		t.Errorf("the fixtures hold %d tasks", first)
	}
	if _, err := Do(TaskNew, made.Store, given("title", "Mine")); err != nil {
		t.Fatal(err)
	}
	again, err := MakeFixtures(folder, "", false)
	if err != nil || again.Made || countTasks(t, again.Store) != first+1 {
		t.Errorf("made again %+v, %v: want the folder opened as it was", again, err)
	}
	fresh, err := MakeFixtures(folder, "", true)
	if err != nil || !fresh.Made || countTasks(t, fresh.Store) != first {
		t.Errorf("made fresh %+v, %v: want it made from nothing", fresh, err)
	}
}

func TestMakeFixtures_leavesAFolderItDidNotMakeAlone(t *testing.T) {
	folder := t.TempDir()
	mine := filepath.Join(folder, "mine.txt")
	if err := os.WriteFile(mine, []byte("mine"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, fresh := range []bool{false, true} {
		if _, err := MakeFixtures(folder, "", fresh); !errors.Is(err, errNotMadeByFixtures) {
			t.Errorf("fresh %v: err = %v, want ErrNotMadeByFixtures", fresh, err)
		}
	}
	if _, err := os.Stat(mine); err != nil {
		t.Errorf("the file went: %v", err)
	}
}

func TestFixturesFile_isAnExportImportCanRead(t *testing.T) {
	from := filepath.Join(t.TempDir(), "fixtures.json")
	if err := os.WriteFile(from, FixturesFile(), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readExport(from); err != nil {
		t.Errorf("the fixtures file cannot be read as an export: %v", err)
	}
}
