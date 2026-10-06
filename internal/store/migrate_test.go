package store

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/migration"
	"cairn/internal/objective"
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

	if _, err := db.Exec(factsOfChange(migration.Known[0]).does); err != nil {
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
	if want := len(migration.Known) - 1; len(waiting) != want {
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
		`INSERT INTO migrations (name, applied_at) VALUES ('never heard of', '2099-01-01T00:00:00Z')`); err != nil {
		t.Fatalf("writing a change from the future: %v", err)
	}
	_ = s.Close()

	if _, err := Open(path); !errors.Is(err, ErrNewerStore) {
		t.Errorf("err = %v, want ErrNewerStore", err)
	}
}

// When a change was applied is read back as the moment it was written, and a
// ledger that does not say one is a ledger that cannot be read. It used to be
// read as the year one, and listed as applied then.
func TestMigrations_refusesALedgerThatDoesNotSayWhen(t *testing.T) {
	s := openStore(t)
	if _, err := s.db.Exec(`UPDATE migrations SET applied_at = 'yesterday'`); err != nil {
		t.Fatalf("spoiling the ledger: %v", err)
	}
	if _, err := s.Migrations(); err == nil || !strings.Contains(err.Error(), "yesterday") {
		t.Errorf("Migrations = %v, want the moment it could not read named", err)
	}
}

