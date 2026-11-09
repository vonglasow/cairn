package engine

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"cairn/internal/objective"
	"cairn/internal/task"
)

func under(t *testing.T, e *Engine, title string, parent string) objective.Objective {
	t.Helper()
	o, err := e.addObjective(title, ObjectiveChanges{Parent: of(parent)})
	if err != nil {
		t.Fatalf("addObjective(%q under %s): %v", title, parent, err)
	}
	return o
}

func servingAll(t *testing.T, e *Engine, title string, numbers ...int) task.Task {
	t.Helper()
	added, err := e.add(title, Changes{Objective: serves(numbers...)})
	if err != nil {
		t.Fatalf("add(%q): %v", title, err)
	}
	return added
}

func numbersOfTasks(tasks []task.Task) []int {
	out := []int{}
	for _, one := range tasks {
		out = append(out, factsOf(one).Number)
	}
	return out
}

func TestEngine_putsAnObjectiveUnderAnother(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Home automation")
	if got := under(t, e, "Enclosure", "1").Carried().Parent; got != 1 {
		t.Errorf("Parent = %d, want 1", got)
	}
	found, _ := e.showObjective(2)
	if found.Carried().Parent != 1 {
		t.Errorf("the parent was not kept: %+v", found.Carried())
	}
}

func TestEngine_refusesAParentThatIsNotThere(t *testing.T) {
	e := openEngine(t)
	if _, err := e.addObjective("Enclosure", ObjectiveChanges{Parent: of("9")}); !errors.Is(err, ErrNoSuchObjective) {
		t.Errorf("err = %v, want ErrNoSuchObjective", err)
	}
}

func TestEngine_refusesAParentThatIsNoNumber(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Home automation")
	for _, said := range []string{"home", "0", "-1"} {
		if _, err := e.editObjective(1, ObjectiveChanges{Parent: of(said)}); !errors.Is(err, ErrObjectiveRequired) {
			t.Errorf("parent %q: err = %v, want ErrObjectiveRequired", said, err)
		}
	}
}

func TestEngine_refusesAThirdLevelAndAParentUnderAnother(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Home automation")
	under(t, e, "Firmware", "1")
	addObjective(t, e, "Garden")
	if _, err := e.addObjective("Bootloader", ObjectiveChanges{Parent: of("2")}); !errors.Is(err, objective.ErrTooDeep) {
		t.Errorf("a third level: err = %v, want ErrTooDeep", err)
	}
	if _, err := e.editObjective(1, ObjectiveChanges{Parent: of("3")}); !errors.Is(err, objective.ErrTooDeep) {
		t.Errorf("a parent put under another: err = %v, want ErrTooDeep", err)
	}
	if _, err := e.editObjective(3, ObjectiveChanges{Parent: of("3")}); !errors.Is(err, objective.ErrUnderItself) {
		t.Errorf("under itself: err = %v, want ErrUnderItself", err)
	}
}

func TestEngine_setsASubObjectiveLooseWithNone(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Home automation")
	under(t, e, "Enclosure", "1")
	for _, said := range []string{"none", ""} {
		edited, err := e.editObjective(2, ObjectiveChanges{Parent: of(said)})
		if err != nil || edited.Carried().Parent != 0 {
			t.Errorf("parent %q: %+v, %v, want it loose", said, edited.Carried(), err)
		}
	}
}

func TestEngine_refusesAMoveThatCountsATaskTwice(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Home automation")
	addObjective(t, e, "Garden")
	under(t, e, "Valves", "1")
	servingAll(t, e, "Buy the valves", 2, 3)
	_, err := e.editObjective(3, ObjectiveChanges{Parent: of("2")})
	if !errors.Is(err, objective.ErrCountedTwice) || !strings.Contains(err.Error(), "#1 is already counted in #2 Garden through #3 Valves") {
		t.Errorf("err = %v, want the move refused", err)
	}
	if found, _ := e.showObjective(3); found.Carried().Parent != 1 {
		t.Errorf("the refused move changed the parent to %d", found.Carried().Parent)
	}
}

