package engine

import (
	"slices"
	"testing"

	"cairn/internal/task"
)

func titlesOf(a Answer) []string {
	titles := []string{}
	for _, one := range append(append([]task.Task{}, a.Tasks...), a.Matched...) {
		titles = append(titles, task.OfTask(one).Title)
	}
	return titles
}

func TestTaskList_keepsWhatTheFactsNamedHoldFor(t *testing.T) {
	e := openEngine(t)
	spec := add(t, e, "Write the spec")
	add(t, e, "Ship it")
	if _, err := e.edit(task.OfTask(spec).Number, Changes{State: of("todo")}); err != nil {
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

func TestTaskList_saysNothingAboutAFocusItNarrowedOut(t *testing.T) {
	e := openEngine(t)
	spec := add(t, e, "Write the spec")
	shipIt := add(t, e, "Ship it")
	if _, err := e.edit(task.OfTask(spec).Number, Changes{State: of("todo")}); err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if _, err := e.putInFocus(task.OfTask(shipIt).Number); err != nil {
		t.Fatalf("PutInFocus: %v", err)
	}

	narrowed, err := TaskList.Run(On{Engine: e}, Input{Given: map[string]string{"state": "todo"}})
	if err != nil {
		t.Fatalf("TaskList: %v", err)
	}
	if narrowed.InFocus == nil || len(*narrowed.InFocus) != 0 {
		t.Errorf("it says %v is in focus, and did not hand that task over", narrowed.InFocus)
	}

	kept, err := TaskList.Run(On{Engine: e}, Input{Given: map[string]string{"state": "backlog"}})
	if err != nil {
		t.Fatalf("TaskList: %v", err)
	}
	if kept.InFocus == nil || !slices.Contains(*kept.InFocus, task.OfTask(shipIt).Number) {
		t.Errorf("in focus = %v, want the task it kept", kept.InFocus)
	}
}

func TestTaskList_refusesOnlyWhatAClosedSetExcludes(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")

	if _, err := TaskList.Run(On{Engine: e}, Input{
		Given: map[string]string{"state": "nowhere"},
	}); err == nil {
		t.Error("a column no task can sit in was accepted")
	}

	answer, err := TaskList.Run(On{Engine: e}, Input{
		Listed: map[string][]string{"objective": {"9"}},
	})
	if err != nil {
		t.Fatalf("an objective nobody added was refused: %v", err)
	}
	if got := titlesOf(answer); len(got) != 0 {
		t.Errorf("it listed %v, want nothing serving objective 9", got)
	}
}

func TestObjectiveList_isHeldToTheObjectiveStates(t *testing.T) {
	e := openEngine(t)
	if _, err := ObjectiveList.Run(On{Engine: e}, Input{
		Given: map[string]string{"state": "todo"},
	}); err == nil {
		t.Error("an objective listing accepted a column a task sits in")
	}
}

func TestTaskList_refusesAnObjectiveNothingCouldBeNumbered(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")
	for _, named := range []string{"soon", "0", "-1"} {
		if _, err := TaskList.Run(On{Engine: e}, Input{
			Listed: map[string][]string{"objective": {named}},
		}); err == nil {
			t.Errorf("%q was accepted as an objective to narrow by", named)
		}
	}
}

func TestTaskList_looksForALabelWrittenTheOneWay(t *testing.T) {
	e := openEngine(t)
	spec := add(t, e, "Write the spec")
	add(t, e, "Ship it")
	if _, err := e.edit(task.OfTask(spec).Number, Changes{Labels: of([]string{"sp26-20"})}); err != nil {
		t.Fatalf("Edit: %v", err)
	}

	answer, err := TaskList.Run(On{Engine: e}, Input{Listed: map[string][]string{"label": {"SP26_20"}}})
	if err != nil {
		t.Fatalf("TaskList: %v", err)
	}
	if got := titlesOf(answer); len(got) != 1 || got[0] != "Write the spec" {
		t.Errorf("it listed %v, want the one task labelled sp26-20", got)
	}
}
