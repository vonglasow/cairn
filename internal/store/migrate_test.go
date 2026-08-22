package store

import (
	"cairn/internal/domain"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

// olderStore writes a store the way cairn wrote them before it knew about
// sub-tasks: the first shape, and no record of what has been applied.
func olderStore(t *testing.T, title string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer func() { _ = db.Close() }()

	if _, err := db.Exec(migrations[0].up); err != nil {
		t.Fatalf("writing the older shape: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO tasks (title, state, priority) VALUES (?, 'backlog', 'normal')`, title); err != nil {
		t.Fatalf("writing a task: %v", err)
	}
	return path
}

func TestOpen_refusesAStoreBehindThisCairn(t *testing.T) {
	_, err := Open(olderStore(t, "Write the spec"))
	if !errors.Is(err, ErrStoreBehind) {
		t.Errorf("err = %v, want ErrStoreBehind: opening must not change the shape", err)
	}
}

func TestOpenBehind_opensAStoreWhateverShapeItIsIn(t *testing.T) {
	s, err := OpenBehind(olderStore(t, "Write the spec"))
	if err != nil {
		t.Fatalf("OpenBehind: %v", err)
	}
	defer func() { _ = s.Close() }()

	waiting, err := s.Waiting()
	if err != nil {
		t.Fatalf("Waiting: %v", err)
	}
	if want := len(migrations) - 1; len(waiting) != want {
		t.Errorf("%d changes waiting, want %d", len(waiting), want)
	}
}

func TestOpen_makesAStoreThatHeldNothingAtAll(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "cairn.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = s.Close() }()

	waiting, err := s.Waiting()
	if err != nil {
		t.Fatalf("Waiting: %v", err)
	}
	if len(waiting) != 0 {
		t.Errorf("%d changes waiting, want none: a store made today is made current", len(waiting))
	}
}

func TestOpen_refusesAStoreFromANewerCairn(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := s.db.Exec(
		`INSERT INTO migrations (name, applied_at) VALUES ('never heard of', '2099-01-01')`); err != nil {
		t.Fatalf("writing a change from the future: %v", err)
	}
	_ = s.Close()

	if _, err := Open(path); !errors.Is(err, ErrNewerStore) {
		t.Errorf("err = %v, want ErrNewerStore", err)
	}
}

func TestMigrations_namesEveryChangeAndItsState(t *testing.T) {
	s := openStore(t)
	all, err := s.Migrations()
	if err != nil {
		t.Fatalf("Migrations: %v", err)
	}
	if len(all) != len(migrations) {
		t.Fatalf("%d changes, want %d", len(all), len(migrations))
	}
	for i, change := range all {
		if change.Number != i+1 || change.Name == "" || change.Does == "" {
			t.Errorf("change %d is incomplete: %+v", i, change)
		}
		if !change.Applied || change.Made.IsZero() {
			t.Errorf("change %d should be applied and dated: %+v", i, change)
		}
	}
}

func TestApply_makesOneChange(t *testing.T) {
	s, err := OpenBehind(olderStore(t, "Write the spec"))
	if err != nil {
		t.Fatalf("OpenBehind: %v", err)
	}
	defer func() { _ = s.Close() }()

	// The number is derived, not written: this test is about applying one
	// change and leaving the rest alone, which is a claim that has to keep
	// holding when a third migration is added.
	name, err := s.Apply(2)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if name != migrations[1].name {
		t.Errorf("applied %q, want %q", name, migrations[1].name)
	}
	waiting, err := s.Waiting()
	if err != nil {
		t.Fatalf("Waiting: %v", err)
	}
	if want := len(migrations) - 2; len(waiting) != want {
		t.Errorf("%d changes still waiting, want %d", len(waiting), want)
	}
}

func TestApply_refusesAChangeThatDoesNotExist(t *testing.T) {
	s := openStore(t)
	for _, number := range []int{0, 99} {
		if _, err := s.Apply(number); !errors.Is(err, ErrNoSuchChange) {
			t.Errorf("Apply(%d) err = %v, want ErrNoSuchChange", number, err)
		}
	}
}

func TestApply_refusesAChangeAlreadyApplied(t *testing.T) {
	if _, err := openStore(t).Apply(1); !errors.Is(err, ErrAlreadyApplied) {
		t.Errorf("err = %v, want ErrAlreadyApplied", err)
	}
}

func TestApplyAll_makesEverythingWaiting(t *testing.T) {
	s, err := OpenBehind(olderStore(t, "Write the spec"))
	if err != nil {
		t.Fatalf("OpenBehind: %v", err)
	}
	defer func() { _ = s.Close() }()

	made, err := s.ApplyAll()
	if err != nil {
		t.Fatalf("ApplyAll: %v", err)
	}
	if want := len(migrations) - 1; len(made) != want {
		t.Errorf("applied %v, want the %d waiting", made, want)
	}
}

func TestApplyAll_saysWhenNothingIsWaiting(t *testing.T) {
	if _, err := openStore(t).ApplyAll(); !errors.Is(err, ErrNothingWaiting) {
		t.Errorf("err = %v, want ErrNothingWaiting", err)
	}
}

func TestApplyAll_keepsWhatAnOlderStoreHeld(t *testing.T) {
	s, err := OpenBehind(olderStore(t, "Write the spec"))
	if err != nil {
		t.Fatalf("OpenBehind: %v", err)
	}
	defer func() { _ = s.Close() }()

	if _, err := s.ApplyAll(); err != nil {
		t.Fatalf("ApplyAll: %v", err)
	}
	all, err := s.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(all) != 1 || all[0].Title != "Write the spec" {
		t.Errorf("All = %v, want what the older store held", all)
	}
}

func TestRollBack_unmakesTheLastChangeApplied(t *testing.T) {
	s := openStore(t)
	one, err := s.RollBack()
	if err != nil {
		t.Fatalf("RollBack: %v", err)
	}
	last := migrations[len(migrations)-1]
	if one.Name != last.name {
		t.Errorf("rolled back %q, want the last change", one.Name)
	}
	if one.Loses != last.loses {
		t.Errorf("Loses = %q, want %q: a rollback has to be able to say what it cost", one.Loses, last.loses)
	}
}

func TestMigrations_saysWhatUndoingEachOneLoses(t *testing.T) {
	all, err := openStore(t).Migrations()
	if err != nil {
		t.Fatalf("Migrations: %v", err)
	}
	for i, one := range all {
		if one.Loses != migrations[i].loses {
			t.Errorf("change %d: Loses = %q, want %q", one.Number, one.Loses, migrations[i].loses)
		}
	}
	// The shape a store starts with cannot be rolled back, so it costs nothing.
	if all[0].Loses != "" {
		t.Errorf("the first shape should lose nothing, got %q", all[0].Loses)
	}
}

func TestRollBack_keepsWhatTheOlderShapeCanStillHold(t *testing.T) {
	s := openStore(t)
	keep(t, s, "Write the spec")
	keep(t, s, "Ship it")

	if _, err := s.RollBack(); err != nil {
		t.Fatalf("RollBack: %v", err)
	}
	held, err := s.Held()
	if err != nil {
		t.Fatalf("Held: %v", err)
	}
	if held.Tasks != 2 {
		t.Errorf("%d tasks left, want 2: rolling back changes the shape, not the work", held)
	}
}

func TestRollBack_refusesTheShapeAStoreStartsWith(t *testing.T) {
	s := openStore(t)
	for range migrations[1:] {
		if _, err := s.RollBack(); err != nil {
			t.Fatalf("RollBack: %v", err)
		}
	}
	if _, err := s.RollBack(); !errors.Is(err, ErrNothingToRollBack) {
		t.Errorf("err = %v, want ErrNothingToRollBack", err)
	}
}

func TestMigrations_reportsAStoreItCannotRead(t *testing.T) {
	s := closed(t)
	if _, err := s.Migrations(); err == nil {
		t.Error("Migrations on a closed store should fail")
	}
	if _, err := s.Waiting(); err == nil {
		t.Error("Waiting on a closed store should fail")
	}
	if _, err := s.Apply(2); err == nil {
		t.Error("Apply on a closed store should fail")
	}
	if _, err := s.ApplyAll(); err == nil {
		t.Error("ApplyAll on a closed store should fail")
	}
	if _, err := s.RollBack(); err == nil {
		t.Error("RollBack on a closed store should fail")
	}
}

// holdsTable answers whether the store's shape has a table by that name.
func holdsTable(t *testing.T, s *Store, name string) bool {
	t.Helper()
	var found string
	err := s.db.QueryRow(
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, name,
	).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return false
	}
	if err != nil {
		t.Fatalf("looking for table %q: %v", name, err)
	}
	return true
}

// holdsColumn answers whether a table carries a column by that name.
func holdsColumn(t *testing.T, s *Store, table, column string) bool {
	t.Helper()
	rows, err := s.db.Query(`SELECT name FROM pragma_table_info(?)`, table)
	if err != nil {
		t.Fatalf("reading the shape of %q: %v", table, err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("reading a column of %q: %v", table, err)
		}
		if name == column {
			return true
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("reading the shape of %q: %v", table, err)
	}
	return false
}

func TestMigrate_makesRoomForObjectives(t *testing.T) {
	s := openStore(t)
	for _, table := range []string{"objectives", "objective_labels"} {
		if !holdsTable(t, s, table) {
			t.Errorf("no %q table: an objective has nowhere to be kept", table)
		}
	}
	if !holdsColumn(t, s, "tasks", "objective") {
		t.Error("tasks carry no objective: nothing records what a task serves")
	}
}

func TestMigrate_rollingBackTakesTheRoomAway(t *testing.T) {
	s := openStore(t)
	if _, err := s.RollBack(); err != nil {
		t.Fatalf("RollBack: %v", err)
	}
	for _, table := range []string{"objectives", "objective_labels"} {
		if holdsTable(t, s, table) {
			t.Errorf("%q survived the rollback: a change that cannot be undone is not one", table)
		}
	}
	if holdsColumn(t, s, "tasks", "objective") {
		t.Error("tasks still carry an objective after the rollback")
	}
}

// The boundary is derived, never written: Apply(2) used to be the last change
// and exercised it by accident. A third migration ended that, and the mutant
// turning `number > len(all)` into `>=` survived in the gap.
func TestChange_knowsExactlyWhereTheNumbersStop(t *testing.T) {
	s := openStore(t)
	last := len(migrations)
	if _, err := s.Change(last); err != nil {
		t.Errorf("Change(%d): %v: the last change is still a change", last, err)
	}
	if _, err := s.Change(last + 1); !errors.Is(err, ErrNoSuchChange) {
		t.Errorf("Change(%d) err = %v, want ErrNoSuchChange", last+1, err)
	}
}

// A backup that named its tasks and stayed quiet about its objectives would
// let somebody restore over a store full of planning, believing they were
// risking two tasks.
func TestHeld_countsObjectivesApartFromTasks(t *testing.T) {
	s := openStore(t)
	keep(t, s, "Write the spec")
	keep(t, s, "Ship it")
	for _, title := range []string{"Ship cairn 1.0", "Learn Go", "Redo the site"} {
		o, err := domain.NewObjective(title)
		if err != nil {
			t.Fatalf("NewObjective: %v", err)
		}
		if _, err := s.KeepObjective(o); err != nil {
			t.Fatalf("KeepObjective: %v", err)
		}
	}
	held, err := s.Held()
	if err != nil {
		t.Fatalf("Held: %v", err)
	}
	if held.Tasks != 2 || held.Objectives != 3 {
		t.Errorf("Held = %+v, want 2 tasks and 3 objectives", held)
	}
}

func TestHeld_countsNothingInAnEmptyStore(t *testing.T) {
	held, err := openStore(t).Held()
	if err != nil {
		t.Fatalf("Held: %v", err)
	}
	if held.Tasks != 0 || held.Objectives != 0 {
		t.Errorf("Held = %+v, want nothing", held)
	}
}

func TestHeld_reportsAStoreItCannotRead(t *testing.T) {
	s := openStore(t)
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := s.Held(); err == nil {
		t.Error("Held on a closed store returned no error")
	}
}
