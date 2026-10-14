package engine

import (
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
	if !ok || len(said.Labels) != 2 || said.Labels[0] != (LabelCarried{Label: "doc", Tasks: 1, Objectives: 1}) || said.Labels[1] != (LabelCarried{Label: "sp26-20", Tasks: 2}) {
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
