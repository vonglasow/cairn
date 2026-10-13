package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/task"
)

func openStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "cairn.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func keep(t *testing.T, s *Store, title string) task.Task {
	t.Helper()
	built, err := task.NewTask(title)
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	kept, err := s.Keep(built)
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	return kept
}

func changed(built task.Task, mutate func(a *task.AsTold)) task.Task {
	a := &task.AsTold{}
	*a = task.From(built)
	mutate(a)
	return a.Task()
}

func factsOf(t task.Task) task.Carried {
	f := t.Carried()
	return f
}

func TestStore_handsBackTheTaskItKept(t *testing.T) {
	s := openStore(t)
	kept := keep(t, s, "Write the spec")

	got, err := s.Find(factsOf(kept).Number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	facts := factsOf(got)
	if facts.Title != "Write the spec" {
		t.Errorf("Title = %q, want the one it was given", facts.Title)
	}
	if facts.State != domain.Backlog || facts.Priority != domain.Normal {
		t.Errorf("got %q/%q, want backlog/normal", facts.State, facts.Priority)
	}
}

func TestStore_numbersTheFirstTaskOne(t *testing.T) {
	s := openStore(t)
	if got := factsOf(keep(t, s, "Write the spec")).Number; got != 1 {
		t.Errorf("Number = %d, want 1", got)
	}
}

func TestStore_givesTheNextNumberToTheNextTask(t *testing.T) {
	s := openStore(t)
	keep(t, s, "First")
	if got := factsOf(keep(t, s, "Second")).Number; got != 2 {
		t.Errorf("Number = %d, want 2", got)
	}
}

func TestStore_neverHandsOutANumberFreedByADeletion(t *testing.T) {
	s := openStore(t)
	keep(t, s, "First")
	second := keep(t, s, "Second")
	if err := s.Delete(factsOf(second).Number); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got := factsOf(keep(t, s, "Third")).Number; got != 3 {
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
	if len(all) != 2 || factsOf(all[0]).Title != "First" || factsOf(all[1]).Title != "Second" {
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
		got, err := s.Find(factsOf(kept).Number)
		if err != nil {
			t.Fatalf("Find: %v", err)
		}
		if facts := factsOf(got); facts.Title != title {
			t.Errorf("Title round-trip lost data: got %q", facts.Title)
		}
	}
}

func TestStore_keepsEveryAttributeItWasGiven(t *testing.T) {
	s := openStore(t)
	built, err := task.NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	deadline, err := domain.ParseDeadline("2026-09-01")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}
	estimate, err := domain.ParseEstimate("90m")
	if err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}
	built = changed(built, func(a *task.AsTold) {
		a.Described("the one")
		a.InColumn(domain.Todo)
		a.Carries(domain.High)
		a.Labelled("spec")
		a.Labelled("team")
		a.Due(deadline)
		a.Estimated(estimate)
	})

	kept, err := s.Keep(built)
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	got, err := s.Find(factsOf(kept).Number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	facts := factsOf(got)
	if facts.Description != "the one" || facts.State != domain.Todo || facts.Priority != domain.High {
		t.Errorf("got %+v, want the attributes it was given", facts)
	}
	if len(facts.Labels) != 2 || facts.Labels[0] != "spec" || facts.Labels[1] != "team" {
		t.Errorf("Labels = %v, want [spec team]", facts.Labels)
	}
	if domain.FormatDay(facts.Deadline) != "2026-09-01" {
		t.Errorf("Deadline = %v, want 2026-09-01", facts.Deadline)
	}
	if domain.FormatDuration(facts.Estimate) != "1h30m" {
		t.Errorf("Estimate = %v, want 1h30m", facts.Estimate)
	}
}

func TestStore_roundTripsTheTimeSpent(t *testing.T) {
	s := openStore(t)
	built, err := task.NewTask("Write the spec")
	if err != nil {
		t.Fatalf("NewTask: %v", err)
	}
	built = changed(built, func(a *task.AsTold) { a.Spent(30 * time.Minute) })

	kept, err := s.Keep(built)
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	got, err := s.Find(factsOf(kept).Number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if facts := factsOf(got); facts.Spent != 30*time.Minute {
		t.Errorf("TimeSpent = %v, want 30m", facts.Spent)
	}

	bare := keep(t, s, "Bare")
	got, err = s.Find(factsOf(bare).Number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if facts := factsOf(got); facts.Spent != 0 {
		t.Errorf("TimeSpent = %v, want none", facts.Spent)
	}
}

func TestStore_replacesTheLabelsWhenATaskIsUpdated(t *testing.T) {
	s := openStore(t)
	built := keep(t, s, "Write the spec")
	built = changed(built, func(a *task.AsTold) {
		a.Labelled("spec")
		a.Labelled("draft")
	})
	if err := s.Update(built); err != nil {
		t.Fatalf("Update: %v", err)
	}
	c := built.Carried()
	c.Labels = []string{"team"}
	built = task.Made(c)
	if err := s.Update(built); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := s.Find(factsOf(built).Number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if facts := factsOf(got); len(facts.Labels) != 1 || facts.Labels[0] != "team" {
		t.Errorf("Labels = %v, want only [team]: what is given replaces what was there", facts.Labels)
	}
}

func TestStore_saysWhenThereIsNothingToUpdate(t *testing.T) {
	s := openStore(t)
	built, _ := task.NewTask("Nope")
	built = renumbered(built, 99)
	if err := s.Update(built); err == nil {
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
	if err := s.Delete(factsOf(second).Number); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	all, err := s.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 2 || factsOf(all[0]).Title != "First" || factsOf(all[1]).Title != "Third" {
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
	if _, err := Open(filepath.Join(path, "nested", "cairn.db")); err == nil {
		t.Error("Open should report a place it cannot create")
	}
}

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
	built, _ := task.NewTask("Write the spec")
	built = renumbered(built, 1)

	if _, err := s.Keep(built); err == nil {
		t.Error("Keep on a closed store should fail")
	}
	if err := s.Update(built); err == nil {
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

	if _, err := s.Find(factsOf(kept).Number); err == nil {
		t.Error("Find should report labels it cannot read")
	}
	if _, err := s.All(); err == nil {
		t.Error("All should report labels it cannot read")
	}
}

func TestStore_reportsLabelsItCannotKeep(t *testing.T) {
	s := openStore(t)
	built, _ := task.NewTask("Write the spec")
	built = changed(built, func(a *task.AsTold) {
		a.Labelled("spec")
		a.Labelled("spec")
	})

	if _, err := s.Keep(built); err == nil {
		t.Error("Keep should report labels it cannot record")
	}
}

func TestStore_reportsLabelsItCannotReplace(t *testing.T) {
	s := openStore(t)
	built := keep(t, s, "Write the spec")
	built = changed(built, func(a *task.AsTold) {
		a.Labelled("spec")
		a.Labelled("spec")
	})

	if err := s.Update(built); err == nil {
		t.Error("Update should report labels it cannot record")
	}
}

func TestStore_reportsADeadlineItCannotRead(t *testing.T) {
	s := openStore(t)
	kept := keep(t, s, "Write the spec")
	if _, err := s.db.Exec(`UPDATE tasks SET deadline = 'soon' WHERE number = ?`, factsOf(kept).Number); err != nil {
		t.Fatalf("writing a bad deadline: %v", err)
	}

	if _, err := s.Find(factsOf(kept).Number); err == nil {
		t.Error("Find should report a deadline it cannot make sense of")
	}
	if _, err := s.All(); err == nil {
		t.Error("All should report a deadline it cannot make sense of")
	}
}

func TestStore_reportsLabelsItCannotClear(t *testing.T) {
	s := openStore(t)
	built := keep(t, s, "Write the spec")
	if _, err := s.db.Exec(`DROP TABLE labels`); err != nil {
		t.Fatalf("dropping labels: %v", err)
	}

	if err := s.Update(built); err == nil {
		t.Error("Update should report labels it cannot clear")
	}
	if err := s.Delete(factsOf(built).Number); err == nil {
		t.Error("Delete should report labels it cannot clear")
	}
}

func TestStore_handsBackTheParentItWasGiven(t *testing.T) {
	s := openStore(t)
	under := keep(t, s, "Ship it")
	built, _ := task.NewTask("Write the tests")
	built = changed(built, func(a *task.AsTold) { a.Under(factsOf(under).Number) })

	kept, err := s.Keep(built)
	if err != nil {
		t.Fatalf("Keep: %v", err)
	}
	got, err := s.Find(factsOf(kept).Number)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if facts := factsOf(got); facts.Parent != factsOf(under).Number {
		t.Errorf("Parent = %d, want %d", facts.Parent, factsOf(under).Number)
	}
}

func TestStore_namesWhatHangsUnderATask(t *testing.T) {
	s := openStore(t)
	under := keep(t, s, "Ship it")
	built, _ := task.NewTask("Write the tests")
	built = changed(built, func(a *task.AsTold) { a.Under(factsOf(under).Number) })
	if _, err := s.Keep(built); err != nil {
		t.Fatalf("Keep: %v", err)
	}
	keep(t, s, "Unrelated")

	children, err := s.Under(factsOf(under).Number)
	if err != nil {
		t.Fatalf("Under: %v", err)
	}
	if len(children) != 1 || factsOf(children[0]).Title != "Write the tests" {
		t.Errorf("Under = %v, want only the sub-task", children)
	}
}

func TestStore_saysNothingHangsUnderALoneTask(t *testing.T) {
	s := openStore(t)
	lone := keep(t, s, "Ship it")
	children, err := s.Under(factsOf(lone).Number)
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

func TestUnusable_tellsAStoreOutOfReachFromAnythingElse(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	made, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := made.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if err := os.Chmod(path, 0o444); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })

	db, err := sql.Open("sqlite", shared(path))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer func() { _ = db.Close() }()

	_, refused := db.Exec(`INSERT INTO tasks (parent, title, description,
		state, priority, deadline, estimate, time_spent)
		VALUES (0, 'x', '', 'backlog', 'normal', '', 0, 0)`)
	if refused == nil {
		t.Fatal("writing to a read-only store was expected to fail")
	}
	if !Unusable(refused) {
		t.Errorf("Unusable(%v) = false, want true", refused)
	}
	if !Unusable(fmt.Errorf("keeping the task: %w", refused)) {
		t.Error("Unusable does not see through a wrapped error")
	}

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

func TestOpen_waitsForAStoreSomebodyElseIsMaking(t *testing.T) {
	path := filepath.Join(t.TempDir(), "made-at-once", "cairn.db")

	const together = 4
	var wg sync.WaitGroup
	failed := make(chan error, together)
	for range together {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s, err := Open(path)
			if err != nil {
				failed <- err
				return
			}
			_ = s.Close()
		}()
	}
	wg.Wait()
	close(failed)

	for err := range failed {
		t.Errorf("one of the %d refused to wait: %v", together, err)
	}
}

func TestBeingHeld_readsEveryShapeOfBeingHeld(t *testing.T) {
	for code, held := range map[int]bool{
		5:   true,
		261: true,
		517: true,
		773: true,
		8:   false,
		1:   false,
		0:   false,
		6:   false,
	} {
		if got := beingHeld(code); got != held {
			t.Errorf("beingHeld(%d) = %v, want %v", code, got, held)
		}
	}
}

func TestLayDownTheShape_declinesWhenSomebodyElseAlreadyMadeIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = s.Close() }()

	if err := s.layDownTheShapeWhenOpeningFindsNothing(); err != nil {
		t.Fatalf("laying the shape down on a made store: %v", err)
	}
	if held, err := s.Held(); err != nil || held.Tasks != 0 {
		t.Errorf("held = %v (%v), want a store still holding nothing and no second shape laid on it", held, err)
	}
}

func TestHeldFor_isTheSameWaitBothWaysRound(t *testing.T) {
	if got := waitedOut(); got != 5*time.Second {
		t.Errorf("waitedOut() = %v, want 5s", got)
	}
	if got := shared("/somewhere/cairn.db"); !strings.Contains(got, "busy_timeout(5000)") {
		t.Errorf("the store is opened as %q, want it to give SQLite the same 5s", got)
	}
}

func TestOpen_reportsAShapeItCouldNotLayDown(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")

	db, err := sql.Open("sqlite", shared(path))
	if err != nil {
		t.Fatalf("opening the file: %v", err)
	}
	if _, err := db.Exec(ledger); err != nil {
		t.Fatalf("making the ledger: %v", err)
	}
	if _, err := db.Exec(`CREATE VIEW tasks AS SELECT 1 AS number`); err != nil {
		t.Fatalf("putting something in the way: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("closing: %v", err)
	}

	_, err = Open(path)
	if err == nil {
		t.Fatal("Open said nothing, want it to report a shape it could not lay down")
	}
	if !strings.Contains(err.Error(), "applying") {
		t.Errorf("Open said %q, want it to name the change it could not apply", err)
	}
}

func TestStore_keepsWhetherATaskIsArchived(t *testing.T) {
	s := openStore(t)
	kept := keep(t, s, "Migration Jenkins")

	a := task.From(kept)
	a.InColumn(domain.Done)
	a.Archived()
	if err := s.Update(a.Task()); err != nil {
		t.Fatalf("Update: %v", err)
	}

	found, err := s.Find(1)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	read := found.Carried()
	if !read.Archived {
		t.Error("the store gave back a task that is not archived")
	}
	if read.State != domain.Done {
		t.Errorf("state = %q, want it still done — archiving is not a column", read.State)
	}
}

func TestWithin_letsGoOfTheStoreWhenAnActPanics(t *testing.T) {
	s := openStore(t)
	func() {
		defer func() { _ = recover() }()
		_ = s.Within(func(*Store) error { panic("an act that went wrong") })
	}()
	if err := s.Within(func(kept *Store) error { _, err := kept.Keep(mustTask(t, "after")); return err }); err != nil {
		t.Errorf("Within after a panic = %v, want the store free again", err)
	}
}

func mustTask(t *testing.T, title string) task.Task {
	t.Helper()
	built, err := task.NewTask(title)
	if err != nil {
		t.Fatal(err)
	}
	return built
}
