package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/domain"
)

// openStore gives a store on a fresh file, closed when the test ends.
func openStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "cairn.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// keep adds a task titled title, failing the test if it cannot.
func keep(t *testing.T, s *Store, title string) domain.Task {
	t.Helper()
	task, err := domain.NewTask(title)
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	kept, err := s.Keep(task)
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	return kept
}

func TestStore_handsBackTheTaskItKept(t *testing.T) {
	s := openStore(t)
	kept := keep(t, s, "Write the spec")

	got, err := s.Find(kept.Number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if got.Title != "Write the spec" {
		t.Errorf("Title = %q, want the one it was given", got.Title)
	}
	if got.State != domain.Backlog || got.Priority != domain.Normal {
		t.Errorf("got %q/%q, want backlog/normal", got.State, got.Priority)
	}
}

func TestStore_numbersTheFirstTaskOne(t *testing.T) {
	s := openStore(t)
	if got := keep(t, s, "Write the spec").Number; got != 1 {
		t.Errorf("Number = %d, want 1", got)
	}
}

func TestStore_givesTheNextNumberToTheNextTask(t *testing.T) {
	s := openStore(t)
	keep(t, s, "First")
	if got := keep(t, s, "Second").Number; got != 2 {
		t.Errorf("Number = %d, want 2", got)
	}
}

func TestStore_neverHandsOutANumberFreedByADeletion(t *testing.T) {
	s := openStore(t)
	keep(t, s, "First")
	second := keep(t, s, "Second")
	if err := s.Delete(second.Number); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got := keep(t, s, "Third").Number; got != 3 {
		t.Errorf("Number = %d, want 3: a freed number is not handed out again", got)
	}
}

func TestStore_listsTasksInTheOrderTheyWereAdded(t *testing.T) {
	s := openStore(t)
	keep(t, s, "First")
	keep(t, s, "Second")

	all, err := s.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 2 || all[0].Title != "First" || all[1].Title != "Second" {
		t.Errorf("All = %v, want First then Second", all)
	}
}

func TestStore_saysThereIsNoSuchTask(t *testing.T) {
	s := openStore(t)
	if _, err := s.Find(99); err == nil {
		t.Error("Find(99) should fail: there is no such task")
	}
}

func TestStore_keepsTextExactlyAsItWasGiven(t *testing.T) {
	s := openStore(t)
	for _, title := range []string{"Écrire la spéc — enfin !", "🧗 climb", "90% done, right?", strings.Repeat("a", 5000)} {
		kept := keep(t, s, title)
		got, err := s.Find(kept.Number)
		if err != nil {
			t.Fatalf("Find: %v", err)
		}
		if got.Title != title {
			t.Errorf("Title round-trip lost data: got %q", got.Title)
		}
	}
}

func TestStore_keepsEveryAttributeItWasGiven(t *testing.T) {
	s := openStore(t)
	task, err := domain.NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	task.Description = "the one"
	task.State = domain.Todo
	task.Priority = domain.High
	task.Labels = []string{"spec", "team"}
	if task.Deadline, err = domain.ParseDeadline("2026-09-01"); err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}
	if task.Estimate, err = domain.ParseEstimate("90m"); err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}

	kept, err := s.Keep(task)
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	got, err := s.Find(kept.Number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if got.Description != "the one" || got.State != domain.Todo || got.Priority != domain.High {
		t.Errorf("got %+v, want the attributes it was given", got)
	}
	if len(got.Labels) != 2 || got.Labels[0] != "spec" || got.Labels[1] != "team" {
		t.Errorf("Labels = %v, want [spec team]", got.Labels)
	}
	if domain.FormatDay(got.Deadline) != "2026-09-01" {
		t.Errorf("Deadline = %v, want 2026-09-01", got.Deadline)
	}
	if domain.FormatDuration(got.Estimate) != "1h30m" {
		t.Errorf("Estimate = %v, want 1h30m", got.Estimate)
	}
}

