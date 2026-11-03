package engine

import (
	"slices"
	"strings"
	"testing"

	"cairn/internal/objective"
)

func oneList(name string, values ...string) Input {
	return Input{Given: map[string]string{"number": "1"}, Listed: map[string][]string{name: values}}
}

func TestTaskEdit_saysWhichLabelsWereAddedAndTakenOff(t *testing.T) {
	e := openEngine(t)
	if _, err := e.add("Write the spec", Changes{Labels: of([]string{"spec", "team"})}); err != nil {
		t.Fatal(err)
	}
	a, err := e.Ask(TaskEdit, oneList("label", "=spec,urgent"))
	if err != nil {
		t.Fatal(err)
	}
	if a.Moved == nil || !slices.Equal(a.Moved.Added.Labels, []string{"urgent"}) || !slices.Equal(a.Moved.TakenOff.Labels, []string{"team"}) {
		t.Errorf("Moved = %+v, want urgent added and team taken off", a.Moved)
	}
}

func TestTaskEdit_saysNothingMovedWhenNothingDid(t *testing.T) {
	e := openEngine(t)
	if _, err := e.add("Write the spec", Changes{Labels: of([]string{"spec"})}); err != nil {
		t.Fatal(err)
	}
	a, err := e.Ask(TaskEdit, oneList("label", "spec"))
	if err != nil {
		t.Fatal(err)
	}
	if a.Moved != nil {
		t.Errorf("Moved = %+v, want nothing: the labels are what they were", a.Moved)
	}
}

func TestTaskEdit_saysWhichNumbersWereAddedAndTakenOff(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")
	add(t, e, "Write the tests")
	add(t, e, "Write the docs")
	if _, err := e.edit(3, Changes{Blocker: of([]string{"1"}), Link: of([]string{"1"})}); err != nil {
		t.Fatal(err)
	}
	in := Input{Given: map[string]string{"number": "3"}, Listed: map[string][]string{"blocker": {"=2"}, "link": {"2"}}}
	a, err := e.Ask(TaskEdit, in)
	if err != nil {
		t.Fatal(err)
	}
	if a.Moved == nil || !slices.Equal(a.Moved.Added.Blockers, []int{2}) || !slices.Equal(a.Moved.TakenOff.Blockers, []int{1}) ||
		!slices.Equal(a.Moved.Added.Links, []int{2}) || len(a.Moved.TakenOff.Links) != 0 {
		t.Errorf("Moved = %+v, want blocker 2 for 1, and link 2 added", a.Moved)
	}
}

func TestTaskEdit_saysWhichObjectivesWereAdded(t *testing.T) {
	e := openEngine(t)
	if _, err := e.addObjective("Ship it", ObjectiveChanges{}); err != nil {
		t.Fatal(err)
	}
	add(t, e, "Write the spec")
	a, err := e.Ask(TaskEdit, oneList("objective", "1"))
	if err != nil {
		t.Fatal(err)
	}
	if a.Moved == nil || !slices.Equal(a.Moved.Added.Objectives, []int{1}) {
		t.Errorf("Moved = %+v, want objective 1 added", a.Moved)
	}
}

func TestObjectiveEdit_saysWhichLabelsWereTakenOff(t *testing.T) {
	e := openEngine(t)
	if _, err := e.addObjective("Ship it", ObjectiveChanges{Labels: of([]string{"perso", "spec"})}); err != nil {
		t.Fatal(err)
	}
	a, err := e.Ask(ObjectiveEdit, oneList("label", "=urgent"))
	if err != nil {
		t.Fatal(err)
	}
	if a.Moved == nil || !slices.Equal(a.Moved.TakenOff.Labels, []string{"perso", "spec"}) || !slices.Equal(a.Moved.Added.Labels, []string{"urgent"}) {
		t.Errorf("Moved = %+v, want urgent for perso and spec", a.Moved)
	}
}

func TestRenderText_saysWhatWasAddedAndTakenOff(t *testing.T) {
	one := builtTask(t, 3, "Ship it")
	moved := &Moved{Added: Sets{Labels: []string{"urgent", "team"}, Links: []int{4}}, TakenOff: Sets{Blockers: []int{1, 2}, Objectives: []int{5}}}
	got := RenderText(TaskEdit, Answer{Task: &one, Moved: moved})
	for _, want := range []string{"labels added      urgent, team\n", "links added       #4\n", "blockers taken off  #1, #2\n", "objectives taken off  #5\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("the answer does not say %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "labels taken off") || strings.Contains(got, "blockers added") {
		t.Errorf("the answer names what did not move:\n%s", got)
	}
}

