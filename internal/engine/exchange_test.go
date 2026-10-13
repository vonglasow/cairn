package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
			if pairs[i] == "label" || pairs[i] == "objective" || pairs[i] == "blocker" {
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
		asked(TaskNew, "title", "Write the spec", "objective", "1"),
		asked(TaskNew, "title", "Old idea"),
		asked(TaskArchive, "number", "2"),
	)
	to := filepath.Join(t.TempDir(), "elsewhere", "tasks.json")
	answer, err := Do(TaskExport, path, Input{Given: map[string]string{"to": to}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if got := RenderText(TaskExport, answer); got != to+" holds 2 task(s) and 1 objective(s)\n" {
		t.Errorf("it says %q", got)
	}
	if shape, ok := Of(TaskExport, answer).(Exported); !ok || shape.Exported != to || shape.Holds.Tasks != 2 {
		t.Errorf("a machine reads %#v", Of(TaskExport, answer))
	}
	written, err := os.ReadFile(to)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"cairn": "tasks"`, `"version": 1`, `"Write the spec"`, `"Old idea"`, `"archived": true`, `"Release 1.0"`} {
		if !strings.Contains(string(written), want) {
			t.Errorf("the file holds no %s:\n%s", want, written)
		}
	}
	if strings.Contains(string(written), "Served by nothing") {
		t.Error("an objective no task serves was written out")
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
		asked(TaskNew, "title", "Write the spec", "parent", "1", "state", "done"),
		asked(TaskNew, "title", "Review", "state", "todo"),
		asked(TaskEdit, "number", "1", "blocker", "3"),
		asked(TaskNew, "title", "Old idea"),
		asked(TaskArchive, "number", "4"),
	)
	file := filepath.Join(t.TempDir(), "tasks.json")
	ran(t, from, asked(TaskExport, "to", file))

	path := storeAt(t)
	ran(t, path, asked(TaskNew, "title", "Already here"), asked(ObjectiveNew, "title", "Already served"))
	answer, err := Do(TaskImport, path, Input{Given: map[string]string{"from": file}})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if got := RenderText(TaskImport, answer); got != "4 task(s) and 1 objective(s) were added from "+file+"\n" {
		t.Errorf("it says %q", got)
	}
	if shape, ok := Of(TaskImport, answer).(Imported); !ok || shape.Imported != file || shape.Added.Tasks != 4 || shape.Added.Objectives != 1 {
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
	for _, want := range []string{"Ship it", "high", "release", "2026-12-01", "2h", "All of it", "blocked by   #4", "objectives   #2"} {
		if !strings.Contains(shipped, want) {
			t.Errorf("task 2 does not say %q:\n%s", want, shipped)
		}
	}
	if part := show("3"); !strings.Contains(part, "parent       2") || !strings.Contains(part, "done") {
		t.Errorf("the part reads:\n%s", part)
	}
	if old := show("5"); !strings.Contains(old, "archived") {
		t.Errorf("the archived task reads:\n%s", old)
	}
	if here := show("1"); !strings.Contains(here, "Already here") {
		t.Errorf("task 1 became:\n%s", here)
	}
	ran(t, path, asked(TaskImport, "from", file))
	if listed, _ := Do(TaskList, path, Input{Given: map[string]string{"with-archived": "true"}}); len(listed.Tasks) < 9 {
		t.Errorf("read twice, the store lists %d tasks", len(listed.Tasks))
	}
}

func TestTaskImport_refusesBeforeAddingAnything(t *testing.T) {
	path := storeAt(t)
	ran(t, path, asked(TaskNew, "title", "Already here"))
	dir := t.TempDir()
	files := map[string]string{
		"not json":       "not tasks\n",
		"not an export":  `{"something": "else"}`,
		"a later format": `{"cairn": "tasks", "version": 2, "tasks": []}`,
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
