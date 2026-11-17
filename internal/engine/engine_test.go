package engine

import (
	"cairn/internal/engine/item"
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"cairn/internal/engine/task"
)

func openEngine(t *testing.T) *Engine {
	t.Helper()
	e, err := Open(filepath.Join(t.TempDir(), "cairn.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = e.Close() })
	return e
}

func add(t *testing.T, e *Engine, title string) task.Task {
	t.Helper()
	added, err := e.add(title, Changes{})
	if err != nil {
		t.Fatalf("Add(%q): %v", title, err)
	}
	return added
}

func of[T any](v T) *T { return &v }

func TestEngine_refusesATaskWithNoTitle(t *testing.T) {
	if _, err := openEngine(t).add("  ", Changes{}); !errors.Is(err, item.ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired", err)
	}
}

func TestEngine_refusesAnAttributeItCannotMakeSenseOf(t *testing.T) {
	for name, c := range map[string]Changes{
		"priority": {Priority: of("urgent")},
		"state":    {State: of("someday")},
		"deadline": {Deadline: of("soon")},
		"estimate": {Estimate: of("ages")},
		"label":    {Labels: of([]string{"  "})},
	} {
		if _, err := openEngine(t).add("Write the spec", c); err == nil {
			t.Errorf("Add with a bad %s should fail", name)
		}
	}
}

func TestEngine_clearsAnAttributeGivenEmpty(t *testing.T) {
	e := openEngine(t)
	kept, err := e.add("Write the spec", Changes{
		Deadline: of("2026-09-01"), Priority: of("critical"), State: of("current"),
		Labels: of([]string{"spec", "team"}),
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	added, err := e.edit(kept.Carried().Number, Changes{
		Deadline: of(""), Priority: of(""), State: of(""), Labels: of([]string{""}),
	})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	got := added.Carried()
	if !got.Deadline.IsZero() {
		t.Errorf("Deadline = %v, want it cleared", got.Deadline)
	}
	if got.Priority != item.Normal {
		t.Errorf("Priority = %q, want normal", got.Priority)
	}
	if got.State != item.Backlog {
		t.Errorf("State = %q, want the backlog", got.State)
	}
	if len(got.Labels) != 0 {
		t.Errorf("Labels = %v, want none", got.Labels)
	}
}

func TestEngine_saysWhenNothingWasNamedToChange(t *testing.T) {
	e := openEngine(t)
	kept := add(t, e, "Write the spec")
	if _, err := e.edit(kept.Carried().Number, Changes{}); !errors.Is(err, errNothingToChange) {
		t.Errorf("err = %v, want ErrNothingToChange", err)
	}
}

func TestOpen_reportsAStoreItCannotUse(t *testing.T) {
	if _, err := Open(filepath.Join(t.TempDir(), "no", "such", "\x00", "cairn.db")); err == nil {
		t.Error("Open should report a store it cannot use")
	}
}

func TestEngine_reportsAStoreItCanNoLongerReach(t *testing.T) {
	e, err := Open(filepath.Join(t.TempDir(), "cairn.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	kept, err := e.add("Write the spec", Changes{})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := e.add("Nope", Changes{}); err == nil {
		t.Error("Add on a closed store should fail")
	}
	if _, err := e.tasks.All(); err == nil {
		t.Error("List on a closed store should fail")
	}
	if _, err := e.edit(kept.Carried().Number, Changes{Title: of("Nope")}); err == nil {
		t.Error("Edit on a closed store should fail")
	}
}

func TestEngine_refusesASignedLabelAtCreation(t *testing.T) {
	_, err := openEngine(t).add("Write the spec", Changes{Labels: of([]string{"+urgent"})})
	if !errors.Is(err, errLabelSignAtCreation) {
		t.Errorf("err = %v, want ErrLabelSignAtCreation: a new task has nothing to add to", err)
	}
}

func TestEngine_refusesToReplaceAndChangeLabelsAtOnce(t *testing.T) {
	e := openEngine(t)
	kept := add(t, e, "Write the spec")
	if _, err := e.edit(kept.Carried().Number, Changes{Labels: of([]string{"+urgent", "=draft"})}); err == nil {
		t.Error("Edit should refuse to both replace and change the labels")
	}
}

func TestEngine_reportsWhatItCannotReach(t *testing.T) {
	e, err := Open(filepath.Join(t.TempDir(), "cairn.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	kept, err := e.add("Ship it", Changes{})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := e.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if err := e.delete(kept.Carried().Number); err == nil {
		t.Error("Delete on a closed store should fail")
	}
}

func TestEngine_takesTimeOffAndStopsAtNothing(t *testing.T) {
	e := openEngine(t)
	kept, err := e.add("Write the spec", Changes{Spent: of("2h")})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := e.edit(kept.Carried().Number, Changes{Spent: of("-3h")})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if facts := got.Carried(); facts.Spent != 0 {
		t.Errorf("TimeSpent = %v, want none", facts.Spent)
	}
}

func TestEngine_refusesATimeItCannotMakeSenseOf(t *testing.T) {
	if _, err := openEngine(t).add("Write the spec", Changes{Spent: of("+ages")}); err == nil {
		t.Error("Add with a time it cannot read should fail")
	}
}

func TestEngine_refusesProseATerminalWouldObey(t *testing.T) {
	for name, c := range map[string]Changes{
		"description": {Description: of("look \x1b]52;c;aGk=\x07 here")},
		"because":     {State: of("blocked"), Because: of("waiting\x1b[8m on legal")},
	} {
		if _, err := openEngine(t).add("Write the spec", c); !errors.Is(err, item.ErrNotText) {
			t.Errorf("%s: err = %v, want ErrNotText", name, err)
		}
	}
	if _, err := openEngine(t).addObjective("Ship cairn", ObjectiveChanges{Description: of("bell\x07")}); !errors.Is(err, item.ErrNotText) {
		t.Errorf("objective: err = %v, want ErrNotText", err)
	}
}

func TestDelete_namesWhatItWasBlocking(t *testing.T) {
	path := storeAt(t, "Write the spec", "Ship it", "Unrelated")
	if _, err := Do(TaskEdit, path, Input{Given: map[string]string{"number": "2"}, Listed: map[string][]string{"blocker": {"1"}}}); err != nil {
		t.Fatal(err)
	}
	answer, err := Do(TaskDelete, path, confirmed(given("number", "1")))
	if err != nil {
		t.Fatal(err)
	}
	if len(answer.Unblocked) != 1 || answer.Unblocked[0] != 2 {
		t.Errorf("Unblocked = %v, want [2]", answer.Unblocked)
	}
	if said := Of(TaskDelete, answer).(Deleted); len(said.WasBlocking) != 1 || said.WasBlocking[0] != 2 {
		t.Errorf("Of = %+v, want what it was blocking named", said)
	}
}

func TestDelete_takesSeveralTogetherOrNone(t *testing.T) {
	path := storeAt(t, "First", "Second", "Third")
	if _, err := Do(TaskDelete, path, Input{Confirmed: true, Given: map[string]string{"number": "1"}, Listed: map[string][]string{"also": {"3", "99"}}}); err == nil {
		t.Fatal("a deletion naming a task that does not exist went through")
	}
	if listed, _ := Do(TaskList, path, Input{}); len(listed.Tasks) != 3 {
		t.Fatalf("a refused deletion of several left %d tasks of 3", len(listed.Tasks))
	}
	answer, err := Do(TaskDelete, path, Input{Confirmed: true, Given: map[string]string{"number": "1"}, Listed: map[string][]string{"also": {"3"}}})
	if err != nil {
		t.Fatal(err)
	}
	if listed, _ := Do(TaskList, path, Input{}); len(listed.Tasks) != 1 {
		t.Errorf("a deletion of two left %d tasks of 3", len(listed.Tasks))
	}
	if said := Of(TaskDelete, answer).(Deleted); said.Deleted != 1 || len(said.Also) != 1 || said.Also[0] != 3 {
		t.Errorf("Of = %+v, want 1 and 3 named", said)
	}
}

func TestArchive_takesSeveralTogetherOrNone(t *testing.T) {
	path := storeAt(t, "First", "Second", "Third")
	if _, err := Do(TaskArchive, path, given("number", "3")); err != nil {
		t.Fatal(err)
	}
	if _, err := Do(TaskArchive, path, Input{Confirmed: true, Given: map[string]string{"number": "1"}, Listed: map[string][]string{"also": {"3"}}}); err == nil {
		t.Fatal("archiving what was already archived went through")
	}
	if listed, _ := Do(TaskList, path, Input{}); len(listed.Tasks) != 2 {
		t.Fatalf("a refused archiving of several left %d tasks on the board of 2", len(listed.Tasks))
	}
	answer, err := Do(TaskArchive, path, Input{Confirmed: true, Given: map[string]string{"number": "1"}, Listed: map[string][]string{"also": {"2"}}})
	if err != nil {
		t.Fatal(err)
	}
	if said := Of(TaskArchive, answer).(Swept); said.Archived != 2 {
		t.Errorf("Of = %+v, want two archived", said)
	}
}

func TestObjectiveDelete_takesSeveralTogetherOrNone(t *testing.T) {
	path := storeAt(t)
	for _, title := range []string{"One", "Two"} {
		if _, err := Do(ObjectiveNew, path, given("title", title)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Do(ObjectiveDelete, path, Input{Confirmed: true, Given: map[string]string{"number": "1"}, Listed: map[string][]string{"also": {"9"}}}); err == nil {
		t.Fatal("a deletion naming an objective that does not exist went through")
	}
	if _, err := Do(ObjectiveDelete, path, Input{Confirmed: true, Given: map[string]string{"number": "1"}, Listed: map[string][]string{"also": {"2"}}}); err != nil {
		t.Fatal(err)
	}
	if listed, _ := Do(ObjectiveList, path, Input{}); len(listed.Objectives) != 0 {
		t.Errorf("a deletion of both left %d objectives", len(listed.Objectives))
	}
}

func TestArchive_refusesAlsoWithoutANumber(t *testing.T) {
	path := storeAt(t, "First", "Second")
	_, err := Do(TaskArchive, path, Input{Confirmed: true, Given: map[string]string{"state": "backlog"}, Listed: map[string][]string{"also": {"2"}}})
	if !errors.Is(err, task.ErrAlsoWithoutNumber) {
		t.Errorf("err = %v, want ErrAlsoWithoutNumber", err)
	}
	if _, err := Do(TaskDelete, path, Input{Confirmed: true, Given: map[string]string{"number": "1"}, Listed: map[string][]string{"also": {"two"}}}); err == nil {
		t.Error("a deletion beside something that is not a number went through")
	}
}

func TestEditAndFocus_takeSeveralTogetherOrNone(t *testing.T) {
	path := storeAt(t, "First", "Second", "Third")
	both := func(number string, more ...string) Input {
		return Input{Confirmed: true, Given: map[string]string{"number": number, "priority": "high"}, Listed: map[string][]string{"also": more, "label": {"+urgent"}}}
	}
	if _, err := Do(TaskEdit, path, both("1", "99")); err == nil {
		t.Fatal("an edit naming a task that does not exist went through")
	}
	if shown, _ := Do(TaskShow, path, given("number", "1")); shown.Task.Carried().Priority == "high" {
		t.Fatal("a refused edit of several changed the first")
	}
	if _, err := Do(TaskEdit, path, both("1", "3")); err != nil {
		t.Fatal(err)
	}
	for _, n := range []string{"1", "3"} {
		shown, _ := Do(TaskShow, path, given("number", n))
		if c := shown.Task.Carried(); c.Priority != "high" || !slices.Contains(c.Labels, "urgent") {
			t.Errorf("task %s after an edit of several: %+v", n, shown.Task.Carried())
		}
	}
	focused, err := Do(TaskFocus, path, Input{Given: map[string]string{"number": "1"}, Listed: map[string][]string{"also": {"2"}}})
	if err != nil || len(*focused.InFocus) != 2 {
		t.Fatalf("focusing two holds %v, %v", focused.InFocus, err)
	}
	unfocused, err := Do(TaskUnfocus, path, Input{Given: map[string]string{"number": "1"}, Listed: map[string][]string{"also": {"2"}}})
	if err != nil || len(*unfocused.InFocus) != 0 {
		t.Errorf("unfocusing two leaves %v, %v", unfocused.InFocus, err)
	}
}
