package render

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cairn/internal/engine"
	"cairn/internal/engine/backup"
	"cairn/internal/engine/item"
	"cairn/internal/engine/migration"
	"cairn/internal/engine/objective"
	"cairn/internal/engine/series"
	"cairn/internal/engine/task"
	"cairn/internal/engine/workspace"
)

func of[T any](v T) *T { return &v }

func TestText_rendersEveryFieldAnswerCarries(t *testing.T) {
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

	backupShow := engine.Operation{Answers: engine.BackupShape{}}
	migrationShow := engine.Operation{Answers: migration.Shape{}}

	cases := []struct {
		name string
		op   engine.Operation
		in   engine.Answer
		want string
	}{
		{"a task", engine.TaskShow, engine.Answer{Task: &built, Progress: progress}, task.Text(built, progress, true)},
		{"an empty list of tasks says so", engine.TaskList, engine.Answer{Tasks: []task.Task{}}, "there is no task\n"},
		{"a list of tasks", engine.TaskList, engine.Answer{Tasks: []task.Task{built}}, task.Lines([]task.Task{built}, task.Marks{})},
		{"a search that found nothing says so", engine.TaskSearch, engine.Answer{Matched: []task.Task{}}, "there is no task\n"},
		{"what a search found is listed one a line", engine.TaskSearch, engine.Answer{Matched: []task.Task{built}}, task.Line(built) + "\n"},
		{"an objective", engine.ObjectiveShow, engine.Answer{Objective: &obj, Progress: progress}, objective.Text(obj, progress, objective.Review{}, objective.Family{}) + objective.Serving(nil)},
		{"an empty list of objectives says so", engine.ObjectiveList, engine.Answer{Objectives: []objective.Objective{}}, "there is no objective\n"},
		{"a list of objectives", engine.ObjectiveList, engine.Answer{Objectives: []objective.Objective{obj}}, objective.List([]objective.Objective{obj}, nil, nil)},
		{"a backup with what it holds", backupShow, engine.Answer{Backup: &made, Holds: &held}, backupLine(made, held)},
		{"an empty list of backups says so", engine.BackupList, engine.Answer{Backups: []backup.Backup{}}, "there is no backup\n"},
		{"a list of backups", engine.BackupList, engine.Answer{Backups: []backup.Backup{made}}, backupsText([]backup.Backup{made})},
		{"a rollback names what it undid, not a change in full", engine.MigrationRollback, engine.Answer{RolledBack: &change}, migration.RolledBackText(change)},
		{"a change in full", migrationShow, engine.Answer{Change: &change}, migration.Text(change)},
		{"a list of changes", engine.MigrationList, engine.Answer{Changes: []migration.Migration{change}, Holds: &held}, migration.List([]migration.Migration{change}, held)},
		{"a number deleted, named by the operation", engine.TaskDelete, engine.Answer{Deleted: of(7)}, "task 7 was deleted\n"},
		{"a number restored", engine.BackupRestore, engine.Answer{Restored: of(9)}, "backup 9 was restored\n"},
		{"nothing waiting says so", engine.MigrationApply, engine.Answer{Applied: []string{}}, "there is nothing waiting to be applied\n"},
		{"names of what was applied", engine.MigrationApply, engine.Answer{Applied: []string{"add labels"}}, "applied \"add labels\"\n"},
		{"a sweep says how many", engine.TaskArchive, engine.Answer{Archived: of(7)}, "7 task(s) archived\n"},
		{"one archived by number shows the card", engine.TaskArchive, engine.Answer{Archived: of(1), Task: &built}, task.Text(built, item.Progress{}, true)},
		{"how many were labelled", engine.TaskLabel, engine.Answer{Labelled: of(7)}, "7 task(s) labelled\n"},
		{"nothing in focus says so", engine.TaskUnfocus, engine.Answer{InFocus: of([]int{})}, "nothing is in focus\n"},
		{"nothing declared, nothing said", engine.Operation{}, engine.Answer{Tasks: []task.Task{built}}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Text(c.op, c.in); got != c.want {
				t.Errorf("Text(%+v) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestText_showsWhatATerminalWouldObey(t *testing.T) {
	var a task.AsTold
	a.Numbered(1)
	a.Titled("look \x1b[8m here")
	a.Described("bell\x07 and \x1b]52;c;aGk=\x07")
	kept := a.Task()
	if got := Text(engine.TaskShow, engine.Answer{Task: &kept}); strings.ContainsAny(got, "\x1b\x07") {
		t.Errorf("Text = %q, want nothing a terminal would obey", got)
	}
}

func TestText_writesTheContextAndMarksTheDeleted(t *testing.T) {
	if got := contextText("/notes/1"); got != "\ncontext  /notes/1\n" {
		t.Errorf("a context reads %q", got)
	}
	got := namedAsShown(map[int]string{1: "kept", 2: "gone"}, map[int]bool{2: true})
	if got[1] != "kept" || got[2] != "gone"+engine.DeletedMark {
		t.Errorf("the names read %v", got)
	}
}

func TestText_saysTheRecentFoldersUnderTheWorkspaces(t *testing.T) {
	said := Text(engine.WorkspaceList, engine.Answer{Workspaces: &workspace.Workspaces{
		Workspaces: []workspace.Workspace{{Name: "work", Folder: "/w", Default: true, Colour: "green"}},
		Recent:     []workspace.RecentFolder{{Store: "/l/cairn.db", Folder: "/l", Opened: time.Now()}, {Store: "/t/x.db", NotFound: true}},
	}})
	for _, want := range []string{"work  /w  default  green\n", "\nrecent folders, opened by path:\n", "  /l  opened today\n", "  /t/x.db  not found now\n"} {
		if !strings.Contains(said, want) {
			t.Errorf("workspace list says:\n%s\nwant it saying %q", said, want)
		}
	}
}

func TestSeriesShown_saysTheTimeSpentAndADeletion(t *testing.T) {
	said := seriesShownText(engine.Series{
		Number: 1, Title: "Review", State: "active", Rule: "every day", Says: "every day", TimeSpent: "1h",
		Occurrences: []series.Occurrence{{Day: "2026-11-09", Task: 3, Is: "done", TimeSpent: "30m"}},
		Acts:        []series.Act{{Did: "deleted", Became: "keep", At: "2026-11-09T10:00:00Z"}},
	})
	for _, want := range []string{"missed · 1h\n", "done  30m\n", "deleted, choosing keep"} {
		if !strings.Contains(said, want) {
			t.Errorf("%q is not in\n%s", want, said)
		}
	}
}

func TestSeriesTexts_sayOnlyWhatIsThere(t *testing.T) {
	for _, c := range []struct{ got, want string }{
		{series.FollowedText([]series.MadeNext{{Series: 1, Task: 2, Day: "2026-11-10"}}), "↻1 made #2 for Tue 10 Nov\n"},
		{engine.NextSaid(engine.Series{}), "none"},
		{engine.NextSaid(engine.Series{Ended: true, Next: []string{"2026-11-10"}}), "Tue 10 Nov"},
		{engine.TemplateSaid(series.Template{Title: "Review", Priority: "normal"}), "Review · normal"},
		{engine.TemplateSaid(series.Template{Title: "Review", Priority: "normal", Labels: []string{"a", "b"}, Steps: []string{"x"}, Estimate: "15m"}), "Review · normal · a, b · 1 step · 15m"},
		{seriesDeletedText(series.Deleted{Deleted: 1}), "↻1 was deleted\n"},
		{seriesDeletedText(series.Deleted{Deleted: 1, TasksDeleted: []int{2}, Kept: []int{1}}), "↻1 was deleted · #2 deleted · 1 task kept as ordinary tasks\n"},
		{seriesActedText("stop", engine.Answer{SeriesShown: &engine.Series{Number: 1}}), "↻1 stopped; nothing more is made\n"},
	} {
		if c.got != c.want {
			t.Errorf("said %q, want %q", c.got, c.want)
		}
	}
	if said := series.OccursText(series.Occurs{Series: 1, Says: "every day", Day: "2026-11-10"}); !strings.Contains(said, "\nday      Tue 10 Nov\n") {
		t.Errorf("an occurrence with no time reads %q", said)
	}
}

func TestDaysSince_countsTheDaysOnTheCalendar(t *testing.T) {
	now := time.Date(2026, 10, 5, 9, 0, 0, 0, time.Local)
	for _, c := range []struct {
		since time.Time
		want  string
	}{
		{time.Date(2026, 10, 5, 0, 30, 0, 0, time.Local), "today"},
		{time.Date(2026, 10, 4, 23, 30, 0, 0, time.Local), "1 day"},
		{time.Date(2026, 10, 3, 9, 30, 0, 0, time.Local), "2 days"},
		{time.Date(2026, 9, 26, 12, 0, 0, 0, time.Local), "9 days"},
		{time.Date(2026, 3, 20, 12, 0, 0, 0, time.Local), "199 days"},
	} {
		if got := daysSince(c.since, now); got != c.want {
			t.Errorf("daysSince(%v) = %q, want %q", c.since, got, c.want)
		}
	}
}

type asked struct {
	given  map[string]string
	listed map[string][]string
}

func given(pairs ...string) asked {
	a := asked{given: map[string]string{}, listed: map[string][]string{}}
	for i := 0; i+1 < len(pairs); i += 2 {
		a.given[pairs[i]] = pairs[i+1]
	}
	return a
}

func (a asked) with(name string, values ...string) asked {
	a.listed[name] = append(a.listed[name], values...)
	return a
}

func said(t *testing.T, path string, op engine.Operation, a asked) string {
	t.Helper()
	answer, err := engine.Do(op, path, engine.Input{Given: a.given, Listed: a.listed, Confirmed: true})
	if err != nil {
		t.Fatalf("%s: %v", op.Line(), err)
	}
	return Text(op, answer)
}

func viewStore(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cairn.db")
	for _, a := range []struct {
		op engine.Operation
		a  asked
	}{
		{engine.ObjectiveNew, given("title", "Home automation")},
		{engine.ObjectiveNew, given("title", "Garden")},
		{engine.TaskNew, given("title", "Wiring diagram").with("objective", "1").with("label", "doc")},
		{engine.TaskNew, given("title", "Read the sensors").with("objective", "1")},
		{engine.TaskNew, given("title", "Upgrade the printer")},
		{engine.TaskEdit, given("number", "1").with("blocker", "3").with("link", "3")},
		{engine.TaskEdit, given("number", "2", "state", "done").with("blocker", "1")},
		{engine.ViewNew, given("name", "Docs").with("label", "doc")},
		{engine.ViewNew, given("name", "Home").with("objective", "1", "2")},
	} {
		said(t, path, a.op, a.a)
	}
	return path
}

func TestText_saysAViewAndWhatLiesOutsideIt(t *testing.T) {
	path := viewStore(t)
	show := named(t, "view show")
	for _, c := range []struct {
		op   engine.Operation
		a    asked
		want []string
	}{
		{engine.TaskList, given("view", "Docs"), []string{"#1  Wiring diagram", "outside this view, never counted:\n", "#3  Upgrade the printer  · backlog · blocks #1 · linked to #1\n", "#2  Read the sensors  · done · waits on #1\n"}},
		{show, given("view", "Docs"), []string{"view     Docs  (⌘1)\n", "holds    label doc\n", "keeps    1 of 3 tasks · 0 done\n", "outside  #2 waits on #1 · #3 blocks #1, linked to #1\n"}},
		{show, given("view", "Home"), []string{"objective #1 Home automation  1/2 · 50% (the store's)\n", "objective #2 Garden  nothing serves it yet\n"}},
		{engine.ViewList, given(), []string{"#  name  holds", "1  Docs  label doc"}},
		{engine.ViewEdit, given("view", "Home", "name", "House"), []string{"changed view House (was Home) · ⌘2 · keeps 0 of 3 tasks\n"}},
		{engine.TaskList, given("count", "state"), []string{"backlog  2\ndone     1\n3 task(s) in all\n"}},
		{engine.TaskSearch, given("limit", "1", "offset", "1").with("named", "the"), []string{"tasks 2 to 2 of 2\n"}},
		{engine.TaskSearch, given("offset", "5").with("named", "the"), []string{"no task past 5, of 2\n"}},
		{engine.LabelDelete, given("label", "doc"), []string{"warning: 1 view holds the label doc; it narrows nothing there now\n"}},
		{engine.TaskList, given("view", "Docs"), []string{"every condition view Docs holds is gone, so it keeps nothing: label doc\n  cairn view edit Docs, or cairn view delete Docs\n"}},
		{engine.ViewList, given(), []string{"label doc, gone"}},
		{engine.TaskMove, given("number", "3", "state", "todo"), []string{"task 3 moved from backlog to todo\n"}},
		{engine.ViewDelete, given("view", "Docs"), []string{"deleted view Docs · no task changed\n"}},
	} {
		if got := said(t, path, c.op, c.a); !containsAll(got, c.want) {
			t.Errorf("%s %v reads %q, want %q", c.op.Line(), c.a, got, c.want)
		}
	}
}

func containsAll(got string, want []string) bool {
	for _, w := range want {
		if !strings.Contains(got, w) {
			return false
		}
	}
	return true
}

func TestText_saysWhatASeriesDid(t *testing.T) {
	s := &engine.Series{Number: 1, Title: "Review", State: "active", Says: "every week", Next: []string{"2026-11-16"}, Open: []int{2}}
	for _, c := range []struct {
		op   engine.Operation
		in   engine.Answer
		want string
	}{
		{engine.SeriesStop, engine.Answer{SeriesShown: s}, "↻1 stopped; #2 stays open, and nothing more is made\n"},
		{engine.SeriesResume, engine.Answer{SeriesShown: s, Followed: []series.MadeNext{{Series: 1, Task: 3, Day: "2026-11-16"}}}, "↻1 resumed · next Mon 16 Nov, made as #3\n"},
		{engine.SeriesResume, engine.Answer{SeriesShown: s}, "↻1 resumed; #2 is still open, and the next comes when it is done or skipped\n"},
		{engine.SeriesResume, engine.Answer{SeriesShown: &engine.Series{Number: 2}}, "↻2 resumed; it has ended, and nothing more is made\n"},
		{engine.SeriesEdit, engine.Answer{SeriesShown: s}, "changed ↻1 Review\nrule      every week\ntemplate  "},
		{engine.SeriesList, engine.Answer{SeriesKept: &engine.SeriesListed{}}, "there is no series\n"},
		{engine.SeriesList, engine.Answer{SeriesKept: &engine.SeriesListed{Series: []engine.Series{*s}}}, "↻1  Review  · active  · next Mon 16 Nov\n    every week · 0 done · 0 skipped · 0 missed\n"},
		{engine.Operation{Answers: task.Steps{}}, engine.Answer{Task: of(must(task.NewTask("Bare")))}, "has no steps\n"},
	} {
		if got := Text(c.op, c.in); !strings.Contains(got, c.want) {
			t.Errorf("%s reads %q, want %q", c.op.Line(), got, c.want)
		}
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func named(t *testing.T, line string) engine.Operation {
	t.Helper()
	for _, op := range append(engine.All(), engine.ForAPersonOnly()...) {
		if op.Line() == line {
			return op
		}
	}
	t.Fatalf("no operation is %q", line)
	return engine.Operation{}
}
