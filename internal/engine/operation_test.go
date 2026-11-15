package engine

import (
	"cairn/internal/engine/app"
	"cairn/internal/engine/backup"
	"cairn/internal/engine/item"
	"cairn/internal/engine/label"
	"cairn/internal/engine/series"
	"cairn/internal/engine/workspace"
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"cairn/internal/engine/act"
	"cairn/internal/engine/migration"
	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
)

func garbageStore(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	if err := os.WriteFile(path, []byte("not a store"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func walFileGone(t *testing.T, path string) bool {
	t.Helper()
	_, err := os.Stat(path + "-wal")
	return errors.Is(err, os.ErrNotExist)
}

func TestDo_opensTheStoreAndClosesItAgain(t *testing.T) {
	path := storeAt(t)
	op := Operation{
		Needs: aStore,
		Run: func(on workedOn, _ Input) (Answer, error) {
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
		Needs: aStoreBehind,
		Run: func(on workedOn, _ Input) (Answer, error) {
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
		Needs: aStore,
		Run: func(workedOn, Input) (Answer, error) {
			t.Fatal("Run was called on a store that should have been refused")
			return Answer{}, nil
		},
	}

	_, err := Do(op, path, Input{})
	if !errors.Is(err, migration.ErrStoreBehind) {
		t.Fatalf("err = %v, want ErrStoreBehind", err)
	}
	if !strings.Contains(err.Error(), "migration apply") {
		t.Errorf("err = %q, want it to say to run the migration", err)
	}
}

func TestDo_needsAStore_passesThroughAnErrorThatIsNotAboutBeingBehind(t *testing.T) {
	path := garbageStore(t)
	op := Operation{
		Needs: aStore,
		Run: func(workedOn, Input) (Answer, error) {
			t.Fatal("Run was called on a store that could not be opened")
			return Answer{}, nil
		},
	}

	_, err := Do(op, path, Input{})
	if err == nil {
		t.Fatal("Do: want an error, got none")
	}
	if errors.Is(err, migration.ErrStoreBehind) {
		t.Errorf("err = %v, want an error unrelated to the store being behind", err)
	}
}

func TestDo_needsNoStore_opensNothingButHandsBackBackups(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	op := Operation{
		Needs: noStore,
		Run: func(on workedOn, _ Input) (Answer, error) {
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
		Needs: aStore,
		Run: func(on workedOn, _ Input) (Answer, error) {
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
		Needs: aStore,
		Takes: []Argument{{Name: "number", Kind: Number, Of: "a task"}},
		Run: func(workedOn, Input) (Answer, error) {
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

	if got := in.change("priority"); got != nil {
		t.Errorf(`Change("priority") = %q, want nothing: it was never named`, *got)
	}
	if got := in.change("desc"); got == nil || *got != "" {
		t.Errorf(`Change("desc") = %v, want a pointer to "": it was named as nothing`, got)
	}
	if got := in.change("state"); got == nil || *got != "todo" {
		t.Errorf(`Change("state") = %v, want "todo"`, got)
	}

	if got := in.changed("priority"); got != nil {
		t.Errorf(`Changed("priority") = %v, want nothing: it was never named`, *got)
	}
	if got := in.changed("label"); got == nil || len(*got) != 0 {
		t.Errorf(`Changed("label") = %v, want an empty set: it was named as nothing`, got)
	}
}

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
		"trace_show":         {{Acts: []act.Act{}}},
		"trace_list":         {{Since: []act.Act{}}},
		"objective_new":      {{Objective: of(objective.Objective{})}},
		"objective_list":     {{Objectives: []objective.Objective{}}},
		"objective_move":     {{Objectives: []objective.Objective{}}},
		"objective_show":     {{Objective: of(objective.Objective{})}},
		"objective_report":   {{Objective: of(objective.Objective{}), Period: of(objective.Period{})}},
		"objective_edit":     {{Objective: of(objective.Objective{})}},
		"objective_search":   {{MatchedObjectives: []objective.Objective{}}},
		"objective_delete":   {{Deleted: of(1)}},
		"backup_new":         {{Backup: of(backup.Backup{}), Holds: of(backup.Held{})}},
		"backup_list":        {{Backups: []backup.Backup{}}},
		"backup_show":        {{Backup: of(backup.Backup{}), Holds: of(backup.Held{})}},
		"backup_delete":      {{Deleted: of(1)}},
		"backup_restore":     {{Restored: of(1)}},
		"task_export":        {{Exported: func() *string { s := "/elsewhere.json"; return &s }(), Holds: of(backup.Held{})}},
		"task_import":        {{Imported: func() *string { s := "/elsewhere.json"; return &s }(), Holds: of(backup.Held{})}},
		"plugin_write":       {{Plugin: func() *string { s := "/claude-plugin"; return &s }()}},
		"plugin_show":        {{Install: &app.PluginShown{Version: "v0.12.0", Install: app.InstallFromTag("v0.12.0")}}},
		"app_install":        {{Installed: func() *string { s := "/Applications/cairn.app"; return &s }(), Runs: func() *string { s := "/usr/local/bin/cairn"; return &s }(), OpenedFrom: func() *string { s := "the Dock"; return &s }()}},
		"migration_list":     {{Changes: []migration.Migration{}, Holds: of(backup.Held{})}},
		"migration_show":     {{Change: of(migration.Migration{})}},
		"migration_apply":    {{Applied: []string{}}},
		"migration_rollback": {{RolledBack: of(migration.Migration{})}},
		"config_show":        {{Config: &workspace.Config{}}},
		"config_init":        {{Written: func() *string { s := "/config.toml"; return &s }()}},
		"workspace_list":     {{Workspaces: &workspace.Workspaces{}}},
		"workspace_add":      {{WorkspaceAdded: &workspace.Added{}}},
		"workspace_use":      {{WorkspaceUsed: &workspace.Used{}}},
		"workspace_show":     {{WorkspaceShown: &workspace.Shown{}}},
		"debug_report":       {{Written: func() *string { s := "/cairn-debug.zip"; return &s }(), Packed: of(2)}},
		"timesheet_show":     {{Sheet: &act.Sheet{}}},
		"timesheet_add":      {{Sheet: &act.Sheet{}}},
		"label_list":         {{Labels: &[]label.Carried{}}},
		"label_new":          {{Labels: &[]label.Carried{}}},
		"label_colour":       {{Labels: &[]label.Carried{}}},
		"label_scope":        {{Labels: &[]label.Carried{}}},
		"label_delete":       {{Labels: &[]label.Carried{}}},
		"label_rename":       {{Labels: &[]label.Carried{}}},
		"view_list":          {{ViewsKept: &Views{}}},
		"task_skip":          {{Task: of(task.Task{})}},
		"series_list":        {{SeriesKept: &seriesListed{}}},
		"series_show":        {{SeriesShown: &Series{}}},
		"series_edit":        {{SeriesShown: &Series{}}},
		"series_stop":        {{SeriesShown: &Series{}}},
		"series_resume":      {{SeriesShown: &Series{}}},
		"series_delete":      {{SeriesGone: &series.Deleted{}}},
		"step_new":           {{Task: of(task.Task{})}},
		"step_list":          {{Task: of(task.Task{})}},
		"step_edit":          {{Task: of(task.Task{})}},
		"step_delete":        {{Task: of(task.Task{})}},
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

func TestCheck_refusesARequiredListGivenEmpty(t *testing.T) {
	for _, op := range []Operation{TaskSearch, ObjectiveSearch} {
		if err := op.check(Input{Listed: map[string][]string{"named": {}}}); err == nil {
			t.Errorf("%s accepted an empty list of things to look for", op.Name())
		}
		if err := op.check(Input{Listed: map[string][]string{"named": {"spec"}}}); err != nil {
			t.Errorf("%s refused something to look for: %v", op.Name(), err)
		}
	}
}

func TestCheck_refusesAValueOutsideAClosedSetAsNotOneOf(t *testing.T) {
	e := openEngine(t)
	if _, err := TaskList.Run(workedOn{Engine: e}, Input{Given: map[string]string{"state": "bogus"}}); !errors.Is(err, item.ErrNotAState) {
		t.Errorf("task list --state bogus = %v, want ErrNotAState, as an edit answers", err)
	}
}

func TestCheck_refusesADayThatIsNotOne(t *testing.T) {
	op := Operation{
		Thing: "task", Verb: "list",
		Takes: []Argument{{Name: "deadline-before", Kind: day, Means: "due on or before this day"}},
	}
	if err := op.check(Input{Given: map[string]string{"deadline-before": "soon"}}); err == nil {
		t.Error("a day that is not one was allowed through")
	}
	if err := op.check(Input{Given: map[string]string{"deadline-before": "2026-10-01"}}); err != nil {
		t.Errorf("a day that is one was refused: %v", err)
	}
	if err := op.check(Input{Given: map[string]string{"deadline-before": ""}}); err != nil {
		t.Errorf("an empty day was refused: %v", err)
	}
}

func TestEveryWriteSaysWhatCallingItAgainDoes(t *testing.T) {
	for _, op := range All() {
		if op.Level != Read && op.Again == 0 {
			t.Errorf("%s writes and does not say what calling it again does", op.Name())
		}
	}
	for op, want := range map[*Operation]again{&TaskNew: doesItAgain, &TaskEdit: ChangesNothingMore, &TimesheetAdd: doesItAgain} {
		if op.Again != want {
			t.Errorf("%s: Again = %v, want %v", op.Name(), op.Again, want)
		}
	}
	if !TaskEdit.Overwrites || TaskNew.Overwrites {
		t.Error("an edit replaces what was written and an addition does not")
	}
}

func TestEveryOperationShowsHowItIsTyped(t *testing.T) {
	for _, op := range All() {
		if !strings.Contains(op.Example, "cairn "+op.Line()) {
			t.Errorf("%s carries no example of itself: %q", op.Name(), op.Example)
		}
	}
}

func TestOperation_waitingNamesThePositionalArgumentsOnly(t *testing.T) {
	op := Operation{Thing: "task", Level: Full, Takes: []Argument{
		{Name: "number", Positional: true},
		{Name: "state"},
	}}
	err := op.waiting(given("number", "3", "state", "done"))
	if want := "task 3 " + errNotConfirmed.Error(); err == nil || err.Error() != want {
		t.Errorf("err = %v, want %q", err, want)
	}
}

func TestOperation_waitingSaysASweepOnlyWhenNothingIsNamed(t *testing.T) {
	op := Operation{Thing: "task", Sweeps: func(Input) bool { return true }, Takes: []Argument{{Name: "number", Positional: true}}}
	if err := op.waiting(given()); err == nil || err.Error() != "every task it reaches "+errNotConfirmed.Error() {
		t.Errorf("nothing named: err = %v", err)
	}
	if err := op.waiting(given("number", "3")); err == nil || err.Error() != "task 3 "+errNotConfirmed.Error() {
		t.Errorf("a number named: err = %v", err)
	}
}

func TestAuthor_aNameSetEmptyIsTheOneLoggedIn(t *testing.T) {
	me, err := user.Current()
	if err != nil {
		t.Skip("no user to compare with:", err)
	}
	empty := Home{Lookup: func(k string) (string, bool) { return "", k == AuthorVariable }}
	if got := author(empty); got != me.Username {
		t.Errorf("named empty, the author is %q, want %q", got, me.Username)
	}
}
