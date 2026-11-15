package engine

import (
	"cairn/internal/engine/backup"
	"cairn/internal/engine/calendar"
	"cairn/internal/engine/item"
	"encoding/json"
	"strings"
	"testing"

	"cairn/internal/engine/migration"
	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
)

func TestOfTask_carriesEveryFieldAScriptReads(t *testing.T) {
	deadline, err := calendar.ParseDeadline("2026-09-01")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}
	estimate, err := calendar.ParseEstimate("90m")
	if err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}
	spent, err := calendar.ParseEstimate("30m")
	if err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}
	var a task.AsTold
	a.Numbered(1)
	a.Step("Plan it", true)
	a.Serves(3)
	a.Titled("Write the spec")
	a.Described("in full")
	a.InColumn(item.Todo)
	a.Carries(item.High)
	a.Labelled("spec")
	a.Labelled("team")
	a.Due(deadline)
	a.Estimated(estimate)
	a.Spent(spent)
	built := a.Task()
	progress := item.Progress{Steps: item.Tally{Total: 4, Done: 1}}

	encoded, err := json.Marshal(ofTask(built, progress))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(encoded, &doc); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	want := map[string]any{
		"number":      float64(1),
		"title":       "Write the spec",
		"description": "in full",
		"state":       string(item.Todo),
		"priority":    string(item.High),
		"deadline":    "2026-09-01",
		"estimate":    "1h30m",
		"time_spent":  "30m",
	}
	for field, expected := range want {
		if doc[field] != expected {
			t.Errorf("%s = %v, want %v", field, doc[field], expected)
		}
	}
	labels, ok := doc["labels"].([]any)
	if !ok || len(labels) != 2 || labels[0] != "spec" || labels[1] != "team" {
		t.Errorf("labels = %v, want [spec team]", doc["labels"])
	}
	prog, ok := doc["progress"].(map[string]any)
	if !ok {
		t.Fatalf("progress = %v, want an object", doc["progress"])
	}
	sub, ok := prog["steps"].(map[string]any)
	if !ok || sub["total"] != float64(4) || sub["done"] != float64(1) {
		t.Errorf("progress.steps = %v, want {total:4 done:1}", prog["steps"])
	}
	if _, present := prog["tasks"]; present {
		t.Errorf("progress carries a count of tasks it does not have: %v", prog)
	}

	t.Run("a count of tasks with no step under it does not carry steps too", func(t *testing.T) {
		encoded, err := json.Marshal(ofTask(built, item.Progress{Tasks: item.Tally{Total: 2, Done: 2}}))
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		var doc map[string]any
		if err := json.Unmarshal(encoded, &doc); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		prog, ok := doc["progress"].(map[string]any)
		if !ok {
			t.Fatalf("progress = %v, want an object", doc["progress"])
		}
		if _, present := prog["steps"]; present {
			t.Errorf("progress carries a count of steps it does not have: %v", prog)
		}
	})

	t.Run("nothing under it reads as null, not as zeroes", func(t *testing.T) {
		encoded, err := json.Marshal(ofTask(built, item.Progress{}))
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		var doc map[string]any
		if err := json.Unmarshal(encoded, &doc); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if _, present := doc["progress"]; !present {
			t.Fatalf("no progress key at all in %s", encoded)
		}
		if doc["progress"] != nil {
			t.Errorf("progress = %v, want null", doc["progress"])
		}
	})

	t.Run("Of maps every field Answer carries to the shape a script reads", func(t *testing.T) {
		one, putBack := 7, 9
		var oa objective.AsTold
		oa.Numbered(5)
		oa.Titled("Ship it")
		oa.Standing(objective.Active)
		oa.Carries(item.Normal)
		obj := oa.Objective()
		made := backup.Backup{Number: 2}
		held := backup.Held{Tasks: 3, Objectives: 1}
		change := migration.Made(migration.Carried{Number: 1, Name: "add labels", Does: "ALTER TABLE tasks ADD COLUMN labels", Applied: true})

		cases := []struct {
			name string
			op   Operation
			in   Answer
			want any
		}{
			{"a task", TaskShow, Answer{Task: &built, Progress: progress}, ofTask(built, progress)},
			{"a list of tasks", TaskList, Answer{Tasks: []task.Task{built}}, Tasks{Tasks: []Row{rowOf(built, nil)}}},
			{"what a search found, the same shape as a list of tasks", TaskSearch, Answer{Matched: []task.Task{built}}, Tasks{Tasks: []Row{rowOf(built, nil)}}},
			{"an objective", ObjectiveShow, Answer{Objective: &obj, Progress: progress}, objectiveRow{
				Shape: objective.Shape{
					Number: 5, Title: "Ship it", State: "active", Priority: "normal",
					Description: "", Labels: []string{}, Deadline: "", Estimate: "",
				},
				Tasks:      []int{},
				Progress:   &item.ProgressShape{Steps: &item.Tally{Total: 4, Done: 1}, Percent: 25},
				Left:       []int{},
				Children:   []int{},
				WaitsOn:    []int{},
				WaitedOnBy: []int{},
			}},
			{"a list of objectives", ObjectiveList, Answer{Objectives: []objective.Objective{obj}}, objectivesListed{Objectives: []objectiveRow{ofObjective(obj, nil, item.Progress{})}}},
			{"what a search found among objectives", ObjectiveSearch, Answer{MatchedObjectives: []objective.Objective{obj}}, objectivesListed{Objectives: []objectiveRow{ofObjective(obj, nil, item.Progress{})}}},
			{"a backup with what it holds", backupShow, Answer{Backup: &made, Holds: &held}, ofBackup(made, &held)},
			{"a list of backups", BackupList, Answer{Backups: []backup.Backup{made}}, kept{Backups: []BackupShape{ofBackup(made, nil)}}},
			{"a change", migrationShow, Answer{Change: &change}, migration.OfMigration(change)},
			{"a list of changes", MigrationList, Answer{Changes: []migration.Migration{change}}, ledger{Changes: []migration.Shape{migration.OfMigration(change)}}},
			{"a rollback names what it undid, not a change in full", MigrationRollback, Answer{RolledBack: &change}, migration.OfRolledBack(change)},
			{"a number deleted", TaskDelete, Answer{Deleted: &one}, Deleted{Deleted: one}},
			{"a number restored", BackupRestore, Answer{Restored: &putBack}, restored{Restored: putBack}},
			{"nothing waiting reads as an empty list, not null", MigrationApply, Answer{Applied: []string{}}, applied{Applied: []string{}}},
			{"names of what was applied", MigrationApply, Answer{Applied: []string{"add labels"}}, applied{Applied: []string{"add labels"}}},
			{"a sweep says how many it took off", TaskArchive, Answer{Archived: &one}, swept{Archived: one}},
			{"one archived by number is a sweep of one, and says which", TaskArchive, Answer{Archived: of(1), Task: &built}, swept{Archived: 1, Task: of(ofTask(built, item.Progress{}))}},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				got, err := json.Marshal(Of(c.op, c.in))
				if err != nil {
					t.Fatalf("Marshal: %v", err)
				}
				want, err := json.Marshal(c.want)
				if err != nil {
					t.Fatalf("Marshal: %v", err)
				}
				if string(got) != string(want) {
					t.Errorf("Of(%+v) = %s, want %s", c.in, got, want)
				}
			})
		}
	})
}