func TestMigrations_namesEveryChangeAndItsState(t *testing.T) {
	s := openStore(t)
	all, err := s.Migrations()
	if err != nil {
		t.Fatalf("Migrations: %v", err)
	}
	if len(all) != len(migration.Known) {
		t.Fatalf("%d changes, want %d", len(all), len(migration.Known))
	}
	for i, change := range all {
		f := factsOfChange(change)
		if f.number != i+1 || f.name == "" || f.does == "" {
			t.Errorf("change %d is incomplete: %+v", i, f)
		}
		if !f.applied || f.made.IsZero() {
			t.Errorf("change %d should be applied and dated: %+v", i, f)
		}
		if want := factsOfChange(migration.Known[i]).undoes; f.undoes != want {
			t.Errorf("change %d: Undoes = %q, want %q", i+1, f.undoes, want)
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
	if want := factsOfChange(migration.Known[1]).name; name != want {
		t.Errorf("applied %q, want %q", name, want)
	}
	waiting, err := s.Waiting()
	if err != nil {
		t.Fatalf("Waiting: %v", err)
	}
	if want := len(migration.Known) - 2; len(waiting) != want {
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
	if want := len(migration.Known) - 1; len(made) != want {
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
	if len(all) != 1 || factsOf(all[0]).title != "Write the spec" {
		t.Errorf("All = %v, want what the older store held", all)
	}
}

func TestRollBack_unmakesTheLastChangeApplied(t *testing.T) {
	s := openStore(t)
	one, err := s.RollBack()
	if err != nil {
		t.Fatalf("RollBack: %v", err)
	}
	last := factsOfChange(migration.Known[len(migration.Known)-1])
	got := factsOfChange(one)
	if got.name != last.name {
		t.Errorf("rolled back %q, want the last change", got.name)
	}
	if got.loses != last.loses {
		t.Errorf("Loses = %q, want %q: a rollback has to be able to say what it cost", got.loses, last.loses)
	}
}

func TestMigrations_saysWhatUndoingEachOneLoses(t *testing.T) {
	all, err := openStore(t).Migrations()
	if err != nil {
		t.Fatalf("Migrations: %v", err)
	}
	for i, one := range all {
		got, want := factsOfChange(one).loses, factsOfChange(migration.Known[i]).loses
		if got != want {
			t.Errorf("change %d: Loses = %q, want %q", i+1, got, want)
		}
	}
	// The shape a store starts with cannot be rolled back, so it costs nothing.
	if got := factsOfChange(all[0]).loses; got != "" {
		t.Errorf("the first shape should lose nothing, got %q", got)
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
	for range migration.Known[1:] {
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
	// A task serves a set, so what it serves lives in a table of its own rather
	// than in a column on the task.
	if !holdsTable(t, s, "task_objectives") {
		t.Error("no task_objectives table: nothing records what a task serves")
	}
}

// rollBackPast undoes changes until the named one has been undone, deriving how
// many that is from migration.Known rather than counting by hand. Written this
// way because it already went wrong once: a test that rolled back once was
// asserting about whichever change happened to be last, and appending one made
// it assert about the wrong change while still reading as though it did not.
func rollBackPast(t *testing.T, s *Store, name string) {
	t.Helper()
	for i := len(migration.Known) - 1; i >= 0; i-- {
		if _, err := s.RollBack(); err != nil {
			t.Fatalf("RollBack: %v", err)
		}
		if factsOfChange(migration.Known[i]).name == name {
			return
		}
	}
	t.Fatalf("no change named %q to roll back to", name)
}

func TestMigrate_makesRoomForPlacesAndTheFocus(t *testing.T) {
	s := openStore(t)
	if !holdsColumn(t, s, "tasks", "position") {
		t.Error("tasks carry no \"position\": the order has nowhere to be kept")
	}
	// "focused" and not "focus": the one row that held one task became a row
	// per task, so that several can be worked on at once.
	if !holdsTable(t, s, "focused") {
		t.Error("no \"focused\" table: what is being worked on has nowhere to be kept")
	}
	if holdsTable(t, s, "focus") {
		t.Error("the old one-row \"focus\" table is still there beside the new one")
	}
}

// Rolling back the change that made room for several keeps the first task
// taken up, which is the most a single row can hold.
func TestMigrate_rollingBackTheManyFocusKeepsTheFirst(t *testing.T) {
	s := openStore(t)
	first := factsOf(keep(t, s, "First")).number
	second := factsOf(keep(t, s, "Second")).number
	for _, n := range []int{first, second} {
		if err := s.AddFocus(n); err != nil {
			t.Fatalf("AddFocus(%d): %v", n, err)
		}
	}
	rollBackPast(t, s, "more than one task can be in focus")
	if holdsTable(t, s, "focused") {
		t.Error("\"focused\" survived the rollback: a change that cannot be undone is not one")
	}
	if !holdsTable(t, s, "focus") {
		t.Fatal("the one-row \"focus\" table did not come back")
	}
	var held int
	if err := s.db.QueryRow(`SELECT task_number FROM focus WHERE id = 1`).Scan(&held); err != nil {
		t.Fatalf("reading the one row back: %v", err)
	}
	if held != first {
		t.Errorf("the one row holds task %d, want the first taken up, %d", held, first)
	}
}

func TestMigrate_rollingBackTakesThePlacesAway(t *testing.T) {
	s := openStore(t)
	rollBackPast(t, s, "a task carries a place, and one task is in focus")
	if holdsColumn(t, s, "tasks", "position") {
		t.Error("tasks still carry \"position\" after the rollback")
	}
	for _, table := range []string{"focus", "focused"} {
		if holdsTable(t, s, table) {
			t.Errorf("%q survived the rollback: a change that cannot be undone is not one", table)
		}
	}
}

func TestMigrate_rollingBackTakesTheRoomAway(t *testing.T) {
	s := openStore(t)
	rollBackPast(t, s, "objectives, and the tasks that serve them")
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
	last := len(migration.Known)
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
		o, err := objective.NewObjective(title)
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

// Opening a store must not write to it. Every command opened the store, and
// every open ran `CREATE TABLE IF NOT EXISTS migrations` — DDL, replayed for
// nothing on a store that already had it. Reading is reading: the shape is
// looked at, and `migration apply` is what changes it.
//
// A read-only file is the one test that cannot be fooled: anything settle
// writes fails outright, whatever locks or timestamps say.
func TestOpen_writesNothingToAStoreThatAlreadyHasItsShape(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	made, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	keep(t, made, "Write the spec")
	if err := made.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := os.Chmod(path, 0o444); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })

	read, err := Open(path)
	if err != nil {
		t.Fatalf("Open on a read-only store: %v", err)
	}
	defer func() { _ = read.Close() }()
	held, err := read.All()
	if err != nil {
		t.Fatalf("All on a read-only store: %v", err)
	}
	if len(held) != 1 {
		t.Errorf("%d tasks, want the one that was there", len(held))
	}
}

// A store that has never been written cannot be made on a file nobody may
// write to. It is reported rather than half-opened: the ledger is the first
// thing a new store needs, and without it nothing else can be recorded.
func TestOpen_reportsAStoreItCannotCreateTheLedgerIn(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	if err := os.WriteFile(path, nil, 0o444); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })

	if _, err := Open(path); err == nil {
		t.Error("Open on a read-only empty file returned no error")
	}
}

// knowing swaps the changes this cairn knows for the first shape and one more,
// for as long as a test runs.
func knowing(t *testing.T, does, undoes string) {
	t.Helper()
	was := migration.Known
	t.Cleanup(func() { migration.Known = was })
	var a migration.AsTold
	a.Named("a change written to fail half way")
	a.Does(does)
	a.Undoes(undoes)
	migration.Known = []migration.Migration{was[0], a.Migration()}
}

func TestApply_leavesNothingOfAChangeThatFailedHalfWay(t *testing.T) {
	knowing(t, `CREATE TABLE half (x); CREATE TABLE half (x);`, `DROP TABLE half`)
	s, err := OpenBehind(olderStore(t, "Write the spec"))
	if err != nil {
		t.Fatalf("OpenBehind: %v", err)
	}
	defer func() { _ = s.Close() }()

	if _, err := s.Apply(2); err == nil {
		t.Fatal("Apply succeeded, want the second statement refused")
	}
	if holdsTable(t, s, "half") {
		t.Error("the first statement was kept: a change is made whole or not at all")
	}
	if waiting, _ := s.Waiting(); len(waiting) != 1 {
		t.Errorf("%d changes waiting, want the failed one still waiting", len(waiting))
	}
}

func TestApplyAll_leavesNothingOfAChangeThatFailedHalfWay(t *testing.T) {
	knowing(t, `CREATE TABLE half (x); CREATE TABLE half (x);`, `DROP TABLE half`)
	s, err := OpenBehind(olderStore(t, "Write the spec"))
	if err != nil {
		t.Fatalf("OpenBehind: %v", err)
	}
	defer func() { _ = s.Close() }()

	if _, err := s.ApplyAll(); err == nil {
		t.Fatal("ApplyAll succeeded, want the second statement refused")
	}
	if holdsTable(t, s, "half") {
		t.Error("the first statement was kept: a change is made whole or not at all")
	}
}

func TestRollBack_leavesTheChangeWholeWhenUndoingItFailsHalfWay(t *testing.T) {
	knowing(t, `CREATE TABLE half (x)`, `DROP TABLE half; DROP TABLE nowhere;`)
	s, err := OpenBehind(olderStore(t, "Write the spec"))
	if err != nil {
		t.Fatalf("OpenBehind: %v", err)
	}
	defer func() { _ = s.Close() }()
	if _, err := s.Apply(2); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if _, err := s.RollBack(); err == nil {
		t.Fatal("RollBack succeeded, want the second statement refused")
	}
	if !holdsTable(t, s, "half") {
		t.Error("the table was dropped though the rollback failed: undone whole or not at all")
	}
	if waiting, _ := s.Waiting(); len(waiting) != 0 {
		t.Errorf("%d changes waiting, want the change still applied", len(waiting))
	}
}

func TestMake_saysAChangeAnotherCairnAppliedFirstIsApplied(t *testing.T) {
	s, err := OpenBehind(olderStore(t, "Write the spec"))
	if err != nil {
		t.Fatalf("OpenBehind: %v", err)
	}
	defer func() { _ = s.Close() }()
	if _, err := s.Apply(2); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	// What a second cairn does when it read the change as waiting a moment
	// before the first applied it.
	if err := s.make(migration.Known[1]); !errors.Is(err, ErrAlreadyApplied) {
		t.Errorf("make of a change already applied = %v, want ErrAlreadyApplied", err)
	}
}

func TestApply_refusesAChangeWithAnEarlierOneWaiting(t *testing.T) {
	s, err := OpenBehind(olderStore(t, "Write the spec"))
	if err != nil {
		t.Fatalf("OpenBehind: %v", err)
	}
	defer func() { _ = s.Close() }()
	if _, err := s.Apply(3); !errors.Is(err, ErrEarlierWaiting) {
		t.Fatalf("Apply(3) with change 2 waiting = %v, want ErrEarlierWaiting", err)
	}
	if _, err := s.Apply(2); err != nil {
		t.Errorf("Apply(2) = %v, want the next change applied", err)
	}
}

func TestRollBack_saysAChangeAnotherCairnUndidFirstWasUndone(t *testing.T) {
	s := openStore(t)
	// What a second cairn read a moment before the first one rolled back.
	stale, err := s.Migrations()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.RollBack(); err != nil {
		t.Fatalf("RollBack: %v", err)
	}
	if _, err := s.rollBackFrom(stale); !errors.Is(err, ErrUndoneMeanwhile) {
		t.Errorf("a rollback read before another one = %v, want ErrUndoneMeanwhile", err)
	}
}

func TestOpen_refusesAnotherApplicationsDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "theirs.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE invoices (id INTEGER PRIMARY KEY, total REAL)`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	if _, err := Open(path); !errors.Is(err, ErrNotAStore) {
		t.Fatalf("Open = %v, want ErrNotAStore", err)
	}
	check, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = check.Close() }()
	var n int
	if err := check.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type = 'table'`).Scan(&n); err != nil || n != 1 {
		t.Errorf("%d tables, %v: want the other application's one table and nothing of cairn's", n, err)
	}
}
