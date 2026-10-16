package engine

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/migration"
	"cairn/internal/objective"
	"cairn/internal/task"
)

func TestOfTask_carriesEveryFieldAScriptReads(t *testing.T) {
	deadline, err := domain.ParseDeadline("2026-09-01")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}
	estimate, err := domain.ParseEstimate("90m")
	if err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}
	spent, err := domain.ParseEstimate("30m")
	if err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}
	var a task.AsTold
	a.Numbered(1)
	a.Step("Plan it", true)
	a.Serves(3)
	a.Titled("Write the spec")
	a.Described("in full")
	a.InColumn(domain.Todo)
	a.Carries(domain.High)
	a.Labelled("spec")
	a.Labelled("team")
	a.Due(deadline)
	a.Estimated(estimate)
	a.Spent(spent)
	built := a.Task()
	progress := domain.Progress{Steps: domain.Tally{Total: 4, Done: 1}}

	encoded, err := json.Marshal(OfTask(built, progress))
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
		"state":       string(domain.Todo),
		"priority":    string(domain.High),
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
		encoded, err := json.Marshal(OfTask(built, domain.Progress{Tasks: domain.Tally{Total: 2, Done: 2}}))
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
		encoded, err := json.Marshal(OfTask(built, domain.Progress{}))
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
		oa.Carries(domain.Normal)
		obj := oa.Objective()
		backup := domain.Backup{Number: 2}
		held := domain.Held{Tasks: 3, Objectives: 1}
		change := migration.Made(migration.Carried{Number: 1, Name: "add labels", Does: "ALTER TABLE tasks ADD COLUMN labels", Applied: true})

		cases := []struct {
			name string
			op   Operation
			in   Answer
			want any
		}{
			{"a task", TaskShow, Answer{Task: &built, Progress: progress}, OfTask(built, progress)},
			{"a list of tasks", TaskList, Answer{Tasks: []task.Task{built}}, Tasks{Tasks: []Row{rowOf(built)}}},
			{"what a search found, the same shape as a list of tasks", TaskSearch, Answer{Matched: []task.Task{built}}, Tasks{Tasks: []Row{rowOf(built)}}},
			{"an objective", ObjectiveShow, Answer{Objective: &obj, Progress: progress}, Objective{
				Shape: objective.Shape{
					Number: 5, Title: "Ship it", State: "active", Priority: "normal",
					Description: "", Labels: []string{}, Deadline: "", Estimate: "",
				},
				Tasks:    []int{},
				Progress: &domain.ProgressShape{Steps: &domain.Tally{Total: 4, Done: 1}, Percent: 25},
			}},
			{"a list of objectives", ObjectiveList, Answer{Objectives: []objective.Objective{obj}}, Objectives{Objectives: []Objective{OfObjective(obj, nil, domain.Progress{})}}},
			{"what a search found among objectives", ObjectiveSearch, Answer{MatchedObjectives: []objective.Objective{obj}}, Objectives{Objectives: []Objective{OfObjective(obj, nil, domain.Progress{})}}},
			{"a backup with what it holds", BackupShow, Answer{Backup: &backup, Holds: &held}, OfBackup(backup, &held)},
			{"a list of backups", BackupList, Answer{Backups: []domain.Backup{backup}}, Kept{Backups: []BackupShape{OfBackup(backup, nil)}}},
			{"a change", MigrationShow, Answer{Change: &change}, migration.OfMigration(change)},
			{"a list of changes", MigrationList, Answer{Changes: []migration.Migration{change}}, Ledger{Changes: []migration.Shape{migration.OfMigration(change)}}},
			{"a rollback names what it undid, not a change in full", MigrationRollback, Answer{RolledBack: &change}, migration.OfRolledBack(change)},
			{"a number deleted", TaskDelete, Answer{Deleted: &one}, Deleted{Deleted: one}},
			{"a number restored", BackupRestore, Answer{Restored: &putBack}, Restored{Restored: putBack}},
			{"nothing waiting reads as an empty list, not null", MigrationApply, Answer{Applied: []string{}}, Applied{Applied: []string{}}},
			{"names of what was applied", MigrationApply, Answer{Applied: []string{"add labels"}}, Applied{Applied: []string{"add labels"}}},
			{"a sweep says how many it took off", TaskArchive, Answer{Archived: &one}, Swept{Archived: one}},
			{"one archived by number is a sweep of one, and says which", TaskArchive, Answer{Archived: of(1), Task: &built}, Swept{Archived: 1, Task: of(OfTask(built, domain.Progress{}))}},
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
	ta.InColumn(domain.Todo)
	ta.Carries(domain.Normal)
	built := ta.Task()
	progress := domain.Progress{Steps: domain.Tally{Total: 4, Done: 1}}

	var oa objective.AsTold
	oa.Numbered(5)
	oa.Titled("Ship it")
	oa.Standing(objective.Active)
	oa.Carries(domain.Normal)
	obj := oa.Objective()

	backup := domain.Backup{Number: 2}
	held := domain.Held{Tasks: 3, Objectives: 1}

	change := migration.Made(migration.Carried{Number: 1, Name: "add labels", Does: "ALTER TABLE tasks ADD COLUMN labels", Applied: true})

	deleted := 7
	putBack := 9

	cases := []struct {
		name string
		op   Operation
		in   Answer
		want string
	}{
		{"a task", TaskShow, Answer{Task: &built, Progress: progress}, task.Text(built, progress)},
		{"an empty list of tasks says so", TaskList, Answer{Tasks: []task.Task{}}, "there is no task\n"},
		{"a list of tasks", TaskList, Answer{Tasks: []task.Task{built}}, task.Lines([]task.Task{built}, task.Marks{})},
		{"a search that found nothing says so", TaskSearch, Answer{Matched: []task.Task{}}, "there is no task\n"},
		{
			"what a search found is listed one a line",
			TaskSearch, Answer{Matched: []task.Task{built}},
			task.Line(built) + "\n",
		},
		{"an objective", ObjectiveShow, Answer{Objective: &obj, Progress: progress}, objective.Text(obj, progress) + objective.Serving(nil)},
		{"an empty list of objectives says so", ObjectiveList, Answer{Objectives: []objective.Objective{}}, "there is no objective\n"},
		{"a list of objectives", ObjectiveList, Answer{Objectives: []objective.Objective{obj}}, objective.List([]objective.Objective{obj})},
		{"a backup with what it holds", BackupShow, Answer{Backup: &backup, Holds: &held}, backupLine(backup, held)},
		{"an empty list of backups says so", BackupList, Answer{Backups: []domain.Backup{}}, "there is no backup\n"},
		{"a list of backups", BackupList, Answer{Backups: []domain.Backup{backup}}, backupsText([]domain.Backup{backup})},
		{"a rollback names what it undid, not a change in full", MigrationRollback, Answer{RolledBack: &change}, migration.RolledBackText(change)},
		{"a change in full", MigrationShow, Answer{Change: &change}, migration.Text(change)},
		{"a list of changes", MigrationList, Answer{Changes: []migration.Migration{change}, Holds: &held}, migration.List([]migration.Migration{change}, held)},
		{"a number deleted, named by the operation", TaskDelete, Answer{Deleted: &deleted}, "task 7 was deleted\n"},
		{"a number restored", BackupRestore, Answer{Restored: &putBack}, "backup 9 was restored\n"},
		{"nothing waiting says so", MigrationApply, Answer{Applied: []string{}}, "there is nothing waiting to be applied\n"},
		{"names of what was applied", MigrationApply, Answer{Applied: []string{"add labels"}}, "applied \"add labels\"\n"},
		{"a sweep says how many", TaskArchive, Answer{Archived: &deleted}, "7 task(s) archived\n"},
		{"one archived by number shows the card", TaskArchive, Answer{Archived: of(1), Task: &built}, task.Text(built, domain.Progress{})},
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

func TestOf_saysWhichTasksAreInFocusAndHandsThemOver(t *testing.T) {
	one := builtTask(t, 7, "The one")
	held, ok := Of(TaskFocus, Answer{Task: &one, InFocus: list(7)}).(Focus)
	if !ok {
		t.Fatalf("Of on the focus = %T, want Focus", Of(TaskFocus, Answer{Task: &one, InFocus: list(7)}))
	}
	if !slices.Equal(held.InFocus, []int{7}) {
		t.Errorf("in_focus = %v, want [7]", held.InFocus)
	}
	if len(held.Tasks) != 1 || held.Tasks[0].Number != 7 {
		t.Errorf("the tasks handed over = %v, want task 7", held.Tasks)
	}
}

func TestOf_handsOverEveryTaskInFocus(t *testing.T) {
	first := builtTask(t, 1, "First")
	second := builtTask(t, 2, "Second")
	held, ok := Of(TaskFocus, Answer{Held: []task.Task{second, first}, InFocus: list(2, 1)}).(Focus)
	if !ok {
		t.Fatal("Of on several in focus did not answer the focus")
	}
	if !slices.Equal(held.InFocus, []int{2, 1}) {
		t.Errorf("in_focus = %v, want [2 1]", held.InFocus)
	}
	if len(held.Tasks) != 2 || held.Tasks[0].Number != 2 || held.Tasks[1].Number != 1 {
		t.Errorf("the tasks handed over = %v, want 2 then 1", held.Tasks)
	}
}

func TestOf_saysNothingIsInFocusAsTheSameShape(t *testing.T) {
	none, ok := Of(TaskUnfocus, Answer{InFocus: list()}).(Focus)
	if !ok {
		t.Fatalf("Of with nothing held = %T, want Focus", Of(TaskUnfocus, Answer{InFocus: list()}))
	}
	if none.InFocus == nil || len(none.InFocus) != 0 {
		t.Errorf("in_focus = %v, want an empty list", none.InFocus)
	}
	if len(none.Tasks) != 0 {
		t.Errorf("a task crossed with nothing in focus: %v", none.Tasks)
	}
}

func TestOf_saysTheFocusOnceInAListing(t *testing.T) {
	first := builtTask(t, 1, "First")
	second := builtTask(t, 2, "Second")

	crossed, ok := Of(TaskList, Answer{Tasks: []task.Task{first, second}, InFocus: list(2)}).(Tasks)
	if !ok {
		t.Fatal("Of on a listing did not answer a listing of tasks")
	}
	if len(crossed.Tasks) != 2 {
		t.Fatalf("%d tasks crossed, want 2", len(crossed.Tasks))
	}
	if !slices.Equal(crossed.InFocus, []int{2}) {
		t.Errorf("in_focus = %v, want [2]", crossed.InFocus)
	}
}

func TestOf_saysNoFocusWhenAnAnswerSaysNothingAboutIt(t *testing.T) {
	crossed, ok := Of(TaskList, Answer{Tasks: []task.Task{builtTask(t, 1, "First")}}).(Tasks)
	if !ok {
		t.Fatal("Of on a listing did not answer a listing of tasks")
	}
	if len(crossed.InFocus) != 0 {
		t.Errorf("in_focus = %v, want nothing", crossed.InFocus)
	}
}

func list(numbers ...int) *[]int { return &numbers }

func TestOf_answersACardWithoutTheFocusWhenNobodyAskedAboutIt(t *testing.T) {
	one := builtTask(t, 7, "The one")
	if _, ok := Of(TaskShow, Answer{Task: &one}).(Task); !ok {
		t.Errorf("Of on a card = %T, want Task", Of(TaskShow, Answer{Task: &one}))
	}
}

func TestRenderText_saysNothingIsInFocus(t *testing.T) {
	got := RenderText(TaskUnfocus, Answer{InFocus: list()})
	if got != "nothing is in focus\n" {
		t.Errorf("RenderText = %q, want it to say nothing is in focus", got)
	}
}

func TestRenderText_marksTheTaskInFocusInAListing(t *testing.T) {
	tasks := []task.Task{builtTask(t, 1, "First"), builtTask(t, 2, "Second")}

	marked := RenderText(TaskList, Answer{Tasks: tasks, InFocus: list(2)})
	if !strings.Contains(marked, "Second") || !strings.Contains(marked, task.InFocusMark) {
		t.Errorf("the listing does not mark Second:\n%s", marked)
	}
	if got := strings.Count(marked, task.InFocusMark); got != 1 {
		t.Errorf("%d lines are marked, want exactly one:\n%s", got, marked)
	}

	if plain := RenderText(TaskList, Answer{Tasks: tasks}); strings.Contains(plain, task.InFocusMark) {
		t.Errorf("a listing asked for without the focus marks a line:\n%s", plain)
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

func TestRenderText_saysNothingIsInFocusWithAnEmptyList(t *testing.T) {
	got := RenderText(TaskFocus, Answer{Held: []task.Task{}, InFocus: list()})
	if got != "nothing is in focus\n" {
		t.Errorf("RenderText = %q, want it to say nothing is in focus", got)
	}
}

func TestRenderText_listsEveryTaskInFocus(t *testing.T) {
	first := builtTask(t, 1, "First")
	second := builtTask(t, 2, "Second")
	got := RenderText(TaskFocus, Answer{Held: []task.Task{first, second}, InFocus: list(1, 2)})
	for _, want := range []string{"First", "Second"} {
		if !strings.Contains(got, want) {
			t.Errorf("the answer does not carry %q:\n%s", want, got)
		}
	}
	if marks := strings.Count(got, task.InFocusMark); marks != 2 {
		t.Errorf("%d lines are marked, want both:\n%s", marks, got)
	}
}

func TestOfAnswer_carriesTheStepsOfATaskShown(t *testing.T) {
	var a task.AsTold
	a.Numbered(1)
	a.Titled("Ship it")
	a.Step("Write the tests", true)
	shown := a.Task()
	said, ok := Of(TaskShow, Answer{Task: &shown}).(Task)
	if !ok {
		t.Fatalf("OfAnswer gave %T, want a Task", Of(TaskShow, Answer{Task: &shown}))
	}
	if len(said.Steps) != 1 || said.Steps[0] != (task.Step{Title: "Write the tests", Done: true}) {
		t.Errorf("Steps = %+v, want the one done", said.Steps)
	}
}

func TestOfAnswer_aListingLeavesWhatATaskSaysToBeShown(t *testing.T) {
	var one, two task.AsTold
	one.Numbered(1)
	one.Titled("Ship it")
	one.Described("everything done so far")
	two.Numbered(2)
	two.Titled("Write the tests")
	tasks := []task.Task{one.Task(), two.Task()}
	for name, asked := range map[string]struct {
		op     Operation
		answer Answer
	}{"listing": {TaskList, Answer{Tasks: tasks}}, "search": {TaskSearch, Answer{Matched: tasks}}} {
		encoded, err := json.Marshal(Of(asked.op, asked.answer))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		var doc struct {
			Tasks []map[string]any `json:"tasks"`
		}
		if err := json.Unmarshal(encoded, &doc); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(doc.Tasks) != 2 {
			t.Fatalf("%s carries %d tasks", name, len(doc.Tasks))
		}
		for i, want := range []bool{true, false} {
			row := doc.Tasks[i]
			if _, carried := row["description"]; carried {
				t.Errorf("%s: task %v carries its description", name, row["number"])
			}
			if _, carried := row["sub_tasks"]; carried {
				t.Errorf("%s: task %v carries what hangs under it", name, row["number"])
			}
			if row["described"] != want || row["title"] == nil {
				t.Errorf("%s: task %v reads %v", name, row["number"], row)
			}
		}
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