func TestRenderText_rendersEveryFieldAnswerCarries(t *testing.T) {
	var ta task.AsTold
	ta.Numbered(1)
	ta.Titled("Write the spec")
	ta.InColumn(item.Todo)
	ta.Carries(item.Normal)
	built := ta.Task()
	progress := item.Progress{Steps: item.Tally{Total: 4, Done: 1}}

	var oa objective.AsTold
	oa.Numbered(5)
	oa.Titled("Ship it")
	oa.Standing(objective.Active)
	oa.Carries(item.Normal)
	obj := oa.Objective()

	made := backup.Backup{Number: 2}
	held := backup.Held{Tasks: 3, Objectives: 1}

	change := migration.Made(migration.Carried{Number: 1, Name: "add labels", Does: "ALTER TABLE tasks ADD COLUMN labels", Applied: true})

	deleted := 7
	putBack := 9

	cases := []struct {
		name string
		op   Operation
		in   Answer
		want string
	}{
		{"a task", TaskShow, Answer{Task: &built, Progress: progress}, task.Text(built, progress, true)},
		{"an empty list of tasks says so", TaskList, Answer{Tasks: []task.Task{}}, "there is no task\n"},
		{"a list of tasks", TaskList, Answer{Tasks: []task.Task{built}}, task.Lines([]task.Task{built}, task.Marks{})},
		{"a search that found nothing says so", TaskSearch, Answer{Matched: []task.Task{}}, "there is no task\n"},
		{
			"what a search found is listed one a line",
			TaskSearch, Answer{Matched: []task.Task{built}},
			task.Line(built) + "\n",
		},
		{"an objective", ObjectiveShow, Answer{Objective: &obj, Progress: progress}, objective.Text(obj, progress, objective.Review{}, objective.Family{}) + objective.Serving(nil)},
		{"an empty list of objectives says so", ObjectiveList, Answer{Objectives: []objective.Objective{}}, "there is no objective\n"},
		{"a list of objectives", ObjectiveList, Answer{Objectives: []objective.Objective{obj}}, objective.List([]objective.Objective{obj}, nil, nil)},
		{"a backup with what it holds", backupShow, Answer{Backup: &made, Holds: &held}, backupLine(made, held)},
		{"an empty list of backups says so", BackupList, Answer{Backups: []backup.Backup{}}, "there is no backup\n"},
		{"a list of backups", BackupList, Answer{Backups: []backup.Backup{made}}, backupsText([]backup.Backup{made})},
		{"a rollback names what it undid, not a change in full", MigrationRollback, Answer{RolledBack: &change}, migration.RolledBackText(change)},
		{"a change in full", migrationShow, Answer{Change: &change}, migration.Text(change)},
		{"a list of changes", MigrationList, Answer{Changes: []migration.Migration{change}, Holds: &held}, migration.List([]migration.Migration{change}, held)},
		{"a number deleted, named by the operation", TaskDelete, Answer{Deleted: &deleted}, "task 7 was deleted\n"},
		{"a number restored", BackupRestore, Answer{Restored: &putBack}, "backup 9 was restored\n"},
		{"nothing waiting says so", MigrationApply, Answer{Applied: []string{}}, "there is nothing waiting to be applied\n"},
		{"names of what was applied", MigrationApply, Answer{Applied: []string{"add labels"}}, "applied \"add labels\"\n"},
		{"a sweep says how many", TaskArchive, Answer{Archived: &deleted}, "7 task(s) archived\n"},
		{"one archived by number shows the card", TaskArchive, Answer{Archived: of(1), Task: &built}, task.Text(built, item.Progress{}, true)},
		{"how many were labelled", TaskLabel, Answer{Labelled: &deleted}, "7 task(s) labelled\n"},
		{"nothing declared, nothing said", Operation{}, Answer{Tasks: []task.Task{built}}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := RenderText(c.op, c.in); got != c.want {
				t.Errorf("RenderText(%+v) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func list(numbers ...int) *[]int { return &numbers }

func TestRenderText_saysNothingIsInFocus(t *testing.T) {
	got := RenderText(TaskUnfocus, Answer{InFocus: list()})
	if got != "nothing is in focus\n" {
		t.Errorf("RenderText = %q, want it to say nothing is in focus", got)
	}
}

func builtTask(t *testing.T, number int, title string) task.Task {
	t.Helper()
	built, err := task.NewTask(title)
	if err != nil {
		t.Fatalf("NewTask(%q): %v", title, err)
	}
	a := &task.AsTold{}
	*a = task.From(built)
	a.Numbered(number)
	return a.Task()
}

func TestRenderText_saysWhatASessionPastItsLengthAsks(t *testing.T) {
	first := builtTask(t, 1, "First")
	ask := "this focus session has been up 2h: renew it by focusing task 1 again, or end it by unfocusing task 1"
	got := RenderText(TaskFocus, Answer{Held: []task.Task{first}, InFocus: list(1), Sessions: []focusSession{{Task: 1, Ask: ask}}})
	if !strings.Contains(got, ask) {
		t.Errorf("the answer does not ask to renew the session:\n%s", got)
	}
}

func TestRenderText_showsWhatATerminalWouldObey(t *testing.T) {
	var a task.AsTold
	a.Numbered(1)
	a.Titled("look \x1b[8m here")
	a.Described("bell\x07 and \x1b]52;c;aGk=\x07")
	kept := a.Task()
	if got := RenderText(TaskShow, Answer{Task: &kept}); strings.ContainsAny(got, "\x1b\x07") {
		t.Errorf("RenderText = %q, want nothing a terminal would obey", got)
	}
}