func TestRenderText_saysWhichLabelsAnObjectiveLost(t *testing.T) {
	o := builtObjective(t, 1, "Ship it")
	got := RenderText(ObjectiveEdit, Answer{Objective: &o, Moved: &Moved{TakenOff: Sets{Labels: []string{"perso"}}}})
	if !strings.Contains(got, "labels taken off  perso\n") {
		t.Errorf("the answer does not say perso was taken off:\n%s", got)
	}
}

func TestOf_carriesWhatWasAddedAndTakenOff(t *testing.T) {
	one := builtTask(t, 3, "Ship it")
	moved := &Moved{Added: Sets{Labels: []string{"urgent"}}, TakenOff: Sets{Blockers: []int{1}}}
	shaped, _ := Of(TaskEdit, Answer{Task: &one, Moved: moved}).(Task)
	if shaped.Added == nil || !slices.Equal(shaped.Added.Labels, []string{"urgent"}) ||
		shaped.TakenOff == nil || !slices.Equal(shaped.TakenOff.Blockers, []int{1}) {
		t.Errorf("Of = %+v %+v, want urgent added and 1 taken off", shaped.Added, shaped.TakenOff)
	}
	if plain, _ := Of(TaskEdit, Answer{Task: &one}).(Task); plain.Added != nil || plain.TakenOff != nil {
		t.Errorf("Of without a change = %+v %+v, want nothing said", plain.Added, plain.TakenOff)
	}
	o := builtObjective(t, 1, "Ship it")
	goal, _ := Of(ObjectiveEdit, Answer{Objective: &o, Moved: &Moved{TakenOff: Sets{Labels: []string{"perso"}}}}).(Objective)
	if goal.TakenOff == nil || !slices.Equal(goal.TakenOff.Labels, []string{"perso"}) {
		t.Errorf("Of an objective = %+v, want perso taken off", goal.TakenOff)
	}
}

func TestEngine_addsAPlainLabelWhenEditing(t *testing.T) {
	e := openEngine(t)
	if _, err := e.add("Write the spec", Changes{Labels: of([]string{"spec"})}); err != nil {
		t.Fatal(err)
	}
	edited, err := e.edit(1, Changes{Labels: of([]string{"team"})})
	if err != nil || !slices.Equal(factsOf(edited).Labels, []string{"spec", "team"}) {
		t.Errorf("edit = %v, %v; want spec and team", factsOf(edited).Labels, err)
	}
}

func TestEngine_takesWhatAnEqualsSignSaysAtCreation(t *testing.T) {
	made, err := openEngine(t).add("Write the spec", Changes{Labels: of([]string{"=spec,team"}), Blocker: of([]string{"="})})
	if err != nil || !slices.Equal(factsOf(made).Labels, []string{"spec", "team"}) {
		t.Errorf("add = %v, %v; want spec and team", factsOf(made).Labels, err)
	}
}

func TestAsked_replacesWhatAListHeld(t *testing.T) {
	in := Asked([]Field{{Name: "labels", Arg: "label", Value: "api urgent"}}, TaskEdit)
	if got := in.Listed["label"]; !slices.Equal(got, []string{"=", "api", "urgent"}) {
		t.Errorf("label = %v, want the form's labels replacing the ones there", got)
	}
}

func builtObjective(t *testing.T, number int, title string) objective.Objective {
	t.Helper()
	built, err := objective.NewObjective(title)
	if err != nil {
		t.Fatalf("NewObjective(%q): %v", title, err)
	}
	told := objective.From(built)
	told.Numbered(number)
	return told.Objective()
}

func TestSets_areEmptyOnlyWhenNothingMoved(t *testing.T) {
	if !(Sets{}).empty() {
		t.Error("nothing at all is not empty")
	}
	for _, one := range []Sets{{Labels: []string{"spec"}}, {Blockers: []int{1}}, {Links: []int{1}}, {Objectives: []int{1}}} {
		if one.empty() {
			t.Errorf("%+v is empty", one)
		}
	}
}
