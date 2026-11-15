package engine

import (
	"slices"
	"strings"
	"testing"

	"cairn/internal/engine/objective"
)

func TestRenderText_saysWhatWasAddedAndTakenOff(t *testing.T) {
	one := builtTask(t, 3, "Ship it")
	moved := &moves{Added: sets{Labels: []string{"urgent", "team"}, Links: []int{4}}, TakenOff: sets{Blockers: []int{1, 2}, Objectives: []int{5}}}
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

func TestOf_carriesWhatWasAddedAndTakenOff(t *testing.T) {
	one := builtTask(t, 3, "Ship it")
	moved := &moves{Added: sets{Labels: []string{"urgent"}}, TakenOff: sets{Blockers: []int{1}}}
	shaped, _ := Of(TaskEdit, Answer{Task: &one, Moved: moved}).(Task)
	if shaped.Added == nil || !slices.Equal(shaped.Added.Labels, []string{"urgent"}) ||
		shaped.TakenOff == nil || !slices.Equal(shaped.TakenOff.Blockers, []int{1}) {
		t.Errorf("Of = %+v %+v, want urgent added and 1 taken off", shaped.Added, shaped.TakenOff)
	}
	if plain, _ := Of(TaskEdit, Answer{Task: &one}).(Task); plain.Added != nil || plain.TakenOff != nil {
		t.Errorf("Of without a change = %+v %+v, want nothing said", plain.Added, plain.TakenOff)
	}
	o := builtObjective(t, 1, "Ship it")
	goal, _ := Of(ObjectiveEdit, Answer{Objective: &o, Moved: &moves{TakenOff: sets{Labels: []string{"perso"}}}}).(objectiveRow)
	if goal.TakenOff == nil || !slices.Equal(goal.TakenOff.Labels, []string{"perso"}) {
		t.Errorf("Of an objective = %+v, want perso taken off", goal.TakenOff)
	}
}

func TestEngine_takesWhatAnEqualsSignSaysAtCreation(t *testing.T) {
	made, err := openEngine(t).add("Write the spec", Changes{Labels: of([]string{"=spec,team"}), Blocker: of([]string{"="})})
	if err != nil || !slices.Equal(made.Carried().Labels, []string{"spec", "team"}) {
		t.Errorf("add = %v, %v; want spec and team", made.Carried().Labels, err)
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