func TestEngine_refusesATaskCountedTwiceInOneTree(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Home automation")
	under(t, e, "Enclosure", "1")
	under(t, e, "Firmware", "1")
	servingAll(t, e, "Mount the board", 2, 3)
	if _, err := e.edit(1, Changes{Objective: of([]string{"+1"})}); !errors.Is(err, objective.ErrCountedTwice) {
		t.Errorf("edit: err = %v, want ErrCountedTwice", err)
	}
	if _, err := e.add("Wiring", Changes{Objective: serves(1, 2)}); !errors.Is(err, objective.ErrCountedTwice) || !strings.HasPrefix(err.Error(), "the task is already counted in #1") {
		t.Errorf("add: err = %v, want ErrCountedTwice naming the task", err)
	}
}

func TestEngine_countsAParentOverItsTreeEachTaskOnce(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Home automation")
	under(t, e, "Enclosure", "1")
	under(t, e, "Firmware", "1")
	servingAll(t, e, "Mount the board", 2, 3)
	servingAll(t, e, "Wiring", 1)
	servingAll(t, e, "Flash it", 3)
	if _, err := e.edit(1, Changes{State: of("done")}); err != nil {
		t.Fatal(err)
	}
	p, err := e.objectiveProgress(1)
	if err != nil || p.Tasks.Total != 3 || p.Tasks.Done != 1 {
		t.Errorf("progress of the parent = %+v, %v, want 1 of 3", p, err)
	}
	got, err := e.tasksUnder(1)
	if err != nil || !slices.Equal(slices.Sorted(slices.Values(numbersOfTasks(got))), []int{1, 2, 3}) {
		t.Errorf("tasks under the parent = %v, %v", numbersOfTasks(got), err)
	}
	if p, _ := e.objectiveProgress(3); p.Tasks.Total != 2 {
		t.Errorf("a sub-objective counts %+v, want its own 2", p)
	}
}

func TestEngine_familyOfAParentNamesItsSubObjectivesAndProposesItReached(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Home automation")
	under(t, e, "Enclosure", "1")
	under(t, e, "Firmware", "1")
	servingAll(t, e, "Print the case", 2)
	f, err := e.familyOf(1)
	if err != nil || len(f.Children) != 2 || f.Children[0].Progress.Tasks.Total != 1 || f.Proposed {
		t.Fatalf("family = %+v, %v", f, err)
	}
	for _, n := range []string{"2", "3"} {
		if _, err := e.Ask(ObjectiveEdit, given("number", n, "state", "reached")); err != nil {
			t.Fatal(err)
		}
	}
	if f, _ := e.familyOf(1); !f.Proposed {
		t.Error("a parent whose every sub-objective is reached is not proposed")
	}
	if f, _ := e.familyOf(2); f.Parent.Carried().Number != 1 || f.Proposed {
		t.Errorf("the family of a sub-objective = %+v", f)
	}
	if shown, _ := e.Ask(ObjectiveShow, given("number", "1")); shown.Objective.Carried().State != objective.Active {
		t.Error("a parent proposed reached was put there by itself")
	}
}

func TestObjectiveEdit_proposesTheParentWhenTheLastSubObjectiveIsReached(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Home automation")
	under(t, e, "Enclosure", "1")
	under(t, e, "Firmware", "1")
	first, err := e.Ask(ObjectiveEdit, given("number", "2", "state", "reached"))
	if err != nil || first.Proposes != nil {
		t.Fatalf("the first reached proposes %v, %v", first.Proposes, err)
	}
	last, err := e.Ask(ObjectiveEdit, given("number", "3", "state", "reached"))
	if err != nil || last.Proposes == nil || last.Proposes.Carried().Number != 1 {
		t.Fatalf("the last reached proposes %v, %v", last.Proposes, err)
	}
	if again, _ := e.Ask(ObjectiveEdit, given("number", "3", "title", "Firmware, flashed")); again.Proposes != nil {
		t.Error("an act that reached nothing proposes the parent")
	}
	if !strings.Contains(RenderText(ObjectiveEdit, last), "every sub-objective of #1 Home automation is reached") {
		t.Errorf("the answer does not propose it:\n%s", RenderText(ObjectiveEdit, last))
	}
}

