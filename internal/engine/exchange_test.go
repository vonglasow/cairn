package engine

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
)

func ran(t *testing.T, path string, asks ...func() (Operation, Input)) {
	t.Helper()
	for _, ask := range asks {
		op, in := ask()
		if _, err := Do(op, path, in); err != nil {
			t.Fatalf("%s %v: %v", op.Line(), in.Given, err)
		}
	}
}

func asked(op Operation, pairs ...string) func() (Operation, Input) {
	return func() (Operation, Input) {
		in := Input{Given: map[string]string{}, Listed: map[string][]string{}}
		for i := 0; i+1 < len(pairs); i += 2 {
			if pairs[i] == "label" || pairs[i] == "objective" || pairs[i] == "blocker" || pairs[i] == "link" {
				in.Listed[pairs[i]] = append(in.Listed[pairs[i]], pairs[i+1])
				continue
			}
			in.Given[pairs[i]] = pairs[i+1]
		}
		return op, in
	}
}

func TestTaskExport_writesTheTasksAndWhatTheyServe(t *testing.T) {
	path := storeAt(t)
	ran(t, path,
		asked(ObjectiveNew, "title", "Release 1.0"),
		asked(ObjectiveNew, "title", "Served by nothing"),
		asked(ObjectiveNew, "title", "Gone"),
		func() (Operation, Input) {
			return ObjectiveDelete, Input{Confirmed: true, Given: map[string]string{"number": "3"}}
		},
		asked(TaskNew, "title", "Write the spec", "objective", "1"),
		asked(TaskNew, "title", "Old idea"),
		asked(TaskArchive, "number", "2"),
	)
	to := filepath.Join(t.TempDir(), "elsewhere", "tasks.json")
	answer, err := Do(TaskExport, path, Input{Given: map[string]string{"to": to}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if got := RenderText(TaskExport, answer); got != to+" holds 2 task(s) and 2 objective(s)\n" {
		t.Errorf("it says %q", got)
	}
	if shape, ok := Of(TaskExport, answer).(Exported); !ok || shape.Exported != to || shape.Holds.Tasks != 2 {
		t.Errorf("a machine reads %#v", Of(TaskExport, answer))
	}
	written, err := os.ReadFile(to)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"cairn": "tasks"`, `"version": 7`, `"Write the spec"`, `"Old idea"`, `"archived": true`, `"Release 1.0"`, `"Served by nothing"`} {
		if !strings.Contains(string(written), want) {
			t.Errorf("the file holds no %s:\n%s", want, written)
		}
	}
	if strings.Contains(string(written), `"Gone"`) {
		t.Error("an objective deleted was written out")
	}
	if _, err := Do(TaskExport, path, Input{Given: map[string]string{"to": to}}); !errors.Is(err, os.ErrExist) {
		t.Errorf("exporting onto it again answered %v", err)
	}
	if TaskExport.Level != Write || TaskImport.Level != Write {
		t.Errorf("export is level %v and import %v", TaskExport.Level, TaskImport.Level)
	}
}

func TestTaskImport_addsTheTasksAsNewOnes(t *testing.T) {
	from := storeAt(t)
	ran(t, from,
		asked(ObjectiveNew, "title", "Release 1.0"),
		asked(TaskNew, "title", "Ship it", "objective", "1", "priority", "high", "label", "release", "deadline", "2026-12-01", "estimate", "2h", "desc", "All of it"),
		asked(StepNew, "number", "1", "title", "Write the spec", "done", "true"),
		asked(TaskNew, "title", "Review", "state", "todo"),
		asked(TaskEdit, "number", "1", "blocker", "2"),
		asked(TaskNew, "title", "Old idea"),
		asked(TaskArchive, "number", "3"),
	)
	file := filepath.Join(t.TempDir(), "tasks.json")
	ran(t, from, asked(TaskExport, "to", file))

	path := storeAt(t)
	ran(t, path, asked(TaskNew, "title", "Already here"), asked(ObjectiveNew, "title", "Already served"))
	answer, err := Do(TaskImport, path, Input{Given: map[string]string{"from": file}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if got := RenderText(TaskImport, answer); got != "3 task(s) and 1 objective(s) were added from "+file+"\n" {
		t.Errorf("it says %q", got)
	}
	if shape, ok := Of(TaskImport, answer).(Imported); !ok || shape.Imported != file || shape.Added.Tasks != 3 || shape.Added.Objectives != 1 {
		t.Errorf("a machine reads %#v", Of(TaskImport, answer))
	}
	show := func(n string) string {
		t.Helper()
		got, err := Do(TaskShow, path, Input{Given: map[string]string{"number": n}})
		if err != nil {
			t.Fatalf("task show %s: %v", n, err)
		}
		return RenderText(TaskShow, got)
	}
	shipped := show("2")
	for _, want := range []string{"Ship it", "high", "release", "2026-12-01", "2h", "All of it", "blocked by   #3", "objectives   #2", "(done)  1  Write the spec"} {
		if !strings.Contains(shipped, want) {
			t.Errorf("task 2 does not say %q:\n%s", want, shipped)
		}
	}
	if old := show("4"); !strings.Contains(old, "archived") {
		t.Errorf("the archived task reads:\n%s", old)
	}
	if here := show("1"); !strings.Contains(here, "Already here") {
		t.Errorf("task 1 became:\n%s", here)
	}
	ran(t, path, asked(TaskImport, "from", file))
	if listed, _ := Do(TaskList, path, Input{Given: map[string]string{"with-archived": "true"}}); len(listed.Tasks) < 7 {
		t.Errorf("read twice, the store lists %d tasks", len(listed.Tasks))
	}
}

func TestTaskImport_readsTheSubTasksOfAnOlderFileAsSteps(t *testing.T) {
	file := filepath.Join(t.TempDir(), "tasks.json")
	if err := os.WriteFile(file, []byte(`{"cairn": "tasks", "version": 1, "objectives": [], "tasks": [
		{"number": 1, "title": "Ship it", "state": "todo", "priority": "normal"},
		{"number": 2, "parent": 1, "title": "Write the spec", "state": "done", "priority": "normal"},
		{"number": 3, "parent": 1, "title": "Write the tests", "state": "todo", "priority": "normal"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	path := storeAt(t)
	answer, err := Do(TaskImport, path, Input{Given: map[string]string{"from": file}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if answer.Holds.Tasks != 1 {
		t.Errorf("%d tasks were added, want the one the others hung under", answer.Holds.Tasks)
	}
	want := []task.Step{{Title: "Write the tests"}, {Title: "Write the spec", Done: true}}
	if got := stepsOf(t, path, "1"); !slices.Equal(got, want) {
		t.Errorf("steps = %v, want %v", got, want)
	}
}

func TestTaskImport_refusesBeforeAddingAnything(t *testing.T) {
	path := storeAt(t)
	ran(t, path, asked(TaskNew, "title", "Already here"))
	dir := t.TempDir()
	files := map[string]string{
		"not json":       "not tasks\n",
		"not an export":  `{"something": "else"}`,
		"a later format": `{"cairn": "tasks", "version": 8, "tasks": []}`,
		"a broken parent": `{"cairn": "tasks", "version": 1, "objectives": [],
			"tasks": [{"number": 1, "title": "Kept", "state": "todo", "priority": "normal"},
			          {"number": 2, "parent": 9, "title": "Orphan", "state": "backlog", "priority": "normal"}]}`,
		"a broken blocker": `{"cairn": "tasks", "version": 1, "objectives": [],
			"tasks": [{"number": 1, "title": "Kept", "state": "blocked", "priority": "normal", "blocked_by": [7]}]}`,
		"a broken objective": `{"cairn": "tasks", "version": 1, "objectives": [],
			"tasks": [{"number": 1, "title": "Kept", "state": "todo", "priority": "normal", "objectives": [3]}]}`,
		"a bad state": `{"cairn": "tasks", "version": 1, "objectives": [],
			"tasks": [{"number": 1, "title": "Kept", "state": "nowhere", "priority": "normal"}]}`,
		"an empty title": `{"cairn": "tasks", "version": 1, "objectives": [],
			"tasks": [{"number": 1, "title": "", "state": "todo", "priority": "normal"}]}`,
		"an empty step": `{"cairn": "tasks", "version": 2, "objectives": [],
			"tasks": [{"number": 1, "title": "Kept", "state": "todo", "priority": "normal", "steps": [{"title": " ", "done": false}]}]}`,
	}
	for name, held := range files {
		t.Run(name, func(t *testing.T) {
			file := filepath.Join(dir, strings.ReplaceAll(name, " ", "-")+".json")
			if err := os.WriteFile(file, []byte(held), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := Do(TaskImport, path, Input{Given: map[string]string{"from": file}}); err == nil {
				t.Error("it was read in")
			}
			if listed, _ := Do(TaskList, path, Input{}); len(listed.Tasks) != 1 {
				t.Errorf("refused, it still added: %d tasks", len(listed.Tasks))
			}
		})
	}
	if _, err := Do(TaskImport, path, Input{Given: map[string]string{"from": filepath.Join(dir, "nowhere.json")}}); err == nil {
		t.Error("a file that is not there was read in")
	}
}

func TestTaskImport_namesTheTaskTheFileDoesNotHold(t *testing.T) {
	path := storeAt(t)
	file := filepath.Join(t.TempDir(), "orphan.json")
	held := `{"cairn": "tasks", "version": 1, "objectives": [],
		"tasks": [{"number": 1, "parent": 9, "title": "Orphan", "state": "backlog", "priority": "normal"}]}`
	if err := os.WriteFile(file, []byte(held), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Do(TaskImport, path, Input{Given: map[string]string{"from": file}})
	if err == nil || !strings.Contains(err.Error(), "task 9 is not in the file") {
		t.Errorf("it answered %v", err)
	}
}

func TestAnImportBringsTheLinksAtTheirNewNumbers(t *testing.T) {
	from := storeAt(t)
	ran(t, from,
		asked(TaskNew, "title", "Ship it"),
		asked(TaskNew, "title", "Write the tests"),
		asked(TaskNew, "title", "It does not hold", "link", "1", "link", "2"),
	)
	file := filepath.Join(t.TempDir(), "tasks.json")
	ran(t, from, asked(TaskExport, "to", file))
	path := storeAt(t)
	ran(t, path, asked(TaskNew, "title", "Already here"))
	if _, err := Do(TaskImport, path, Input{Given: map[string]string{"from": file}}); err != nil {
		t.Fatalf("Do: %v", err)
	}
	for n, want := range map[string][]int{"4": {2, 3}, "2": {4}, "3": {4}} {
		got, err := Do(TaskShow, path, Input{Given: map[string]string{"number": n}})
		if err != nil {
			t.Fatalf("task show %s: %v", n, err)
		}
		if text := RenderText(TaskShow, got); !strings.Contains(text, "linked       "+linkedLine(want)) {
			t.Errorf("task %s does not say linked to %v:\n%s", n, want, text)
		}
	}
	outside := filepath.Join(t.TempDir(), "outside.json")
	held := `{"cairn": "tasks", "version": 3, "objectives": [],
		"tasks": [{"number": 1, "title": "Kept", "state": "todo", "priority": "normal", "linked": [9]}]}`
	if err := os.WriteFile(outside, []byte(held), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Do(TaskImport, path, Input{Given: map[string]string{"from": outside}})
	if !errors.Is(err, ErrExportPointsOutside) || !strings.Contains(err.Error(), "task 9 is not in the file") {
		t.Errorf("it answered %v", err)
	}
	if listed, _ := Do(TaskList, path, Input{}); len(listed.Tasks) != 4 {
		t.Errorf("refused, it still added: %d tasks", len(listed.Tasks))
	}
}

func linkedLine(numbers []int) string {
	parts := make([]string, 0, len(numbers))
	for _, n := range numbers {
		parts = append(parts, "#"+strconv.Itoa(n))
	}
	return strings.Join(parts, ", ")
}

func TestReadExport_refusesAFileOfAnotherKind(t *testing.T) {
	_, err := importInto(t, openEngine(t), `{"cairn": "notes", "version": 4, "tasks": []}`)
	if !errors.Is(err, ErrNotAnExport) {
		t.Errorf("a file of notes answered %v", err)
	}
}

func TestTaskImport_filesEveryLabelPastOneHeldOrOneWithoutColour(t *testing.T) {
	e := openEngine(t)
	if _, err := e.Ask(LabelNew, given("label", "held")); err != nil {
		t.Fatal(err)
	}
	if _, err := importInto(t, e, `{"cairn": "tasks", "version": 4, "tasks": [],
		"labels": [{"label": "held"}, {"label": "idle"}, {"label": "release", "colour": "teal"}]}`); err != nil {
		t.Fatal(err)
	}
	labels, err := e.labelsCarried()
	if err != nil {
		t.Fatal(err)
	}
	var said []string
	for _, l := range labels {
		if l.Label == "release" {
			said = append(said, l.Label+":"+l.Colour)
			continue
		}
		said = append(said, l.Label)
	}
	slices.Sort(said)
	if want := []string{"held", "idle", "release:teal"}; !slices.Equal(said, want) {
		t.Errorf("labels %v, want %v", said, want)
	}
}

func TestTaskImport_countsWhatItAddedAndKeepsWhatEachOneSaid(t *testing.T) {
	e := openEngine(t)
	held, err := importInto(t, e, `{"cairn": "tasks", "version": 4,
		"objectives": [{"number": 1, "title": "Release", "state": "active", "priority": "normal", "labels": ["release"], "estimate": "40h"}],
		"tasks": [{"number": 1, "title": "Ship", "state": "todo", "priority": "normal", "focused": true, "objectives": [1]},
		  {"number": 2, "title": "Watch", "state": "watching", "priority": "normal"}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if held != (domain.Held{Tasks: 2, Objectives: 1}) {
		t.Errorf("added %+v, want 2 tasks and 1 objective", held)
	}
	o, err := e.showObjective(1)
	if err != nil {
		t.Fatal(err)
	}
	if shape := objective.OfObjective(o); !slices.Equal(shape.Labels, []string{"release"}) || shape.Estimate != "40h" {
		t.Errorf("objective labels %v, estimate %q", shape.Labels, shape.Estimate)
	}
	watched, err := e.show(2)
	if err != nil {
		t.Fatal(err)
	}
	if factsOf(watched).State != domain.Watching {
		t.Errorf("task 2 is in %s, want watching", factsOf(watched).State)
	}
	if held, _ := e.focus(); !slices.Equal(held, []int{1}) {
		t.Errorf("in focus %v, want task 1", held)
	}
}

func TestTaskImport_refusesABadObjectiveOrTaskBeforeAddingAnything(t *testing.T) {
	const head = `{"cairn": "tasks", "version": 4, "labels": [{"label": "release"}],`
	for _, c := range []struct{ said, file string }{
		{"objective 1:", head + `"objectives": [{"number": 1, "title": "Release", "state": "nowhere", "priority": "normal"}], "tasks": []}`},
		{"task 1:", head + `"objectives": [], "tasks": [{"number": 1, "title": "Ship", "state": "todo", "priority": "pressing"}]}`},
		{"task 1:", head + `"objectives": [], "tasks": [{"number": 1, "title": "Ship", "state": "todo", "priority": "normal", "spent": {"someday": "1h"}}]}`},
	} {
		e := openEngine(t)
		_, err := importInto(t, e, c.file)
		if err == nil || !strings.HasPrefix(err.Error(), c.said) {
			t.Errorf("importing %s answered %v, want it to start with %q", c.file, err, c.said)
		}
		if labels, _ := e.labelsCarried(); len(labels) != 0 {
			t.Errorf("importing %s filed labels %v before refusing", c.file, labels)
		}
	}
}

func TestTaskExport_writesTimeAddedOnADayPastUnderThatDay(t *testing.T) {
	e := openEngine(t)
	add(t, e, "one")
	day := domain.FormatDay(time.Now().AddDate(0, 0, -10))
	if _, err := e.Ask(TimesheetAdd, given("number", "1", "time", "2h", "on", day)); err != nil {
		t.Fatal(err)
	}
	out, err := e.exporting()
	if err != nil {
		t.Fatal(err)
	}
	if got := out.Tasks[0].Spent; got[day] != "2h" || len(got) != 1 {
		t.Errorf("spent %v, want 2h on %s", got, day)
	}
}

func TestTaskImport_keepsALinkMadeFromTheOtherSide(t *testing.T) {
	e := openEngine(t)
	if _, err := importInto(t, e, `{"cairn": "tasks", "version": 4, "objectives": [],
		"tasks": [{"number": 2, "title": "Two", "state": "backlog", "priority": "normal", "blocked_by": [], "linked": []},
		  {"number": 1, "title": "One", "state": "backlog", "priority": "normal", "linked": [2]}]}`); err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{1, 2} {
		got, err := e.show(n)
		if err != nil {
			t.Fatal(err)
		}
		if linked := factsOf(got).Linked; len(linked) != 1 {
			t.Errorf("task %d is linked to %v, want the other one", n, linked)
		}
	}
}

func TestTaskImport_keepsWhatBlocksATask(t *testing.T) {
	e := openEngine(t)
	if _, err := importInto(t, e, `{"cairn": "tasks", "version": 4, "objectives": [],
		"tasks": [{"number": 1, "title": "One", "state": "todo", "priority": "normal"},
		  {"number": 2, "title": "Two", "state": "blocked", "priority": "normal", "blocked_by": [1]}]}`); err != nil {
		t.Fatal(err)
	}
	got, err := e.show(2)
	if err != nil {
		t.Fatal(err)
	}
	if f := factsOf(got); !slices.Equal(f.Blockers, []int{1}) || f.State != domain.Blocked {
		t.Errorf("task 2 is in %s blocked by %v, want blocked by task 1", f.State, f.Blockers)
	}
}

func exportFile(t *testing.T, said string) string {
	t.Helper()
	from := filepath.Join(t.TempDir(), "held.json")
	if err := os.WriteFile(from, []byte(said), 0o600); err != nil {
		t.Fatal(err)
	}
	return from
}

func TestReadExport_refusesAnotherKindOrAVersionItDoesNotRead(t *testing.T) {
	for _, said := range []string{
		`{"cairn": "notes", "version": 7, "tasks": [{"number": 1, "title": "a"}]}`,
		`{"cairn": "tasks", "version": 8, "tasks": [{"number": 1, "title": "a"}]}`,
		`{"cairn": "tasks", "version": 0, "tasks": [{"number": 1, "title": "a"}]}`,
	} {
		if _, err := readExport(exportFile(t, said)); !errors.Is(err, ErrNotAnExport) {
			t.Errorf("%s was read: %v", said, err)
		}
	}
}

func TestReadExport_takesAFileOfALabelAndASeries(t *testing.T) {
	if _, err := readExport(exportFile(t, `{"cairn": "tasks", "version": 7, "tasks": [], "labels": [{"name": "x"}], "series": [{"number": 1}]}`)); err != nil {
		t.Errorf("a label and a series are something to add: %v", err)
	}
}

func TestImport_aLabelWithAScopeKeepsItsColour(t *testing.T) {
	e := imported(t, `{"cairn": "tasks", "version": 7, "objectives": [], "tasks": [], "labels": [{"label": "area", "colour": "teal", "scope": "objectives"}]}`)
	colours, err := e.store.LabelColours()
	if err != nil || colours["area"] != "teal" {
		t.Errorf("the colours are %v, %v", colours, err)
	}
}
