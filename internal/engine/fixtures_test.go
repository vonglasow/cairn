package engine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
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
		if _, err := MakeFixtures(folder, "", fresh); !errors.Is(err, ErrNotMadeByFixtures) {
			t.Errorf("fresh %v: err = %v, want ErrNotMadeByFixtures", fresh, err)
		}
	}
	if _, err := os.Stat(mine); err != nil {
		t.Errorf("the file went: %v", err)
	}
}

func TestMakeFixtures_fillsAnEmptyFolderAndReadsAnotherFile(t *testing.T) {
	from := filepath.Join(t.TempDir(), "mine.json")
	if err := os.WriteFile(from, []byte(`{"cairn": "tasks", "version": 4, "objectives": [], "tasks": [{"number": 1, "title": "Only", "state": "todo", "priority": "normal"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	made, err := MakeFixtures(t.TempDir(), from, false)
	if err != nil || countTasks(t, made.Store) != 1 {
		t.Errorf("made %+v, %v: want the one task of the file", made, err)
	}
}

func TestFixturesFolder_isInTheTemporaryFolder(t *testing.T) {
	if got := FixturesFolder(); got != filepath.Join(os.TempDir(), "cairn-fixtures") {
		t.Errorf("FixturesFolder = %q", got)
	}
}