func TestTaskList_reachesTheSubObjectivesUnlessItselfIsAsked(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Home automation")
	under(t, e, "Enclosure", "1")
	servingAll(t, e, "Print the case", 2)
	servingAll(t, e, "Wiring", 1)
	add(t, e, "Elsewhere")
	all, err := e.Ask(TaskList, Input{Listed: map[string][]string{"objective": {"1"}}})
	if err != nil || !slices.Equal(numbersOfTasks(all.Matched), []int{1, 2}) {
		t.Errorf("--objective 1 = %v, %v", numbersOfTasks(all.Matched), err)
	}
	itself, err := e.Ask(TaskList, Input{Given: map[string]string{"itself": "true"}, Listed: map[string][]string{"objective": {"1"}}})
	if err != nil || !slices.Equal(numbersOfTasks(itself.Matched), []int{2}) {
		t.Errorf("--objective 1 --itself = %v, %v", numbersOfTasks(itself.Matched), err)
	}
}

func deleting(t *testing.T, e *Engine, given map[string]string, listed map[string][]string) (Answer, error) {
	t.Helper()
	return e.Ask(ObjectiveDelete, Agreed(Input{Given: given, Listed: listed}))
}

func TestObjectiveDelete_asksWhatGoesWithASubObjectiveOrATask(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Home automation")
	under(t, e, "Enclosure", "1")
	servingAll(t, e, "Print the case", 2)
	servingAll(t, e, "Wiring", 1)
	_, err := deleting(t, e, map[string]string{"number": "1"}, nil)
	if !errors.Is(err, ErrObjectiveCarriesTasks) || !errors.Is(err, ErrObjectiveHoldsSubObjectives) {
		t.Fatalf("err = %v, want it to carry tasks and hold sub-objectives", err)
	}
	for _, said := range []string{"#2  Enclosure", "#1  Print the case", "#2  Wiring", "--sub-objectives keep|delete", "--tasks detach|delete"} {
		if !strings.Contains(err.Error(), said) {
			t.Errorf("the refusal does not say %q:\n%s", said, err)
		}
	}
	_, err = deleting(t, e, map[string]string{"number": "1", "sub-objectives": "keep"}, nil)
	if !errors.Is(err, ErrObjectiveCarriesTasks) || errors.Is(err, ErrObjectiveHoldsSubObjectives) || strings.Contains(err.Error(), "--sub-objectives") {
		t.Errorf("with the sub-objectives seen to: err = %v, want only the tasks asked about", err)
	}
	if all, _ := e.listObjectives(); len(all) != 2 {
		t.Errorf("a refused deletion left %d objectives", len(all))
	}
}

