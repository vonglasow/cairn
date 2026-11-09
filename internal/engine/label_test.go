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
	if !ok || len(said.Labels) != 2 || said.Labels[0] != (LabelCarried{Label: "doc", Tasks: 1, Objectives: 1, Colour: domain.ColourOf("doc"), Scope: "both"}) || said.Labels[1] != (LabelCarried{Label: "sp26-20", Tasks: 2, Colour: domain.ColourOf("sp26-20"), Scope: "both"}) {
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
	if !ok || len(said.Labels) != 1 || said.Labels[0] != (LabelCarried{Label: "release-train", Colour: domain.ColourOf("release-train"), Scope: "both"}) {
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

func made(pairs ...string) func() (Operation, Input) {
	return func() (Operation, Input) { return LabelNew, given(pairs...) }
}

func scopesOf(t *testing.T, path string, in Input) map[string]string {
	t.Helper()
	listed, err := Do(LabelList, path, in)
	if err != nil {
		t.Fatal(err)
	}
	said := map[string]string{}
	for _, one := range *listed.Labels {
		said[one.Label] = one.Scope
	}
	return said
}

func TestLabels_carryAScope(t *testing.T) {
	path := storeAt(t)
	ran(t, path,
		made("label", "Area", "scope", "objectives"),
		made("label", "chore", "scope", "tasks"),
		made("label", "doc"),
	)
	if got := scopesOf(t, path, Input{}); got["area"] != "objectives" || got["chore"] != "tasks" || got["doc"] != "both" {
		t.Fatalf("the scopes are %v", got)
	}
	if got := scopesOf(t, path, given("for", "objectives")); len(got) != 2 || got["chore"] != "" {
		t.Errorf("the labels offered for objectives are %v", got)
	}
	if got := scopesOf(t, path, given("for", "tasks")); len(got) != 2 || got["area"] != "" {
		t.Errorf("the labels offered for tasks are %v", got)
	}
	listed, _ := Do(LabelList, path, Input{})
	if text := RenderText(LabelList, listed); !strings.Contains(text, "for objectives") || !strings.Contains(text, "for tasks") || !strings.Contains(text, "for both") {
		t.Errorf("a person reads:\n%s", text)
	}
	if _, err := Do(LabelScope, path, given("label", "area", "scope", "both")); err != nil {
		t.Fatal(err)
	}
	if got := scopesOf(t, path, Input{}); got["area"] != "both" {
		t.Errorf("area is for %q once put back to both", got["area"])
	}
	if _, err := Do(LabelScope, path, given("label", "doc", "scope", "everything")); !errors.Is(err, domain.ErrNoSuchScope) {
		t.Errorf("a scope that is none of the three reads %v", err)
	}
	if _, err := Do(LabelScope, path, given("label", "nope", "scope", "tasks")); !errors.Is(err, ErrNoSuchLabel) {
		t.Errorf("scoping a label nothing carries reads %v", err)
	}
	if _, err := Do(LabelNew, path, given("label", "x", "scope", "nowhere")); !errors.Is(err, domain.ErrNoSuchScope) {
		t.Errorf("a label made with no such scope reads %v", err)
	}
}

func TestLabels_outOfTheirScopeAreRefused(t *testing.T) {
	path := storeAt(t)
	ran(t, path,
		made("label", "area", "scope", "objectives"),
		made("label", "chore", "scope", "tasks"),
		asked(TaskNew, "title", "Write the spec", "label", "chore"),
		asked(ObjectiveNew, "title", "Home", "label", "area"),
	)
	for _, refused := range []func() (Operation, Input){
		asked(TaskEdit, "number", "1", "label", "area"),
		asked(TaskNew, "title", "Other", "label", "area"),
		asked(ObjectiveEdit, "number", "1", "label", "chore"),
		asked(ObjectiveNew, "title", "Other", "label", "chore"),
	} {
		op, in := refused()
		if _, err := Do(op, path, in); !errors.Is(err, domain.ErrOutOfScope) {
			t.Errorf("%s %v reads %v", op.Line(), in.Listed, err)
		}
	}
	ran(t, path, asked(TaskEdit, "number", "1", "title", "Write the specs"))
	if _, err := Do(LabelScope, path, given("label", "chore", "scope", "objectives")); err == nil || !strings.Contains(err.Error(), `"chore" is carried by 1 task`) {
		t.Errorf("narrowing a label something outside the scope carries reads %v", err)
	}
	if _, err := Do(LabelScope, path, given("label", "area", "scope", "tasks")); err == nil || !strings.Contains(err.Error(), `"area" is carried by 1 objective`) {
		t.Errorf("narrowing a label an objective carries to tasks reads %v", err)
	}
}