func TestStore_replacesTheLabelsWhenATaskIsUpdated(t *testing.T) {
	s := openStore(t)
	task := keep(t, s, "Write the spec")
	task.Labels = []string{"spec", "draft"}
	if err := s.Update(task); err != nil {
		t.Fatalf("Update: %v", err)
	}
	task.Labels = []string{"team"}
	if err := s.Update(task); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := s.Find(task.Number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if len(got.Labels) != 1 || got.Labels[0] != "team" {
		t.Errorf("Labels = %v, want only [team]: what is given replaces what was there", got.Labels)
	}
}

func TestStore_saysWhenThereIsNothingToUpdate(t *testing.T) {
	s := openStore(t)
	task, _ := domain.NewTask("Nope")
	task.Number = 99
	if err := s.Update(task); err == nil {
		t.Error("Update on a task that does not exist should fail")
	}
}

func TestStore_saysWhenThereIsNothingToDelete(t *testing.T) {
	s := openStore(t)
	if err := s.Delete(99); err == nil {
		t.Error("Delete on a task that does not exist should fail")
	}
}

func TestStore_deletingOneLeavesTheOthersAlone(t *testing.T) {
	s := openStore(t)
	keep(t, s, "First")
	second := keep(t, s, "Second")
	keep(t, s, "Third")
	if err := s.Delete(second.Number); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	all, err := s.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 2 || all[0].Title != "First" || all[1].Title != "Third" {
		t.Errorf("All = %v, want First and Third", all)
	}
}

func TestOpen_createsTheFolderItNeeds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "deep", "down", "cairn.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	_ = s.Close()
}