func TestObjectiveDelete_keepsTheSubObjectivesAndDetachesTheTasks(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Home automation")
	under(t, e, "Enclosure", "1")
	addObjective(t, e, "Garden")
	servingAll(t, e, "Wiring", 1, 3)
	servingAll(t, e, "Print the case", 2)
	answer, err := deleting(t, e, map[string]string{"number": "1", "sub-objectives": "keep", "tasks": "detach"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(answer.Went.Kept, []int{2}) || !slices.Equal(answer.Went.Detached, []int{1}) {
		t.Errorf("went = %+v", answer.Went)
	}
	kept, _ := e.showObjective(2)
	wiring, _ := e.show(1)
	if kept.Carried().Parent != 0 || !slices.Equal(factsOf(wiring).Objectives, []int{3}) {
		t.Errorf("kept %+v, the task serves %v", kept.Carried(), factsOf(wiring).Objectives)
	}
	if !strings.Contains(RenderText(ObjectiveDelete, answer), "task 1 was detached from it") {
		t.Errorf("the answer reads:\n%s", RenderText(ObjectiveDelete, answer))
	}
}

func TestObjectiveDelete_takesTheSubObjectivesAndTheirTasks(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Home automation")
	under(t, e, "Enclosure", "1")
	under(t, e, "Firmware", "1")
	servingAll(t, e, "Print the case", 2)
	servingAll(t, e, "Flash it", 3)
	answer, err := deleting(t, e, map[string]string{"number": "1", "sub-objectives": "delete", "tasks": "delete"},
		map[string][]string{"keep": {"3"}})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(answer.Went.Gone, []int{2}) || !slices.Equal(answer.Went.Kept, []int{3}) || !slices.Equal(answer.Went.Deleted, []int{1}) {
		t.Errorf("went = %+v", answer.Went)
	}
	all, _ := e.listObjectives()
	left, _ := e.Ask(TaskList, Input{})
	if len(all) != 1 || !slices.Equal(numbersOfTasks(left.Tasks), []int{2}) {
		t.Errorf("left %d objectives and tasks %v", len(all), numbersOfTasks(left.Tasks))
	}
}

func TestObjectiveDelete_refusesANumberThatIsNoSubObjectiveOfIt(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Home automation")
	under(t, e, "Enclosure", "1")
	addObjective(t, e, "Garden")
	_, err := deleting(t, e, map[string]string{"number": "1", "sub-objectives": "delete"}, map[string][]string{"keep": {"3"}})
	if !errors.Is(err, ErrNotASubObjective) || err.Error() != "#3 is not a sub-objective of #1; only a sub-objective of what is deleted is kept or deleted on its own" {
		t.Errorf("err = %v, want ErrNotASubObjective", err)
	}
	if _, err := deleting(t, e, map[string]string{"number": "1"}, map[string][]string{"keep": {"2"}, "delete": {"2"}}); !errors.Is(err, ErrNotASubObjective) {
		t.Errorf("kept and deleted at once: err = %v, want it refused", err)
	}
	if _, err := deleting(t, e, map[string]string{"number": "1", "sub-objectives": "maybe"}, nil); !errors.Is(err, ErrNotOneOf) {
		t.Errorf("a choice it does not know: err = %v", err)
	}
}

func TestObjectiveReport_splitsEachMonthBySubObjective(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Probation")
	under(t, e, "Release", "1")
	servingAll(t, e, "Tag it", 2)
	servingAll(t, e, "Write it up", 1)
	for _, spent := range [][]string{{"1", "2h", "2026-09-10"}, {"2", "30m", "2026-09-11"}} {
		if _, err := e.Ask(TimesheetAdd, given("number", spent[0], "time", spent[1], "on", spent[2])); err != nil {
			t.Fatal(err)
		}
	}
	answer, err := e.Ask(ObjectiveReport, given("number", "1", "since", "2026-09-01", "until", "2026-09-30"))
	if err != nil {
		t.Fatal(err)
	}
	months := answer.Period.Months
	if len(months) != 1 || months[0].Spent != 150*time.Minute ||
		!slices.Equal(months[0].Split, []objective.Part{{Number: 2, Spent: 2 * time.Hour}, {Spent: 30 * time.Minute}}) {
		t.Errorf("months = %+v", months)
	}
	shape := Of(ObjectiveReport, answer).(Review)
	if shape.Months[0].Itself != "30m" || len(shape.Months[0].SubObjectives) != 1 || shape.Months[0].SubObjectives[0].TimeSpent != "2h" {
		t.Errorf("the JSON month = %+v", shape.Months[0])
	}
}

func TestObjective_crossesWithItsParentAndChildren(t *testing.T) {
	e := openEngine(t)
	addObjective(t, e, "Home automation")
	under(t, e, "Enclosure", "1")
	listed, err := e.Ask(ObjectiveList, Input{})
	if err != nil {
		t.Fatal(err)
	}
	shape := Of(ObjectiveList, listed).(Objectives)
	if shape.Objectives[0].Parent != nil || !slices.Equal(shape.Objectives[0].Children, []int{2}) ||
		shape.Objectives[1].Parent == nil || *shape.Objectives[1].Parent != 1 || shape.Objectives[1].Children == nil {
		t.Errorf("the listing crosses as %+v", shape.Objectives)
	}
	shown, _ := e.Ask(ObjectiveShow, given("number", "1"))
	if said := Of(ObjectiveShow, shown).(Objective); !slices.Equal(said.Children, []int{2}) {
		t.Errorf("show crosses children %v", said.Children)
	}
}
