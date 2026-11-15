package engine

import (
	"cairn/internal/engine/label"
	"errors"
	"strings"
	"testing"
)

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
	if labels := (*shown.Task).Carried().Labels; strings.Join(labels, ",") != "doc,sprint-20" {
		t.Errorf("task 1 carries %v", labels)
	}
	goal, _ := Do(ObjectiveShow, path, given("number", "1"))
	if labels := (*goal.Objective).Carried().Labels; strings.Join(labels, ",") != "sprint-20" {
		t.Errorf("the objective carries %v", labels)
	}
	if _, err := Do(LabelRename, path, Input{Given: map[string]string{"label": "doc", "to": "sprint-20"}}); err != nil {
		t.Fatal(err)
	}
	shown, _ = Do(TaskShow, path, given("number", "1"))
	if labels := (*shown.Task).Carried().Labels; strings.Join(labels, ",") != "sprint-20" {
		t.Errorf("joined, task 1 carries %v", labels)
	}
	if _, err := Do(LabelRename, path, Input{Given: map[string]string{"label": "nope", "to": "x"}}); !errors.Is(err, label.ErrNoSuchLabel) {
		t.Errorf("renaming what nothing carries answered %v", err)
	}
}

func TestALabelWithNoNameIsRefusedNotMade(t *testing.T) {
	path := storeAt(t)
	if _, err := Do(LabelNew, path, Input{Given: map[string]string{"label": ""}}); !errors.Is(err, label.ErrLabelRequired) {
		t.Errorf("a label with no name answers %v", err)
	}
	if _, err := Do(LabelNew, path, Input{Given: map[string]string{"label": "doc"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Do(LabelRename, path, Input{Given: map[string]string{"label": "doc", "to": "  "}}); !errors.Is(err, label.ErrLabelRequired) {
		t.Errorf("a rename to no name answers %v", err)
	}
	if _, err := Do(LabelDelete, path, Input{Given: map[string]string{"label": ""}, Confirmed: true}); !errors.Is(err, label.ErrLabelRequired) {
		t.Errorf("deleting no name answers %v", err)
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
	if _, err := Do(labelScope, path, given("label", "area", "scope", "both")); err != nil {
		t.Fatal(err)
	}
	if got := scopesOf(t, path, Input{}); got["area"] != "both" {
		t.Errorf("area is for %q once put back to both", got["area"])
	}
	if _, err := Do(labelScope, path, given("label", "doc", "scope", "everything")); !errors.Is(err, label.ErrNoSuchScope) {
		t.Errorf("a scope that is none of the three reads %v", err)
	}
	if _, err := Do(labelScope, path, given("label", "nope", "scope", "tasks")); !errors.Is(err, label.ErrNoSuchLabel) {
		t.Errorf("scoping a label nothing carries reads %v", err)
	}
	if _, err := Do(LabelNew, path, given("label", "x", "scope", "nowhere")); !errors.Is(err, label.ErrNoSuchScope) {
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
		if _, err := Do(op, path, in); !errors.Is(err, label.ErrOutOfScope) {
			t.Errorf("%s %v reads %v", op.Line(), in.Listed, err)
		}
	}
	ran(t, path, asked(TaskEdit, "number", "1", "title", "Write the specs"))
	if _, err := Do(labelScope, path, given("label", "chore", "scope", "objectives")); err == nil || !strings.Contains(err.Error(), `"chore" is carried by 1 task`) {
		t.Errorf("narrowing a label something outside the scope carries reads %v", err)
	}
	if _, err := Do(labelScope, path, given("label", "area", "scope", "tasks")); err == nil || !strings.Contains(err.Error(), `"area" is carried by 1 objective`) {
		t.Errorf("narrowing a label an objective carries to tasks reads %v", err)
	}
}
