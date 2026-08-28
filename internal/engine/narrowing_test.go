package engine

// What a narrowed listing answers. The narrowing itself is the store's, and
// tested there; what is tested here is what an operation makes of what it was
// given — which facts it passes on, and what it says about the focus once part
// of the store has been left out.

import (
	"testing"

	"cairn/internal/task"
)

// titlesOf reads back the titles a listing answered with, whichever field it
// used to carry them.
func titlesOf(a Answer) []string {
	titles := []string{}
	for _, one := range append(append([]task.Task{}, a.Tasks...), a.Matched...) {
		titles = append(titles, task.OfTask(one).Title)
	}
	return titles
}

// A fact named on a listing keeps only what carries it. The one case the store
// cannot be asked about: whether the operation passes on what it was given.
func TestTaskList_keepsWhatTheFactsNamedHoldFor(t *testing.T) {
	e := openEngine(t)
	spec := add(t, e, "Write the spec")
	add(t, e, "Ship it")
	if _, err := e.Edit(task.OfTask(spec).Number, Changes{State: of("todo")}); err != nil {
		t.Fatalf("Edit: %v", err)
	}

	answer, err := TaskList.Run(On{Engine: e}, Input{Given: map[string]string{"state": "todo"}})
	if err != nil {
		t.Fatalf("TaskList: %v", err)
	}
	if got := titlesOf(answer); len(got) != 1 || got[0] != "Write the spec" {
		t.Errorf("it listed %v, want the one task in todo", got)
	}
}

// A narrowed listing is answered flat, an unnarrowed one as the tree it is.
// Which field carries the tasks is how Of and RenderText tell the two apart,
// so it is asserted rather than assumed.
func TestTaskList_answersNarrowedFlatAndUnnarrowedAsATree(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")

	whole, err := TaskList.Run(On{Engine: e}, Input{Given: map[string]string{}})
	if err != nil {
		t.Fatalf("TaskList: %v", err)
	}
	if whole.Tasks == nil || whole.Matched != nil {
		t.Error("a listing nobody narrowed is not answered as a tree")
	}

	narrowed, err := TaskList.Run(On{Engine: e}, Input{Given: map[string]string{"state": "backlog"}})
	if err != nil {
		t.Fatalf("TaskList: %v", err)
	}
	if narrowed.Matched == nil || narrowed.Tasks != nil {
		t.Error("a narrowed listing is not answered flat")
	}
}

// The focus is a flag on a task, so a listing the task was narrowed out of has
// nothing to say about it. Said anyway, an agent would read a number it cannot
// find among the tasks it was handed.
func TestTaskList_saysNothingAboutAFocusItNarrowedOut(t *testing.T) {
	e := openEngine(t)
	spec := add(t, e, "Write the spec")
	shipIt := add(t, e, "Ship it")
	if _, err := e.Edit(task.OfTask(spec).Number, Changes{State: of("todo")}); err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if _, err := e.PutInFocus(task.OfTask(shipIt).Number); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}

	narrowed, err := TaskList.Run(On{Engine: e}, Input{Given: map[string]string{"state": "todo"}})
	if err != nil {
		t.Fatalf("TaskList: %v", err)
	}
	if narrowed.InFocus != nil {
		t.Errorf("it says %d is in focus, and did not hand that task over", *narrowed.InFocus)
	}

	kept, err := TaskList.Run(On{Engine: e}, Input{Given: map[string]string{"state": "backlog"}})
	if err != nil {
		t.Fatalf("TaskList: %v", err)
	}
	if kept.InFocus == nil || *kept.InFocus != task.OfTask(shipIt).Number {
		t.Errorf("in focus = %v, want the task it kept", kept.InFocus)
	}
}

// A value outside a closed set is refused, and one from an argument that has no
// closed set is passed on untouched. Both halves matter: refusing everything
// would turn `--objective 9` into a failure, and refusing nothing would turn
// `--state nowhere` into an empty column.
func TestTaskList_refusesOnlyWhatAClosedSetExcludes(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")

	if _, err := TaskList.Run(On{Engine: e}, Input{
		Given: map[string]string{"state": "nowhere"},
	}); err == nil {
		t.Error("a column no task can sit in was accepted")
	}

	// An objective nothing serves has no closed set to fall outside of, so it
	// is answered with nothing rather than refused: a filter says what is
	// wanted, and whether it exists is a question about objectives.
	answer, err := TaskList.Run(On{Engine: e}, Input{
		Given: map[string]string{"objective": "9"},
	})
	if err != nil {
		t.Fatalf("an objective nobody added was refused: %v", err)
	}
	if got := titlesOf(answer); len(got) != 0 {
		t.Errorf("it listed %v, want nothing serving objective 9", got)
	}
}

// A listing of objectives is held to its own closed set, which is not the
// tasks': an objective stands somewhere rather than sitting in a column, and a
// column name is not one of the places it can stand.
func TestObjectiveList_isHeldToTheObjectiveStates(t *testing.T) {
	e := openEngine(t)
	if _, err := ObjectiveList.Run(On{Engine: e}, Input{
		Given: map[string]string{"state": "todo"},
	}); err == nil {
		t.Error("an objective listing accepted a column a task sits in")
	}
}
