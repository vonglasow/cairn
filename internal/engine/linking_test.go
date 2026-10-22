package engine

import (
	"errors"
	"reflect"
	"testing"

	"cairn/internal/trace"
)

func linkedOf(t *testing.T, e *Engine, number int) []int {
	t.Helper()
	found, err := e.tasks.Find(number)
	if err != nil {
		t.Fatalf("Find(%d): %v", number, err)
	}
	return factsOf(found).Linked
}

func linking(t *testing.T, given ...string) Changes {
	t.Helper()
	c, err := taskChanges(Input{Listed: map[string][]string{"link": given}})
	if err != nil {
		t.Fatalf("taskChanges: %v", err)
	}
	return c
}

func TestLinkIsAppliedToTheTask(t *testing.T) {
	e := openEngine(t)
	add(t, e, "1")
	add(t, e, "2")
	add(t, e, "3")
	if _, err := e.add("4", linking(t, "1")); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if got := linkedOf(t, e, 4); !reflect.DeepEqual(got, []int{1}) {
		t.Errorf("task 4 linked = %v, want [1]", got)
	}
	if got := linkedOf(t, e, 1); !reflect.DeepEqual(got, []int{4}) {
		t.Errorf("task 1 linked = %v, want [4]", got)
	}
	if _, err := e.edit(4, linking(t, "+2")); err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if got := linkedOf(t, e, 4); !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("task 4 linked = %v, want [1 2]", got)
	}
	if _, err := e.edit(4, linking(t, "3")); err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if got := linkedOf(t, e, 4); !reflect.DeepEqual(got, []int{3}) {
		t.Errorf("task 4 linked = %v, want [3]", got)
	}
	if got := linkedOf(t, e, 1); len(got) != 0 {
		t.Errorf("task 1 linked = %v, want none", got)
	}
}

func TestLinkToItselfOrToNothingIsRefused(t *testing.T) {
	e := openEngine(t)
	add(t, e, "1")
	if _, err := e.edit(1, linking(t, "1")); !errors.Is(err, ErrLinkedToItself) {
		t.Errorf("edit 1 link 1: %v, want ErrLinkedToItself", err)
	}
	if _, err := e.edit(1, linking(t, "9")); !errors.Is(err, ErrNoSuchTask) {
		t.Errorf("edit 1 link 9: %v, want ErrNoSuchTask", err)
	}
	if _, err := e.add("x", linking(t, "9")); !errors.Is(err, ErrNoSuchTask) {
		t.Errorf("add link 9: %v, want ErrNoSuchTask", err)
	}
	if all, err := e.tasks.All(); err != nil || len(all) != 1 {
		t.Errorf("store holds %d tasks (%v), want 1", len(all), err)
	}
	add(t, e, "2")
	add(t, e, "3")
	if _, err := e.edit(2, Changes{State: of("done")}); err != nil {
		t.Fatalf("done: %v", err)
	}
	if err := e.archive(2); err != nil {
		t.Fatalf("archive: %v", err)
	}
	if _, err := e.edit(3, Changes{State: of("done")}); err != nil {
		t.Fatalf("done: %v", err)
	}
	if _, err := e.edit(1, linking(t, "2", "3")); err != nil {
		t.Errorf("link to archived and done: %v", err)
	}
}

func linksActsOn(t *testing.T, e *Engine, number int) [][2]string {
	t.Helper()
	acts, err := e.history.ActsOn(number)
	if err != nil {
		t.Fatalf("ActsOn(%d): %v", number, err)
	}
	var told [][2]string
	for _, a := range acts {
		c := a.Carried()
		if c.Field == "links" && c.Did == trace.Changed {
			told = append(told, [2]string{c.Was, c.Became})
		}
	}
	return told
}

func TestTheOtherEndIsToldOfTheLink(t *testing.T) {
	e := openEngine(t)
	add(t, e, "1")
	add(t, e, "2")
	if _, err := e.edit(2, linking(t, "+1")); err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if got, want := linksActsOn(t, e, 1), [][2]string{{"", "#2"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("acts on 1 = %v, want %v", got, want)
	}
	if got, want := linksActsOn(t, e, 2), [][2]string{{"", "#1"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("acts on 2 = %v, want %v", got, want)
	}
	if _, err := e.add("3", linking(t, "1")); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if got, want := linksActsOn(t, e, 1), [][2]string{{"", "#2"}, {"#2", "#2, #3"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("acts on 1 = %v, want %v", got, want)
	}
	if _, err := e.edit(3, linking(t, "-1")); err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if got, want := linksActsOn(t, e, 1), [][2]string{{"", "#2"}, {"#2", "#2, #3"}, {"#2, #3", "#2"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("acts on 1 = %v, want %v", got, want)
	}
	if _, err := e.edit(1, linking(t, "+2")); err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if got := len(linksActsOn(t, e, 1)); got != 3 {
		t.Errorf("acts on 1 = %d, want 3", got)
	}
	if got := len(linksActsOn(t, e, 2)); got != 1 {
		t.Errorf("acts on 2 = %d, want 1", got)
	}
}

func TestDeletingATaskTellsTheOtherEnd(t *testing.T) {
	e := openEngine(t)
	add(t, e, "1")
	add(t, e, "2")
	add(t, e, "3")
	if _, err := e.edit(2, linking(t, "1", "3")); err != nil {
		t.Fatalf("Edit: %v", err)
	}
	if err := e.delete(2); err != nil {
		t.Fatalf("delete: %v", err)
	}
	for _, n := range []int{1, 3} {
		told := linksActsOn(t, e, n)
		if len(told) == 0 {
			t.Fatalf("acts on %d: no links act", n)
		}
		if got, want := told[len(told)-1], [2]string{"#2", ""}; got != want {
			t.Errorf("last links act on %d = %v, want %v", n, got, want)
		}
	}
}
