package engine

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/task"
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

func TestEngine_numbersTheFirstTaskOne(t *testing.T) {
	if got := factsOf(add(t, openEngine(t), "Write the spec")).number; got != 1 {
		t.Errorf("Number = %d, want 1", got)
	}
}

func TestEngine_givesTheNextNumberToTheNextTask(t *testing.T) {
	e := openEngine(t)
	add(t, e, "First")
	if got := factsOf(add(t, e, "Second")).number; got != 2 {
		t.Errorf("Number = %d, want 2", got)
	}
}

func TestEngine_refusesATaskWithNoTitle(t *testing.T) {
	if _, err := openEngine(t).add("  ", Changes{}); !errors.Is(err, domain.ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired", err)
	}
}

func TestEngine_carriesTheAttributesItWasGiven(t *testing.T) {
	added, err := openEngine(t).add("Write the spec", Changes{
		Description: of("the one"),
		State:       of("todo"),
		Priority:    of("high"),
		Labels:      of([]string{"Team", "spec"}),
		Deadline:    of("2026-09-01"),
		Estimate:    of("90m"),
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	got := factsOf(added)
	if got.description != "the one" || got.state != domain.Todo || got.priority != domain.High {
		t.Errorf("got %+v, want the attributes it was given", got)
	}
	if len(got.labels) != 2 || got.labels[0] != "spec" || got.labels[1] != "team" {
		t.Errorf("Labels = %v, want [spec team]: lower case, sorted", got.labels)
	}
	if domain.FormatDay(got.deadline) != "2026-09-01" {
		t.Errorf("Deadline = %v", got.deadline)
	}
	if domain.FormatDuration(got.estimate) != "1h30m" {
		t.Errorf("Estimate = %v, want 1h30m", got.estimate)
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

func TestEngine_listsTasksInTheOrderTheyWereAdded(t *testing.T) {
	e := openEngine(t)
	add(t, e, "First")
	add(t, e, "Second")

	got, err := e.tasks.All()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 || factsOf(got[0]).title != "First" || factsOf(got[1]).title != "Second" {
		t.Errorf("List = %v", got)
	}
}

func TestEngine_showsATaskByItsNumber(t *testing.T) {
	e := openEngine(t)
	kept := add(t, e, "Write the spec")

	got, err := e.show(factsOf(kept).number)
	if err != nil {
		t.Fatalf("Show: %v", err)
	}
	if facts := factsOf(got); facts.title != "Write the spec" {
		t.Errorf("Title = %q", facts.title)
	}
}

func TestEngine_saysThereIsNoSuchTask(t *testing.T) {
	if _, err := openEngine(t).show(99); err == nil {
		t.Error("Show(99) should fail")
	}
}

func TestEngine_changesOnlyWhatWasNamed(t *testing.T) {
	e := openEngine(t)
	kept, err := e.add("Write the spec", Changes{Description: of("the one")})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	added, err := e.edit(factsOf(kept).number, Changes{Priority: of("high")})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	got := factsOf(added)
	if got.priority != domain.High {
		t.Errorf("Priority = %q, want high", got.priority)
	}
	if got.description != "the one" {
		t.Errorf("Description = %q, want it left alone", got.description)
	}
	if got.title != "Write the spec" {
		t.Errorf("Title = %q, want it left alone", got.title)
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

	added, err := e.edit(factsOf(kept).number, Changes{
		Deadline: of(""), Priority: of(""), State: of(""), Labels: of([]string{""}),
	})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	got := factsOf(added)
	if !got.deadline.IsZero() {
		t.Errorf("Deadline = %v, want it cleared", got.deadline)
	}
	if got.priority != domain.Normal {
		t.Errorf("Priority = %q, want normal", got.priority)
	}
	if got.state != domain.Backlog {
		t.Errorf("State = %q, want the backlog", got.state)
	}
	if len(got.labels) != 0 {
		t.Errorf("Labels = %v, want none", got.labels)
	}
}

func TestEngine_refusesToEmptyATitle(t *testing.T) {
	e := openEngine(t)
	kept := add(t, e, "Write the spec")
	if _, err := e.edit(factsOf(kept).number, Changes{Title: of("")}); !errors.Is(err, domain.ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired", err)
	}
}

func TestEngine_saysWhenNothingWasNamedToChange(t *testing.T) {
	e := openEngine(t)
	kept := add(t, e, "Write the spec")
	if _, err := e.edit(factsOf(kept).number, Changes{}); !errors.Is(err, ErrNothingToChange) {
		t.Errorf("err = %v, want ErrNothingToChange", err)
	}
}

func TestEngine_doesNotRenumberATaskItChanges(t *testing.T) {
	e := openEngine(t)
	add(t, e, "First")
	second := add(t, e, "Second")

	got, err := e.edit(factsOf(second).number, Changes{Title: of("Second, revised")})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if factsOf(got).number != factsOf(second).number {
		t.Errorf("Number = %d, want %d", factsOf(got).number, factsOf(second).number)
	}
}

func TestEngine_deletesATask(t *testing.T) {
	e := openEngine(t)
	kept := add(t, e, "Write the spec")
	if err := e.delete(factsOf(kept).number); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := e.show(factsOf(kept).number); err == nil {
		t.Error("the task should be gone")
	}
}

func TestEngine_saysThereIsNoSuchTaskToDelete(t *testing.T) {
	if err := openEngine(t).delete(99); err == nil {
		t.Error("Delete(99) should fail")
	}
}

func TestEngine_neverHandsOutANumberFreedByADeletion(t *testing.T) {
	e := openEngine(t)
	add(t, e, "First")
	second := add(t, e, "Second")
	if err := e.delete(factsOf(second).number); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got := factsOf(add(t, e, "Third")).number; got != 3 {
		t.Errorf("Number = %d, want 3", got)
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
	if _, err := e.edit(factsOf(kept).number, Changes{Title: of("Nope")}); err == nil {
		t.Error("Edit on a closed store should fail")
	}
}

func TestEngine_addsALabelToTheOnesAlreadyThere(t *testing.T) {
	e := openEngine(t)
	kept, err := e.add("Write the spec", Changes{Labels: of([]string{"spec", "team"})})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := e.edit(factsOf(kept).number, Changes{Labels: of([]string{"+urgent"})})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if labels := factsOf(got).labels; len(labels) != 3 {
		t.Errorf("Labels = %v, want three", labels)
	}
}

func TestEngine_takesALabelAwayFromTheOnesAlreadyThere(t *testing.T) {
	e := openEngine(t)
	kept, err := e.add("Write the spec", Changes{Labels: of([]string{"spec", "team"})})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := e.edit(factsOf(kept).number, Changes{Labels: of([]string{"-spec"})})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if labels := factsOf(got).labels; len(labels) != 1 || labels[0] != "team" {
		t.Errorf("Labels = %v, want [team]", labels)
	}
}

func TestEngine_refusesASignedLabelAtCreation(t *testing.T) {
	_, err := openEngine(t).add("Write the spec", Changes{Labels: of([]string{"+urgent"})})
	if !errors.Is(err, ErrLabelSignAtCreation) {
		t.Errorf("err = %v, want ErrLabelSignAtCreation: a new task has nothing to add to", err)
	}
}

func TestEngine_refusesToReplaceAndChangeLabelsAtOnce(t *testing.T) {
	e := openEngine(t)
	kept := add(t, e, "Write the spec")
	if _, err := e.edit(factsOf(kept).number, Changes{Labels: of([]string{"+urgent", "draft"})}); err == nil {
		t.Error("Edit should refuse to both replace and change the labels")
	}
}

func TestEngine_hangsATaskUnderAnother(t *testing.T) {
	e := openEngine(t)
	under := add(t, e, "Ship it")

	got, err := e.add("Write the tests", Changes{Parent: of(factsOf(under).number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if facts := factsOf(got); facts.parent != factsOf(under).number {
		t.Errorf("Parent = %d, want %d", facts.parent, factsOf(under).number)
	}
}

func TestEngine_setsATaskFreeWhenTheParentIsCleared(t *testing.T) {
	e := openEngine(t)
	under := add(t, e, "Ship it")
	hung, err := e.add("Write the tests", Changes{Parent: of(factsOf(under).number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := e.edit(factsOf(hung).number, Changes{Parent: of(task.NoParent)})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if facts := factsOf(got); facts.parent != task.NoParent {
		t.Errorf("Parent = %d, want none", facts.parent)
	}
}

func TestEngine_refusesAParentThatDoesNotExist(t *testing.T) {
	if _, err := openEngine(t).add("Write the tests", Changes{Parent: of(99)}); !errors.Is(err, ErrNoSuchTask) {
		t.Errorf("err = %v, want ErrNoSuchTask", err)
	}
}

func TestEngine_refusesATaskHangingUnderItself(t *testing.T) {
	e := openEngine(t)
	alone := add(t, e, "Ship it")
	if _, err := e.edit(factsOf(alone).number, Changes{Parent: of(factsOf(alone).number)}); !errors.Is(err, ErrParentIsSelf) {
		t.Errorf("err = %v, want ErrParentIsSelf", err)
	}
}

func TestEngine_refusesToHangATaskUnderASubTask(t *testing.T) {
	e := openEngine(t)
	under := add(t, e, "Ship it")
	hung, err := e.add("Write the tests", Changes{Parent: of(factsOf(under).number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	if _, err := e.add("Write the first one", Changes{Parent: of(factsOf(hung).number)}); !errors.Is(err, ErrParentIsSubTask) {
		t.Errorf("err = %v, want ErrParentIsSubTask", err)
	}
}

func TestEngine_deletesATaskOnceNothingHangsUnderIt(t *testing.T) {
	e := openEngine(t)
	under := add(t, e, "Ship it")
	hung, err := e.add("Write the tests", Changes{Parent: of(factsOf(under).number)})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := e.delete(factsOf(hung).number); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if err := e.delete(factsOf(under).number); err != nil {
		t.Errorf("Delete: %v, want it to go now that nothing hangs under it", err)
	}
}

func TestEngine_namesWhatHangsUnderATask(t *testing.T) {
	e := openEngine(t)
	under := add(t, e, "Ship it")
	if _, err := e.add("Write the tests", Changes{Parent: of(factsOf(under).number)}); err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := e.under(factsOf(under).number)
	if err != nil {
		t.Fatalf("Under: %v", err)
	}
	if len(got) != 1 || factsOf(got[0]).title != "Write the tests" {
		t.Errorf("Under = %v, want the one sub-task", got)
	}
}

func TestEngine_reportsSubTasksItCannotReach(t *testing.T) {
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

	if _, err := e.under(factsOf(kept).number); err == nil {
		t.Error("Under on a closed store should fail")
	}
	if err := e.delete(factsOf(kept).number); err == nil {
		t.Error("Delete on a closed store should fail")
	}
}

func TestEngine_writesDownTheTimeSpent(t *testing.T) {
	e := openEngine(t)
	kept := add(t, e, "Write the spec")

	got, err := e.edit(factsOf(kept).number, Changes{Spent: of("2h")})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if domain.FormatDuration(factsOf(got).timeSpent) != "2h" {
		t.Errorf("TimeSpent = %v, want 2h", factsOf(got).timeSpent)
	}
}

func TestEngine_addsToTheTimeAlreadySpent(t *testing.T) {
	e := openEngine(t)
	kept, err := e.add("Write the spec", Changes{Spent: of("2h")})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := e.edit(factsOf(kept).number, Changes{Spent: of("+30m")})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if domain.FormatDuration(factsOf(got).timeSpent) != "2h30m" {
		t.Errorf("TimeSpent = %v, want 2h30m", factsOf(got).timeSpent)
	}
}

func TestEngine_takesTimeOffAndStopsAtNothing(t *testing.T) {
	e := openEngine(t)
	kept, err := e.add("Write the spec", Changes{Spent: of("2h")})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	got, err := e.edit(factsOf(kept).number, Changes{Spent: of("-3h")})
	if err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if facts := factsOf(got); facts.timeSpent != 0 {
		t.Errorf("TimeSpent = %v, want none", facts.timeSpent)
	}
}

func TestEngine_refusesATimeItCannotMakeSenseOf(t *testing.T) {
	if _, err := openEngine(t).add("Write the spec", Changes{Spent: of("+ages")}); err == nil {
		t.Error("Add with a time it cannot read should fail")
	}
}

// A task goes off with what hangs under it. It used to be refused while a part
// was there, and the parts are reachable only in the form of the task they
// belong to — so a task that carried one could not be deleted at all once the
// board stopped drawing them.
//
// It is the rule archiving already goes by: a part goes off the board with the
// work it belongs to. A part is not work in its own right; it is a line of a
// list its task owns.
func TestDeletingATaskTakesItsPartsWithIt(t *testing.T) {
	e := openEngine(t)
	ship := factsOf(add(t, e, "Ship it")).number
	parent := ship
	first := factsOf(add(t, e, "Relire l'ADR")).number
	second := factsOf(add(t, e, "Trancher")).number
	for _, part := range []int{first, second} {
		if _, err := e.edit(part, Changes{Parent: &parent}); err != nil {
			t.Fatal(err)
		}
	}
	if err := e.delete(ship); err != nil {
		t.Fatalf("deleting a task carrying parts: %v", err)
	}
	for _, gone := range []int{ship, first, second} {
		if _, err := e.show(gone); err == nil {
			t.Errorf("task %d is still there", gone)
		}
	}
}

// What is written about a task reaches a terminal as it was written, from
// more than one surface, so an escape among it is refused before it is kept.
func TestEngine_refusesProseATerminalWouldObey(t *testing.T) {
	for name, c := range map[string]Changes{
		"description": {Description: of("look \x1b]52;c;aGk=\x07 here")},
		"because":     {State: of("blocked"), Because: of("waiting\x1b[8m on legal")},
	} {
		if _, err := openEngine(t).add("Write the spec", c); !errors.Is(err, domain.ErrNotText) {
			t.Errorf("%s: err = %v, want ErrNotText", name, err)
		}
	}
	if _, err := openEngine(t).addObjective("Ship cairn", ObjectiveChanges{Description: of("bell\x07")}); !errors.Is(err, domain.ErrNotText) {
		t.Errorf("objective: err = %v, want ErrNotText", err)
	}
}

// A deletion names the tasks it was in the way of: the link goes with it,
// and afterwards nothing says a task once waited on it.
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
	if got := RenderText(TaskDelete, answer); !strings.Contains(got, "task 2 was waiting on it") {
		t.Errorf("RenderText = %q", got)
	}
	if said := Of(TaskDelete, answer).(Deleted); len(said.WasBlocking) != 1 || said.WasBlocking[0] != 2 {
		t.Errorf("Of = %+v, want what it was blocking named", said)
	}
}

// The version a board watches moves when somebody else writes, and only then.
func TestVersion_movesWhenAnotherWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	watching, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = watching.Close() })
	before, err := watching.Version()
	if err != nil {
		t.Fatal(err)
	}
	if again, _ := watching.Version(); again != before {
		t.Errorf("the version moved from %d to %d with nobody writing", before, again)
	}
	if _, err := Do(TaskNew, path, given("title", "from elsewhere")); err != nil {
		t.Fatal(err)
	}
	if after, _ := watching.Version(); after == before {
		t.Error("the version stayed put though another wrote to the store")
	}
}

// Several tasks are deleted in one act, all of them or none: a deletion of
// three that met a number with no task behind it took the first two anyway,
// and left somebody to work out which had gone.
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

// Several tasks are archived in one act, all of them or none.
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

// Several objectives are deleted in one act, all of them or none.
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

// Archiving names more tasks beside a number, never instead of one.
func TestArchive_refusesAlsoWithoutANumber(t *testing.T) {
	path := storeAt(t, "First", "Second")
	_, err := Do(TaskArchive, path, Input{Confirmed: true, Given: map[string]string{"state": "backlog"}, Listed: map[string][]string{"also": {"2"}}})
	if !errors.Is(err, ErrAlsoWithoutNumber) {
		t.Errorf("err = %v, want ErrAlsoWithoutNumber", err)
	}
	if _, err := Do(TaskDelete, path, Input{Confirmed: true, Given: map[string]string{"number": "1"}, Listed: map[string][]string{"also": {"two"}}}); err == nil {
		t.Error("a deletion beside something that is not a number went through")
	}
}

// A number beside the first is a task's: nought and less are refused, the way
// a word is.
func TestDelete_refusesNoughtBesideTheFirst(t *testing.T) {
	path := storeAt(t, "First", "Second")
	for _, beside := range []string{"0", "-2"} {
		if _, err := Do(TaskDelete, path, Input{Confirmed: true, Given: map[string]string{"number": "1"}, Listed: map[string][]string{"also": {beside}}}); err == nil {
			t.Errorf("a deletion beside %q went through", beside)
		}
	}
}
