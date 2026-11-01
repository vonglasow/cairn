package engine

import (
	"cairn/internal/domain"
	"errors"
	"strings"
	"testing"
)

func TestLabels_areListedWithWhatCarriesThem(t *testing.T) {
	path := storeAt(t)
	for _, in := range []Input{
		{Given: map[string]string{"title": "one"}, Listed: map[string][]string{"label": {"doc", "sp26-20"}}},
		{Given: map[string]string{"title": "two"}, Listed: map[string][]string{"label": {"sp26-20"}}},
	} {
		if _, err := Do(TaskNew, path, in); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Do(ObjectiveNew, path, Input{Given: map[string]string{"title": "goal"}, Listed: map[string][]string{"label": {"doc"}}}); err != nil {
		t.Fatal(err)
	}
	listed, err := Do(LabelList, path, Input{})
	if err != nil {
		t.Fatal(err)
	}
	said, ok := Of(LabelList, listed).(Labels)
	if !ok || len(said.Labels) != 2 || said.Labels[0] != (LabelCarried{Label: "doc", Tasks: 1, Objectives: 1, Colour: domain.ColourOf("doc")}) || said.Labels[1] != (LabelCarried{Label: "sp26-20", Tasks: 2, Colour: domain.ColourOf("sp26-20")}) {
		t.Errorf("the labels read %+v", Of(LabelList, listed))
	}
	if text := RenderText(LabelList, listed); !strings.Contains(text, "sp26-20") || !strings.Contains(text, "2 tasks") {
		t.Errorf("a person reads:\n%s", text)
	}
	if text := RenderText(LabelList, Answer{Labels: &[]LabelCarried{}}); !strings.Contains(text, "no label") {
		t.Errorf("no label reads %q", text)
	}
}

func TestLabels_areDeletedFromEverythingCarryingThem(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskNew, path, Input{Given: map[string]string{"title": "one"}, Listed: map[string][]string{"label": {"doc", "sp26-20"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Do(ObjectiveNew, path, Input{Given: map[string]string{"title": "goal"}, Listed: map[string][]string{"label": {"sp26-20"}}}); err != nil {
		t.Fatal(err)
	}
	if !LabelDelete.AsksFirst(Input{}) {
		t.Error("deleting a label everywhere does not ask first")
	}
	if _, err := Do(LabelDelete, path, Agreed(Input{Given: map[string]string{"label": "SP26 20"}})); err != nil {
		t.Fatal(err)
	}
	shown, _ := Do(TaskShow, path, given("number", "1"))
	if labels := factsOf(*shown.Task).Labels; strings.Join(labels, ",") != "doc" {
		t.Errorf("task 1 carries %v", labels)
	}
	goal, _ := Do(ObjectiveShow, path, given("number", "1"))
	if labels := factsOfObjective(*goal.Objective).Labels; len(labels) != 0 {
		t.Errorf("the objective carries %v", labels)
	}
	if _, err := Do(LabelDelete, path, Agreed(Input{Given: map[string]string{"label": "nope"}})); !errors.Is(err, ErrNoSuchLabel) {
		t.Errorf("deleting what nothing carries answered %v", err)
	}
}

func TestLabels_areRenamedOnEverythingCarryingThem(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskNew, path, Input{Given: map[string]string{"title": "one"}, Listed: map[string][]string{"label": {"doc", "sp26-20"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Do(ObjectiveNew, path, Input{Given: map[string]string{"title": "goal"}, Listed: map[string][]string{"label": {"sp26-20"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Do(LabelRename, path, Input{Given: map[string]string{"label": "sp26-20", "to": "Sprint 20"}}); err != nil {
		t.Fatal(err)
	}
	shown, _ := Do(TaskShow, path, given("number", "1"))
	if labels := factsOf(*shown.Task).Labels; strings.Join(labels, ",") != "doc,sprint-20" {
		t.Errorf("task 1 carries %v", labels)
	}
	goal, _ := Do(ObjectiveShow, path, given("number", "1"))
	if labels := factsOfObjective(*goal.Objective).Labels; strings.Join(labels, ",") != "sprint-20" {
		t.Errorf("the objective carries %v", labels)
	}
	if _, err := Do(LabelRename, path, Input{Given: map[string]string{"label": "doc", "to": "sprint-20"}}); err != nil {
		t.Fatal(err)
	}
	shown, _ = Do(TaskShow, path, given("number", "1"))
	if labels := factsOf(*shown.Task).Labels; strings.Join(labels, ",") != "sprint-20" {
		t.Errorf("joined, task 1 carries %v", labels)
	}
	if _, err := Do(LabelRename, path, Input{Given: map[string]string{"label": "nope", "to": "x"}}); !errors.Is(err, ErrNoSuchLabel) {
		t.Errorf("renaming what nothing carries answered %v", err)
	}
}

func TestALabelIsMadeOnItsOwnAndListedCarriedByNothing(t *testing.T) {
	path := storeAt(t)
	made, err := Do(LabelNew, path, Input{Given: map[string]string{"label": "Release Train"}})
	if err != nil {
		t.Fatal(err)
	}
	said, ok := Of(LabelNew, made).(Labels)
	if !ok || len(said.Labels) != 1 || said.Labels[0] != (LabelCarried{Label: "release-train", Colour: domain.ColourOf("release-train")}) {
		t.Fatalf("making a label answers %+v", Of(LabelNew, made))
	}
	if _, err := Do(LabelRename, path, Input{Given: map[string]string{"label": "release-train", "to": "train"}}); err != nil {
		t.Fatalf("renaming a label nothing carries: %v", err)
	}
	gone, err := Do(LabelDelete, path, Input{Given: map[string]string{"label": "train"}, Confirmed: true})
	if err != nil {
		t.Fatalf("deleting a label nothing carries: %v", err)
	}
	if left := Of(LabelDelete, gone).(Labels); len(left.Labels) != 0 {
		t.Errorf("deleted, the labels left are %+v", left)
	}
	if _, err := Do(LabelNew, path, Input{Given: map[string]string{"label": "a:b"}}); err == nil {
		t.Error("a label the rule refuses was made")
	}
}

func TestALabelWithNoNameIsRefusedNotMade(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(LabelNew, path, Input{Given: map[string]string{"label": ""}}); !errors.Is(err, domain.ErrLabelRequired) {
		t.Errorf("a label with no name answers %v", err)
	}
	if _, err := Do(LabelNew, path, Input{Given: map[string]string{"label": "doc"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Do(LabelRename, path, Input{Given: map[string]string{"label": "doc", "to": "  "}}); !errors.Is(err, domain.ErrLabelRequired) {
		t.Errorf("a rename to no name answers %v", err)
	}
	if _, err := Do(LabelDelete, path, Input{Given: map[string]string{"label": ""}, Confirmed: true}); !errors.Is(err, domain.ErrLabelRequired) {
		t.Errorf("deleting no name answers %v", err)
	}
}

func TestALabelWearsTheColourChosenOrOneOfItsName(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(TaskNew, path, Input{Given: map[string]string{"title": "one"}, Listed: map[string][]string{"label": {"doc", "release"}}}); err != nil {
		t.Fatal(err)
	}
	coloured, err := Do(LabelColour, path, Input{Given: map[string]string{"label": "release", "colour": "Peach"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, one := range Of(LabelColour, coloured).(Labels).Labels {
		want := map[string]string{"release": "peach", "doc": domain.ColourOf("doc")}[one.Label]
		if one.Colour != want {
			t.Errorf("%s wears %q, want %q", one.Label, one.Colour, want)
		}
	}
	if _, err := Do(LabelColour, path, Input{Given: map[string]string{"label": "release", "colour": "brown"}}); !errors.Is(err, domain.ErrNoSuchColour) {
		t.Errorf("brown answers %v", err)
	}
	opened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = opened.Close() }()
	looked, err := Look(opened)
	if err != nil {
		t.Fatal(err)
	}
	if looked.Colours["release"] != "peach" || looked.Colours["doc"] != domain.ColourOf("doc") {
		t.Errorf("the board reads the colours %v", looked.Colours)
	}
}