func TestOpen_reportsSomethingThatIsNotACairnStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	if err := os.WriteFile(path, []byte("this is not a database"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := Open(path); err == nil {
		t.Error("Open should refuse a file that is not a cairn store")
	}
}

func TestOpen_reportsAPlaceItCannotWriteTo(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	// A file where a directory must go: the folder cannot be created.
	if _, err := Open(filepath.Join(path, "nested", "cairn.db")); err == nil {
		t.Error("Open should report a place it cannot create")
	}
}

// closed gives a store whose database is shut, so that every call has to report
// a failure it cannot do anything about. Error paths are behaviour too, and a
// store made only of happy paths is a store nobody has tested.
func closed(t *testing.T) *Store {
	t.Helper()
	s := openStore(t)
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return s
}

func TestStore_reportsAStoreItCanNoLongerReach(t *testing.T) {
	s := closed(t)
	task, _ := domain.NewTask("Write the spec")
	task.Number = 1

	if _, err := s.Keep(task); err == nil {
		t.Error("Keep on a closed store should fail")
	}
	if err := s.Update(task); err == nil {
		t.Error("Update on a closed store should fail")
	}
	if err := s.Delete(1); err == nil {
		t.Error("Delete on a closed store should fail")
	}
	if _, err := s.Find(1); err == nil {
		t.Error("Find on a closed store should fail")
	}
	if _, err := s.All(); err == nil {
		t.Error("All on a closed store should fail")
	}
}

func TestStore_reportsLabelsItCannotRead(t *testing.T) {
	s := openStore(t)
	kept := keep(t, s, "Write the spec")
	if _, err := s.db.Exec(`DROP TABLE labels`); err != nil {
		t.Fatalf("dropping labels: %v", err)
	}

	if _, err := s.Find(kept.Number); err == nil {
		t.Error("Find should report labels it cannot read")
	}
	if _, err := s.All(); err == nil {
		t.Error("All should report labels it cannot read")
	}
}

func TestStore_reportsLabelsItCannotKeep(t *testing.T) {
	s := openStore(t)
	task, _ := domain.NewTask("Write the spec")
	// The domain deduplicates; the store is handed whatever it is handed, and
	// two identical labels break the key that makes them a set.
	task.Labels = []string{"spec", "spec"}

	if _, err := s.Keep(task); err == nil {
		t.Error("Keep should report labels it cannot record")
	}
}

func TestStore_reportsLabelsItCannotReplace(t *testing.T) {
	s := openStore(t)
	task := keep(t, s, "Write the spec")
	task.Labels = []string{"spec", "spec"}

	if err := s.Update(task); err == nil {
		t.Error("Update should report labels it cannot record")
	}
}

func TestStore_reportsADeadlineItCannotRead(t *testing.T) {
	s := openStore(t)
	kept := keep(t, s, "Write the spec")
	if _, err := s.db.Exec(`UPDATE tasks SET deadline = 'soon' WHERE number = ?`, kept.Number); err != nil {
		t.Fatalf("writing a bad deadline: %v", err)
	}

	if _, err := s.Find(kept.Number); err == nil {
		t.Error("Find should report a deadline it cannot make sense of")
	}
	if _, err := s.All(); err == nil {
		t.Error("All should report a deadline it cannot make sense of")
	}
}

func TestStore_reportsLabelsItCannotClear(t *testing.T) {
	s := openStore(t)
	task := keep(t, s, "Write the spec")
	if _, err := s.db.Exec(`DROP TABLE labels`); err != nil {
		t.Fatalf("dropping labels: %v", err)
	}

	if err := s.Update(task); err == nil {
		t.Error("Update should report labels it cannot clear")
	}
	if err := s.Delete(task.Number); err == nil {
		t.Error("Delete should report labels it cannot clear")
	}
}

func TestStore_handsBackTheParentItWasGiven(t *testing.T) {
	s := openStore(t)
	under := keep(t, s, "Ship it")
	task, _ := domain.NewTask("Write the tests")
	task.Parent = under.Number

	kept, err := s.Keep(task)
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	got, err := s.Find(kept.Number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if got.Parent != under.Number {
		t.Errorf("Parent = %d, want %d", got.Parent, under.Number)
	}
}

func TestStore_namesWhatHangsUnderATask(t *testing.T) {
	s := openStore(t)
	under := keep(t, s, "Ship it")
	task, _ := domain.NewTask("Write the tests")
	task.Parent = under.Number
	if _, err := s.Keep(task); err != nil {
		t.Fatalf("Keep: %v", err)
	}
	keep(t, s, "Unrelated")

	children, err := s.Under(under.Number)
	if err != nil {
		t.Fatalf("Under: %v", err)
	}
	if len(children) != 1 || children[0].Title != "Write the tests" {
		t.Errorf("Under = %v, want only the sub-task", children)
	}
}

func TestStore_saysNothingHangsUnderALoneTask(t *testing.T) {
	s := openStore(t)
	lone := keep(t, s, "Ship it")
	children, err := s.Under(lone.Number)
	if err != nil {
		t.Fatalf("Under: %v", err)
	}
	if len(children) != 0 {
		t.Errorf("Under = %v, want none", children)
	}
}

func TestStore_reportsSubTasksItCannotRead(t *testing.T) {
	s := closed(t)
	if _, err := s.Under(1); err == nil {
		t.Error("Under on a closed store should fail")
	}
}

// Unusable answers for the store being out of reach, not for anything SQLite
// happens to complain about. Both halves matter: an error from another package
// is not it, and neither is a SQLite error of another kind.
func TestUnusable_tellsAStoreOutOfReachFromAnythingElse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	made, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := made.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Locked down before the connection is opened: SQLite keeps the descriptor
	// it was given, so a file made read-only afterwards is still written to.
	if err := os.Chmod(path, 0o444); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })

	db, err := sql.Open("sqlite", shared(path))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, refused := db.Exec(`INSERT INTO tasks (parent, objective, title, description,
		state, priority, deadline, estimate, time_spent)
		VALUES (0, 0, 'x', '', 'backlog', 'normal', '', 0, 0)`)
	if refused == nil {
		t.Fatal("writing to a read-only store was expected to fail")
	}
	if !Unusable(refused) {
		t.Errorf("Unusable(%v) = false, want true", refused)
	}
	if !Unusable(fmt.Errorf("keeping the task: %w", refused)) {
		t.Error("Unusable does not see through a wrapped error")
	}

	// A SQLite error, and not the one: asking for a table that is not there.
	var n int
	other := db.QueryRow(`SELECT count(*) FROM nowhere`).Scan(&n)
	if other == nil {
		t.Fatal("the query was expected to fail")
	}
	if Unusable(other) {
		t.Errorf("Unusable(%v) = true, want false: that is a request, not the store", other)
	}
	if Unusable(nil) || Unusable(ErrNoSuchTask) {
		t.Error("Unusable says yes to something that is not a store problem")
	}
}
