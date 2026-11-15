package engine

import (
	"cairn/internal/engine/app"
	"cairn/internal/engine/backup"
	"cairn/internal/engine/label"
	"cairn/internal/engine/series"
	"cairn/internal/engine/workspace"
	"errors"
	"os"
	"path/filepath"
	"reflect"
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
