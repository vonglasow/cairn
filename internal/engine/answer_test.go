package engine

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/migration"
	"cairn/internal/objective"
	"cairn/internal/task"
)

// TestOfTask_carriesEveryFieldAScriptReads pins the shape a script reads for a
// task: the field names are the wire, and a script that reads one has no way
// to notice a change unless every one of them is checked here.
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
	a.Under(2)
	a.Serves(3)
	a.Titled("Write the spec")
	a.Described("in full")
	a.InColumn(domain.Todo)
	a.Carries(domain.High)
	a.Labelled("spec")
	a.Labelled("team")
	a.LabelsTold()
	a.Due(deadline)
	a.Estimated(estimate)
	a.Spent(spent)
	built := a.Task()
	progress := domain.Progress{SubTasks: domain.Tally{Total: 4, Done: 1}}

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
		"parent":      float64(2),
		"objective":   float64(3),
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
	sub, ok := prog["sub_tasks"].(map[string]any)
	if !ok || sub["total"] != float64(4) || sub["done"] != float64(1) {
		t.Errorf("progress.sub_tasks = %v, want {total:4 done:1}", prog["sub_tasks"])
	}
	if _, present := prog["tasks"]; present {
		t.Errorf("progress carries a count of tasks it does not have: %v", prog)
	}

	t.Run("a count of tasks with none under it does not carry sub-tasks too", func(t *testing.T) {
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
		if _, present := prog["sub_tasks"]; present {
			t.Errorf("progress carries a count of sub-tasks it does not have: %v", prog)
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
		oa.NotDescribed()
		oa.NotLabelled()
		oa.LabelsTold()
		oa.NotDue()
		oa.NotEstimated()
		obj := oa.Objective()
		backup := domain.Backup{Number: 2}
		held := domain.Held{Tasks: 3, Objectives: 1}
		var ca migration.AsTold
		ca.Numbered(1)
		ca.Named("add labels")
		ca.Does("ALTER TABLE tasks ADD COLUMN labels")
		ca.NotUndoable()
		ca.NothingLost()
		ca.Applied(time.Time{})
		change := ca.Migration()

		cases := []struct {
			name string
			in   Answer
			want any
		}{
			{"a task outranks a list", Answer{Task: &built, Progress: progress}, OfTask(built, progress)},
			{"a list of tasks", Answer{Tasks: []task.Task{built}}, Tasks{Tasks: []Task{OfTask(built, domain.Progress{})}}},
			{"what a search found, the same shape as a list of tasks", Answer{Matched: []task.Task{built}}, Tasks{Tasks: []Task{OfTask(built, domain.Progress{})}}},
			{"an objective", Answer{Objective: &obj, Progress: progress}, Objective{
				Shape: objective.Shape{
					Number: 5, Title: "Ship it", State: "active", Priority: "normal",
					Description: "", Labels: []string{}, Deadline: "", Estimate: "",
				},
				Tasks:    []int{},
				Progress: &domain.ProgressShape{SubTasks: &domain.Tally{Total: 4, Done: 1}, Percent: 25},
			}},
			{"a list of objectives", Answer{Objectives: []objective.Objective{obj}}, Objectives{Objectives: []Objective{OfObjective(obj, nil, domain.Progress{})}}},
			{"a backup with what it holds", Answer{Backup: &backup, Holds: &held}, domain.OfBackup(backup, &held)},
			{"a list of backups", Answer{Backups: []domain.Backup{backup}}, Kept{Backups: []domain.BackupShape{domain.OfBackup(backup, nil)}}},
			{"a change", Answer{Change: &change}, migration.OfMigration(change)},
			{"a list of changes", Answer{Changes: []migration.Migration{change}}, Ledger{Changes: []migration.Shape{migration.OfMigration(change)}}},
			{"a rollback names what it undid, not a change in full", Answer{RolledBack: &change}, migration.OfRolledBack(change)},
			{"a number deleted", Answer{Deleted: &one}, Deleted{Deleted: one}},
			{"a number put back", Answer{PutBack: &putBack}, PutBack{PutBack: putBack}},
			{"nothing waiting reads as an empty list, not null", Answer{Applied: []string{}}, Applied{Applied: []string{}}},
			{"names of what was applied", Answer{Applied: []string{"add labels"}}, Applied{Applied: []string{"add labels"}}},
			{"what the store holds, on its own", Answer{Holds: &held}, held},
			{"nothing answered at all", Answer{}, nil},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				got, err := json.Marshal(Of(c.in))
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

// TestRenderText_rendersEveryFieldAnswerCarries pins what a person reads for
// every field RenderText resolves on, the same way the table above pins what
// a script reads: each field of Answer takes RenderText down a different
// branch, and a list carries two shapes of its own — empty and not — that
// only differ in whether it says so.
func TestRenderText_rendersEveryFieldAnswerCarries(t *testing.T) {
	var ta task.AsTold
	ta.Numbered(1)
	ta.Titled("Write the spec")
	ta.InColumn(domain.Todo)
	ta.Carries(domain.Normal)
	ta.NotDescribed()
	ta.NotLabelled()
	ta.LabelsTold()
	ta.NotDue()
	ta.NotEstimated()
	ta.NothingSpent()
	ta.UnderNothing()
	ta.ServesNothing()
	built := ta.Task()
	progress := domain.Progress{SubTasks: domain.Tally{Total: 4, Done: 1}}

	// sub hangs under built, so a listing would indent it — a search's flat
	// rendering must not, which is what tells the two apart.
	var sa task.AsTold
	built.Tell(&sa)
	sa.Numbered(2)
	sa.Titled("Read it first")
	sa.Under(1)
	sub := sa.Task()

	var oa objective.AsTold
	oa.Numbered(5)
	oa.Titled("Ship it")
	oa.Standing(objective.Active)
	oa.Carries(domain.Normal)
	oa.NotDescribed()
	oa.NotLabelled()
	oa.LabelsTold()
	oa.NotDue()
	oa.NotEstimated()
	obj := oa.Objective()

	backup := domain.Backup{Number: 2}
	held := domain.Held{Tasks: 3, Objectives: 1}

	var ca migration.AsTold
	ca.Numbered(1)
	ca.Named("add labels")
	ca.Does("ALTER TABLE tasks ADD COLUMN labels")
	ca.NotUndoable()
	ca.NothingLost()
	ca.Applied(time.Time{})
	change := ca.Migration()

	deleted := 7
	putBack := 9
	deleting := Operation{Thing: "task"}

	cases := []struct {
		name string
		op   Operation
		in   Answer
		want string
	}{
		{"a task outranks a list", Operation{}, Answer{Task: &built, Progress: progress}, task.Text(built, progress) + task.Under(nil)},
		{"an empty list of tasks says so", Operation{}, Answer{Tasks: []task.Task{}}, "there is no task\n"},
		{"a list of tasks", Operation{}, Answer{Tasks: []task.Task{built}}, task.Tree([]task.Task{built}, task.NothingInFocus)},
		{"a search that found nothing says so", Operation{}, Answer{Matched: []task.Task{}}, "there is no task\n"},
		{
			"what a search found is listed flat, not as a tree",
			Operation{}, Answer{Matched: []task.Task{built, sub}},
			task.Line(built) + "\n" + task.Line(sub) + "\n",
		},
		{"an objective", Operation{}, Answer{Objective: &obj, Progress: progress}, objective.Text(obj, progress) + objective.Serving(nil)},
		{"an empty list of objectives says so", Operation{}, Answer{Objectives: []objective.Objective{}}, "there is no objective\n"},
		{"a list of objectives", Operation{}, Answer{Objectives: []objective.Objective{obj}}, objective.List([]objective.Objective{obj})},
		{"a backup with what it holds", Operation{}, Answer{Backup: &backup, Holds: &held}, backupLine(backup, held)},
		{"an empty list of backups says so", Operation{}, Answer{Backups: []domain.Backup{}}, "there is no backup\n"},
		{"a list of backups", Operation{}, Answer{Backups: []domain.Backup{backup}}, backupsText([]domain.Backup{backup})},
		{"a rollback names what it undid, not a change in full", Operation{}, Answer{RolledBack: &change}, migration.RolledBackText(change)},
		{"a change in full", Operation{}, Answer{Change: &change}, migration.Text(change)},
		{"a list of changes", Operation{}, Answer{Changes: []migration.Migration{change}, Holds: &held}, migration.List([]migration.Migration{change}, held)},
		{"a number deleted, named by the operation", deleting, Answer{Deleted: &deleted}, "task 7 was deleted\n"},
		{"a number put back", Operation{}, Answer{PutBack: &putBack}, "backup 9 was put back\n"},
		{"nothing waiting says so", Operation{}, Answer{Applied: []string{}}, "there is nothing waiting to be applied\n"},
		{"names of what was applied", Operation{}, Answer{Applied: []string{"add labels"}}, "applied \"add labels\"\n"},
		{"nothing answered at all", Operation{}, Answer{}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := RenderText(c.op, c.in); got != c.want {
				t.Errorf("RenderText(%+v) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// What crosses about the focus: the task being worked on and its number, in one
// shape.
func TestOf_saysWhichTaskIsInFocusAndHandsItOver(t *testing.T) {
	one := builtTask(t, 7, "The one")
	held, ok := Of(Answer{Task: &one, InFocus: of(7)}).(Focus)
	if !ok {
		t.Fatalf("Of on the focus = %T, want Focus", Of(Answer{Task: &one, InFocus: of(7)}))
	}
	if held.InFocus == nil || *held.InFocus != 7 {
		t.Errorf("in_focus = %v, want 7", held.InFocus)
	}
	if held.Task == nil || held.Task.Number != 7 {
		t.Errorf("the task handed over = %v, want task 7", held.Task)
	}
}

// Nothing in focus is the same shape with both fields empty, so a machine reads
// one field of one shape either way. Null rather than zero: zero is a number no
// task carries, and a machine testing for it would have to know that.
func TestOf_saysNothingIsInFocusAsTheSameShape(t *testing.T) {
	none, ok := Of(Answer{InFocus: of(task.NothingInFocus)}).(Focus)
	if !ok {
		t.Fatalf("Of with nothing held = %T, want Focus", Of(Answer{InFocus: of(task.NothingInFocus)}))
	}
	if none.InFocus != nil {
		t.Errorf("in_focus = %v, want null", none.InFocus)
	}
	if none.Task != nil {
		t.Errorf("a task crossed with nothing in focus: %v", none.Task)
	}
}

// A listing says which task is in focus once, at the top. Said on every row the
// rows can disagree, and a reader has to pick one to believe.
func TestOf_saysTheFocusOnceInAListing(t *testing.T) {
	first := builtTask(t, 1, "First")
	second := builtTask(t, 2, "Second")

	crossed, ok := Of(Answer{Tasks: []task.Task{first, second}, InFocus: of(2)}).(Tasks)
	if !ok {
		t.Fatal("Of on a listing did not answer a listing of tasks")
	}
	if len(crossed.Tasks) != 2 {
		t.Fatalf("%d tasks crossed, want 2", len(crossed.Tasks))
	}
	if crossed.InFocus == nil || *crossed.InFocus != 2 {
		t.Errorf("in_focus = %v, want 2", crossed.InFocus)
	}
}

// A listing asked for without the focus says none, rather than saying the task
// numbered zero is being worked on.
func TestOf_saysNoFocusWhenAnAnswerSaysNothingAboutIt(t *testing.T) {
	crossed, ok := Of(Answer{Tasks: []task.Task{builtTask(t, 1, "First")}}).(Tasks)
	if !ok {
		t.Fatal("Of on a listing did not answer a listing of tasks")
	}
	if crossed.InFocus != nil {
		t.Errorf("in_focus = %v, want null", crossed.InFocus)
	}
}

// A card asked for by number is a task and nothing else: it was not asked about
// the focus, so it does not answer about it.
func TestOf_answersACardWithoutTheFocusWhenNobodyAskedAboutIt(t *testing.T) {
	one := builtTask(t, 7, "The one")
	if _, ok := Of(Answer{Task: &one}).(Task); !ok {
		t.Errorf("Of on a card = %T, want Task", Of(Answer{Task: &one}))
	}
}

// Nothing in focus is an answer a person reads, not an empty one.
func TestRenderText_saysNothingIsInFocus(t *testing.T) {
	got := RenderText(Operation{}, Answer{InFocus: of(task.NothingInFocus)})
	if got != "nothing is in focus\n" {
		t.Errorf("RenderText = %q, want it to say nothing is in focus", got)
	}
}

// A listing marks the task in focus, and a listing asked for without it marks
// nothing — the same claim as the JSON above, on the other rendering.
func TestRenderText_marksTheTaskInFocusInAListing(t *testing.T) {
	tasks := []task.Task{builtTask(t, 1, "First"), builtTask(t, 2, "Second")}

	marked := RenderText(Operation{}, Answer{Tasks: tasks, InFocus: of(2)})
	if want := "Second" + task.InFocusMark; !strings.Contains(marked, want) {
		t.Errorf("the listing does not carry %q:\n%s", want, marked)
	}
	if got := strings.Count(marked, task.InFocusMark); got != 1 {
		t.Errorf("%d lines are marked, want exactly one:\n%s", got, marked)
	}

	if plain := RenderText(Operation{}, Answer{Tasks: tasks}); strings.Contains(plain, task.InFocusMark) {
		t.Errorf("a listing asked for without the focus marks a line:\n%s", plain)
	}
}

// builtTask builds a numbered task, for the assertions above.
func builtTask(t *testing.T, number int, title string) task.Task {
	t.Helper()
	built, err := task.NewTask(title)
	if err != nil {
		t.Fatalf("NewTask(%q): %v", title, err)
	}
	a := &task.AsTold{}
	built.Tell(a)
	a.Numbered(number)
	return a.Task()
}
