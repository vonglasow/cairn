package board

import (
	"cairn/internal/engine"
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"cairn/internal/engine/label"
	"cairn/internal/engine/objective"
	"cairn/internal/engine/task"
)

func TestAGoalIsToldItHasADescription(t *testing.T) {
	built, err := objective.NewObjective("Ship it")
	if err != nil {
		t.Fatalf("building an objective: %v", err)
	}
	plainGoal, err := goalOf(built, nil)
	if err != nil {
		t.Fatalf("reading it: %v", err)
	}
	if plainGoal.described() {
		t.Fatal("an objective with nothing written about it was said to have a description")
	}

	told := objective.From(built)
	told.Described("why it matters")
	described, err := goalOf(told.Objective(), nil)
	if err != nil {
		t.Fatalf("reading it: %v", err)
	}
	if !described.described() {
		t.Fatal("an objective with something written about it was said to have none")
	}
}

func TestLook_readsEachCardWithItsStepsAndTheFocus(t *testing.T) {
	told := task.From(aTask("Ship it"))
	told.Numbered(1)
	told.Described("why it matters")
	told.Step("Write it", false)
	told.Step("Test it", true)
	f := newFake()
	f.tasks = []task.Task{told.Task(), aTask("Bare")}
	f.focus = []int{1}
	s, err := Look(f)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(s.Focus, []int{1}) {
		t.Errorf("focus = %v, want [1]", s.Focus)
	}
	c := s.Cards[0]
	if c.Number != 1 || !c.InFocus || !c.Described() {
		t.Errorf("card = %+v", c)
	}
	if want := []engine.Part{{Number: 1, Title: "Write it"}, {Number: 2, Title: "Test it", Done: true}}; !slices.Equal(c.Steps, want) {
		t.Errorf("steps = %+v, want %+v", c.Steps, want)
	}
	if bare := s.Cards[1]; bare.Steps != nil || bare.Described() {
		t.Errorf("a bare card = %+v", bare)
	}
}

func TestCardOf_saysWhereInTheFocusATaskCame(t *testing.T) {
	c, err := cardOf(builtTask(t, 1, "a"), []int{5, 1}, nil, nil, nil)
	if err != nil || c.Focused != 2 || !c.InFocus {
		t.Errorf("card %+v, %v; want second in focus", c, err)
	}
}

func TestLook_saysOnEachCardHowManyTasksItIsInTheWayOf(t *testing.T) {
	told := task.From(aTask("Write the spec"))
	told.Numbered(1)
	other := task.From(aTask("Ship it"))
	other.Numbered(2)
	f := newFake()
	f.tasks = []task.Task{told.Task(), other.Task()}
	f.chosen = map[int]task.Choosing{1: {Blocks: []int{2, 3}}}
	s, err := Look(f)
	if err != nil {
		t.Fatal(err)
	}
	if s.Cards[0].BlocksSaid() != "⊳2" || s.Cards[1].BlocksSaid() != "" {
		t.Errorf("cards say %q and %q, want ⊳2 and nothing", s.Cards[0].BlocksSaid(), s.Cards[1].BlocksSaid())
	}
}

func TestALabelWearsTheColourChosenOrOneOfItsName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cairn.db")
	if _, err := engine.Do(engine.TaskNew, path, engine.Input{Given: map[string]string{"title": "one"}, Listed: map[string][]string{"label": {"doc", "release"}}}); err != nil {
		t.Fatal(err)
	}
	coloured, err := engine.Do(engine.LabelColour, path, engine.Input{Given: map[string]string{"label": "release", "colour": "Peach"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, one := range engine.Of(engine.LabelColour, coloured).(label.List).Labels {
		want := map[string]string{"release": "peach", "doc": label.ColourOf("doc")}[one.Label]
		if one.Colour != want {
			t.Errorf("%s wears %q, want %q", one.Label, one.Colour, want)
		}
	}
	if _, err := engine.Do(engine.LabelColour, path, engine.Input{Given: map[string]string{"label": "release", "colour": "brown"}}); !errors.Is(err, label.ErrNoSuchColour) {
		t.Errorf("brown answers %v", err)
	}
	opened, err := engine.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = opened.Close() }()
	looked, err := Look(opened)
	if err != nil {
		t.Fatal(err)
	}
	if looked.Colours["release"] != "peach" || looked.Colours["doc"] != label.ColourOf("doc") {
		t.Errorf("the board reads the colours %v", looked.Colours)
	}
}
