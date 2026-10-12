package engine

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/migration"
	"cairn/internal/objective"
	"cairn/internal/task"
	"cairn/internal/trace"
)

// storeAt and behindStore are declared in store_test.go, which needs the same
// two shapes.

// garbageStore gives a path to a file that is not a store at all, so opening
// it fails with something other than the store being behind.
func garbageStore(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	if err := os.WriteFile(path, []byte("not a store"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

// walFileGone reports whether the store at path was left with nothing
// checkpointed and waiting — which only happens once every connection to it
// has closed.
func walFileGone(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Stat(path + "-wal")
	return errors.Is(err, os.ErrNotExist)
}

func TestDo_opensTheStoreAndClosesItAgain(t *testing.T) {
	path := storeAt(t)
	op := Operation{
		Needs: AStore,
		Run: func(on On, _ Input) (Answer, error) {
			if on.Engine == nil {
				t.Fatal("on.Engine is nil, want an engine open on the store")
			}
			if _, err := on.Engine.add("hello", Changes{}); err != nil {
				t.Fatalf("Add: %v", err)
			}
			return Answer{}, nil
		},
	}

	if _, err := Do(op, path, Input{}); err != nil {
		t.Fatalf("Do: %v", err)
	}

	if !walFileGone(t, path) {
		t.Errorf("the store was not closed: a write-ahead log is still waiting")
	}
}

func TestDo_needsAStoreBehind_opensTheStoreWhateverShapeItIsIn(t *testing.T) {
	path := behindStore(t)
	called := false
	op := Operation{
		Needs: AStoreBehind,
		Run: func(on On, _ Input) (Answer, error) {
			called = true
			if on.Engine == nil {
				t.Fatal("on.Engine is nil, want an engine open on the store")
			}
			if _, err := on.Engine.held(); err != nil {
				t.Fatalf("Held: %v", err)
			}
			return Answer{}, nil
		},
	}

	if _, err := Do(op, path, Input{}); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if !called {
		t.Errorf("Run was never called")
	}
}

func TestDo_needsAStore_reportsWhenTheStoreIsBehind(t *testing.T) {
	path := behindStore(t)
	op := Operation{
		Needs: AStore,
		Run: func(On, Input) (Answer, error) {
			t.Fatal("Run was called on a store that should have been refused")
			return Answer{}, nil
		},
	}

	_, err := Do(op, path, Input{})
	if !errors.Is(err, ErrStoreBehind) {
		t.Fatalf("err = %v, want ErrStoreBehind", err)
	}
	if !strings.Contains(err.Error(), "migration apply") {
		t.Errorf("err = %q, want it to say to run the migration", err)
	}
}

func TestDo_needsAStore_passesThroughAnErrorThatIsNotAboutBeingBehind(t *testing.T) {
	path := garbageStore(t)
	op := Operation{
		Needs: AStore,
		Run: func(On, Input) (Answer, error) {
			t.Fatal("Run was called on a store that could not be opened")
			return Answer{}, nil
		},
	}

	_, err := Do(op, path, Input{})
	if err == nil {
		t.Fatal("Do: want an error, got none")
	}
	if errors.Is(err, ErrStoreBehind) {
		t.Errorf("err = %v, want an error unrelated to the store being behind", err)
	}
}

func TestDo_needsNoStore_opensNothingButHandsBackBackups(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	op := Operation{
		Needs: NoStore,
		Run: func(on On, _ Input) (Answer, error) {
			if on.Engine != nil {
				t.Errorf("on.Engine = %v, want nil: nothing should have been opened", on.Engine)
			}
			if on.Backups == nil {
				t.Errorf("on.Backups = nil, want the on.Backups kept beside the store")
			}
			return Answer{}, nil
		},
	}

	if _, err := Do(op, path, Input{}); err != nil {
		t.Fatalf("Do: %v", err)
	}
}

func TestDo_closesTheStoreEvenWhenRunFails(t *testing.T) {
	path := storeAt(t)
	failure := errors.New("boom")
	op := Operation{
		Needs: AStore,
		Run: func(on On, _ Input) (Answer, error) {
			if on.Engine == nil {
				t.Fatal("on.Engine is nil, want an engine open on the store")
			}
			if _, err := on.Engine.add("hello", Changes{}); err != nil {
				t.Fatalf("Add: %v", err)
			}
			return Answer{}, failure
		},
	}

	_, err := Do(op, path, Input{})
	if !errors.Is(err, failure) {
		t.Fatalf("err = %v, want %v", err, failure)
	}
	if !walFileGone(t, path) {
		t.Errorf("the store was not closed after Run failed: a write-ahead log is still waiting")
	}
}

func TestDo_refusesANumberThatDoesNotParseBeforeOpeningTheStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	op := Operation{
		Needs: AStore,
		Takes: []Argument{{Name: "number", Kind: Number, Of: "a task"}},
		Run: func(On, Input) (Answer, error) {
			t.Fatal("Run was called on a number that does not parse")
			return Answer{}, nil
		},
	}

	_, err := Do(op, path, Input{Given: map[string]string{"number": "spec"}})
	if err == nil {
		t.Fatal("Do: want an error, got none")
	}
	if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("the store exists at %s, want a refusal to leave it untouched", path)
	}
}

func TestInput_tellsNotNamedFromNamedAsNothing(t *testing.T) {
	in := Input{
		Given:  map[string]string{"desc": "", "state": "todo"},
		Listed: map[string][]string{"label": {}},
	}

	if in.Named("priority") {
		t.Error("priority was never named, yet Named says it was")
	}
	for _, name := range []string{"desc", "state", "label"} {
		if !in.Named(name) {
			t.Errorf("%s was named, yet Named says it was not", name)
		}
	}

	if got := in.Change("priority"); got != nil {
		t.Errorf(`Change("priority") = %q, want nothing: it was never named`, *got)
	}
	if got := in.Change("desc"); got == nil || *got != "" {
		t.Errorf(`Change("desc") = %v, want a pointer to "": it was named as nothing`, got)
	}
	if got := in.Change("state"); got == nil || *got != "todo" {
		t.Errorf(`Change("state") = %v, want "todo"`, got)
	}

	if got := in.Changed("priority"); got != nil {
		t.Errorf(`Changed("priority") = %v, want nothing: it was never named`, *got)
	}
	if got := in.Changed("label"); got == nil || len(*got) != 0 {
		t.Errorf(`Changed("label") = %v, want an empty set: it was named as nothing`, got)
	}
}

// Every operation says what it answers, because a surface that describes itself
// to a machine has to say it before anything runs — and nothing else knows it.
func TestAll_everyOperationSaysWhatItAnswers(t *testing.T) {
	silent := []string{}
	for _, op := range All() {
		if op.Answers == nil {
			silent = append(silent, op.Name())
		}
	}
	if len(silent) > 0 {
		t.Errorf("%d operations say nothing about their answer: %v", len(silent), silent)
	}
}

// What an operation declares it answers is what Of actually builds from what
// its Run answers. Read off All() rather than listed: a hand-kept list held 19
// of 33 operations, and the ones left out — config show among them — were the
// ones whose answer nothing compared to anything.
func TestAnswers_agreeWithWhatOfBuilds(t *testing.T) {
	answered := map[string][]Answer{
		"task_new":           {{Task: of(task.Task{})}},
		"task_list":          {{Tasks: []task.Task{}, InFocus: list()}},
		"task_show":          {{Task: of(task.Task{})}},
		"task_archive":       {{Archived: of(1), Task: of(task.Task{})}, {Archived: of(2)}},
		"task_label":         {{Labelled: of(2)}},
		"task_unarchive":     {{Task: of(task.Task{})}},
		"task_edit":          {{Task: of(task.Task{})}},
		"task_delete":        {{Deleted: of(1)}},
		"task_search":        {{Matched: []task.Task{}}},
		"task_move":          {{Tasks: []task.Task{}}},
		"task_focus":         {{Task: of(task.Task{}), InFocus: list(1)}, {Held: []task.Task{}, InFocus: list()}},
		"task_unfocus":       {{InFocus: list()}},
		"trace_show":         {{Acts: []trace.Act{}}},
		"trace_list":         {{Since: []trace.Act{}}},
		"objective_new":      {{Objective: of(objective.Objective{})}},
		"objective_list":     {{Objectives: []objective.Objective{}}},
		"objective_show":     {{Objective: of(objective.Objective{})}},
		"objective_edit":     {{Objective: of(objective.Objective{})}},
		"objective_search":   {{MatchedObjectives: []objective.Objective{}}},
		"objective_delete":   {{Deleted: of(1)}},
		"backup_new":         {{Backup: of(domain.Backup{}), Holds: of(domain.Held{})}},
		"backup_list":        {{Backups: []domain.Backup{}}},
		"backup_show":        {{Backup: of(domain.Backup{}), Holds: of(domain.Held{})}},
		"backup_delete":      {{Deleted: of(1)}},
		"backup_restore":     {{Restored: of(1)}},
		"task_export":        {{Exported: func() *string { s := "/elsewhere.json"; return &s }(), Holds: of(domain.Held{})}},
		"task_import":        {{Imported: func() *string { s := "/elsewhere.json"; return &s }(), Holds: of(domain.Held{})}},
		"plugin_write":       {{Plugin: func() *string { s := "/claude-plugin"; return &s }()}},
		"migration_list":     {{Changes: []migration.Migration{}, Holds: of(domain.Held{})}},
		"migration_show":     {{Change: of(migration.Migration{})}},
		"migration_apply":    {{Applied: []string{}}},
		"migration_rollback": {{RolledBack: of(migration.Migration{})}},
		"config_show":        {{Config: &Config{}}},
		"timesheet_show":     {{Sheet: &trace.Sheet{}}},
		"timesheet_add":      {{Sheet: &trace.Sheet{}}},
	}
	for _, op := range All() {
		samples, known := answered[op.Name()]
		if !known {
			t.Errorf("%s: no answer to compare its declaration with — add one here", op.Name())
			continue
		}
		for _, sample := range samples {
			built := reflect.TypeOf(Of(op, sample))
			if declared := reflect.TypeOf(op.Answers); declared != built {
				t.Errorf("%s builds a %v, and says it answers a %s", op.Name(), built, declared)
			}
		}
	}
}

// Every listing answers an object naming what it carries, never a bare list. A
// client refuses a description that is not of an object, and one refused tool
// costs the whole list — so this is asserted over every operation there is,
// rather than over the ones that happen to have been thought of.
func TestAnswers_everyOperationAnswersAnObject(t *testing.T) {
	bare := []string{}
	for _, op := range All() {
		if reflect.TypeOf(op.Answers).Kind() != reflect.Struct {
			bare = append(bare, op.Name()+" answers a "+reflect.TypeOf(op.Answers).String())
		}
	}
	if len(bare) > 0 {
		t.Errorf("%d operations do not answer an object: %v", len(bare), bare)
	}
}

// A required list has to hold something: being handed an empty one is not
// being given anything to go on. The command line refused it already, through
// its own count of arguments; a server handed `{"named": []}` answered the whole
// board as if everything had matched.
func TestCheck_refusesARequiredListGivenEmpty(t *testing.T) {
	for _, op := range []Operation{TaskSearch, ObjectiveSearch} {
		if err := op.Check(Input{Listed: map[string][]string{"named": {}}}); err == nil {
			t.Errorf("%s accepted an empty list of things to look for", op.Name())
		}
		if err := op.Check(Input{Listed: map[string][]string{"named": {"spec"}}}); err != nil {
			t.Errorf("%s refused something to look for: %v", op.Name(), err)
		}
	}
}

// A value outside a closed set is a refusal, whichever operation it was given
// to — the same sentinel for a listing as for an edit, so it is answered the
// same way.
func TestCheck_refusesAValueOutsideAClosedSetAsNotOneOf(t *testing.T) {
	e := openEngine(t)
	if _, err := TaskList.Run(On{Engine: e}, Input{Given: map[string]string{"state": "bogus"}}); !errors.Is(err, domain.ErrNotAState) {
		t.Errorf("task list --state bogus = %v, want ErrNotAState, as an edit answers", err)
	}
}

// A day that is not one is refused before any Run sees it, the way a number
// that is not one already is. Without it a narrowing compared a column of days
// against whatever text was typed, and answered whatever that happened to sort
// beside — silence where a refusal belonged.
func TestCheck_refusesADayThatIsNotOne(t *testing.T) {
	op := Operation{
		Thing: "task", Verb: "list",
		Takes: []Argument{{Name: "due-before", Kind: Day, Means: "due on or before this day"}},
	}
	if err := op.Check(Input{Given: map[string]string{"due-before": "soon"}}); err == nil {
		t.Error("a day that is not one was allowed through")
	}
	if err := op.Check(Input{Given: map[string]string{"due-before": "2026-10-01"}}); err != nil {
		t.Errorf("a day that is one was refused: %v", err)
	}
	// Empty is how such an argument is cleared, and clearing is not a mistake.
	if err := op.Check(Input{Given: map[string]string{"due-before": ""}}); err != nil {
		t.Errorf("an empty day was refused: %v", err)
	}
}

// A client deciding whether to ask before a call needs to know what a second
// call does, and whether the first one can replace what was written. Every
// operation that writes says so; one added without saying is caught here.
func TestEveryWriteSaysWhatCallingItAgainDoes(t *testing.T) {
	for _, op := range All() {
		if op.Level != Read && op.Again == 0 {
			t.Errorf("%s writes and does not say what calling it again does", op.Name())
		}
	}
	for op, want := range map[*Operation]Again{&TaskNew: DoesItAgain, &TaskEdit: ChangesNothingMore, &TimesheetAdd: DoesItAgain} {
		if op.Again != want {
			t.Errorf("%s: Again = %v, want %v", op.Name(), op.Again, want)
		}
	}
	if !TaskEdit.Overwrites || TaskNew.Overwrites {
		t.Error("an edit replaces what was written and an addition does not")
	}
}

// A person reads how a command is typed faster than what it takes, so every
// operation carries at least one example of itself, typed as it is typed.
func TestEveryOperationShowsHowItIsTyped(t *testing.T) {
	for _, op := range All() {
		if !strings.Contains(op.Example, "cairn "+op.Line()) {
			t.Errorf("%s carries no example of itself: %q", op.Name(), op.Example)
		}
	}
}
