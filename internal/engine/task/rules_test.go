package task

import (
	"cairn/internal/engine/item"
	"strings"
	"testing"
)

func standing(n int, state item.State, because string, blockers ...int) Task {
	var a AsTold
	a.Numbered(n)
	a.Titled("Ship it")
	a.InColumn(state)
	for _, b := range blockers {
		a.BlockedBy(b)
	}
	if because != "" {
		a.Because(because)
	}
	return a.Task()
}

func TestMayLeaveFor_refusesOnlyTheColumnsThatClaimTheWork(t *testing.T) {
	for _, tc := range []struct {
		what    string
		was     Task
		becomes Task
		refused bool
	}{
		{"into todo, still blocked",
			standing(1, item.Blocked, "", 2),
			standing(1, item.Todo, "", 2), true},
		{"into current, blocked by words",
			standing(1, item.Blocked, "the client"),
			standing(1, item.Current, "the client"), true},
		{"into done, still blocked",
			standing(1, item.Blocked, "", 2),
			standing(1, item.Done, "", 2), false},
		{"into todo with nothing recorded",
			standing(1, item.Blocked, ""),
			standing(1, item.Todo, ""), false},
		{"into todo from somewhere that was never blocked",
			standing(1, item.Backlog, "", 2),
			standing(1, item.Todo, "", 2), false},
	} {
		err := tc.was.MayLeaveFor(tc.becomes)
		if tc.refused && err == nil {
			t.Errorf("%s: it was allowed", tc.what)
		}
		if !tc.refused && err != nil {
			t.Errorf("%s: it was refused: %v", tc.what, err)
		}
	}
}

func TestMayLeaveFor_saysWhatIsInTheWayAndHowToTakeItAway(t *testing.T) {
	err := standing(1, item.Blocked, "the client's answer", 2, 3).
		MayLeaveFor(standing(1, item.Todo, "the client's answer", 2, 3))
	if err == nil {
		t.Fatal("it was allowed")
	}
	for _, said := range []string{"#2", "#3", "the client's answer",
		"task 2", "task 3", "the reason it is waiting"} {
		if !strings.Contains(err.Error(), said) {
			t.Errorf("the refusal does not say %q: %v", said, err)
		}
	}
}

func TestFreedBy_putsATaskBackOnlyWhenTheLastThingGoes(t *testing.T) {
	for _, tc := range []struct {
		what    string
		was     Task
		becomes Task
		freed   bool
	}{
		{"the last blocker",
			standing(1, item.Blocked, "", 2),
			standing(1, item.Blocked, ""), true},
		{"the last words",
			standing(1, item.Blocked, "the client"),
			standing(1, item.Blocked, ""), true},
		{"one of two",
			standing(1, item.Blocked, "", 2, 3),
			standing(1, item.Blocked, "", 3), false},
		{"the blocker, with words left",
			standing(1, item.Blocked, "the client", 2),
			standing(1, item.Blocked, "the client"), false},
		{"a task that was not blocked",
			standing(1, item.Todo, "", 2),
			standing(1, item.Todo, ""), false},
		{"a task leaving the column itself",
			standing(1, item.Blocked, "", 2),
			standing(1, item.Done, ""), false},
	} {
		freed, put := tc.was.FreedBy(tc.becomes)
		if put != tc.freed {
			t.Errorf("%s: freed = %v, want %v", tc.what, put, tc.freed)
			continue
		}
		if put && freed != item.Todo {
			t.Errorf("%s: it landed in %q, want todo", tc.what, freed)
		}
	}
}

func TestFreedBy_actsOnTheWordsBeingErasedAndNeverOnTheirMeaning(t *testing.T) {
	if _, put := standing(1, item.Blocked, "the client").
		FreedBy(standing(1, item.Blocked, "the client")); put {
		t.Error("a task moved with its words untouched")
	}
	if _, put := standing(1, item.Blocked, "the client").
		FreedBy(standing(1, item.Blocked, "")); !put {
		t.Error("erasing the last thing in its way left it blocked")
	}
}
