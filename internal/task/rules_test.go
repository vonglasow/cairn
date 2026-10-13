package task

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"cairn/internal/domain"
)

func standing(n int, state domain.State, because string, blockers ...int) Task {
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

func TestMayFinishAs_asksWhatHangsUnderItOnlyWhenItIsBeingFinished(t *testing.T) {
	asked := 0
	open := []Task{standing(2, domain.Todo, "")}
	under := func(int) ([]Task, error) { asked++; return open, nil }
	if err := standing(1, domain.Todo, "").MayFinishAs(standing(1, domain.Current, ""), under); err != nil || asked != 0 {
		t.Errorf("moving to current asked %d times and said %v", asked, err)
	}
	if err := standing(1, domain.Done, "").MayFinishAs(standing(1, domain.Done, ""), under); err != nil || asked != 0 {
		t.Errorf("staying done asked %d times and said %v", asked, err)
	}
	if err := standing(1, domain.Todo, "").MayFinishAs(standing(1, domain.Done, ""), under); !errors.Is(err, ErrSubTaskStillOpen) {
		t.Errorf("finishing over an open part = %v, want ErrSubTaskStillOpen", err)
	}
	broken := errors.New("disk on fire")
	if err := standing(1, domain.Todo, "").MayFinishAs(standing(1, domain.Done, ""),
		func(int) ([]Task, error) { return nil, broken }); !errors.Is(err, broken) {
		t.Errorf("a store that could not answer = %v, want its failure", err)
	}
}

func part(n, parent int) Task {
	var a AsTold
	a.Numbered(n)
	a.Titled("a part")
	a.InColumn(domain.Todo)
	a.Under(parent)
	return a.Task()
}

func TestMayHang_isAskedOfTheTaskAndOfItsParent(t *testing.T) {
	parents := map[int]Task{1: standing(1, domain.Todo, ""), 2: part(2, 1)}
	parentOf := func(n int) (Task, error) {
		found, ok := parents[n]
		if !ok {
			return Task{}, errors.New("no such task")
		}
		return found, nil
	}
	if err := standing(5, domain.Current, "").MayHang(parentOf); err != nil {
		t.Errorf("a task hanging under nothing was refused: %v", err)
	}
	if err := part(5, 1).MayHang(parentOf); err != nil {
		t.Errorf("a part under a task was refused: %v", err)
	}
	if err := part(5, 2).MayHang(parentOf); !errors.Is(err, ErrParentIsSubTask) {
		t.Errorf("a part under a part = %v, want ErrParentIsSubTask", err)
	}
	if err := part(5, 5).MayHang(parentOf); !errors.Is(err, ErrParentIsSelf) {
		t.Errorf("a part under itself = %v, want ErrParentIsSelf", err)
	}
	if err := part(5, 9).MayHang(parentOf); err == nil {
		t.Error("a part under a task nobody holds was accepted")
	}
	var a AsTold
	a.Numbered(5)
	a.Titled("a part")
	a.InColumn(domain.Current)
	a.Under(1)
	if err := a.Task().MayHang(parentOf); !errors.Is(err, ErrSubTaskHasNoColumn) {
		t.Errorf("a part in current = %v, want ErrSubTaskHasNoColumn", err)
	}
	a.InColumn(domain.Todo)
	a.Serves(3)
	if err := a.Task().MayHang(parentOf); !errors.Is(err, ErrSubTaskServesNoObjective) {
		t.Errorf("a part serving an objective = %v, want ErrSubTaskServesNoObjective", err)
	}
}

func TestWork_leavesThePartsOut(t *testing.T) {
	got := Work([]Task{standing(1, domain.Done, ""), part(2, 1), standing(3, domain.Done, "")})
	var numbers []int
	for _, one := range got {
		numbers = append(numbers, one.number)
	}
	if !slices.Equal(numbers, []int{1, 3}) {
		t.Errorf("Work kept %v, want [1 3]", numbers)
	}
}

func TestMayLeaveFor_refusesOnlyTheColumnsThatClaimTheWork(t *testing.T) {
	for _, tc := range []struct {
		what    string
		was     Task
		becomes Task
		refused bool
	}{
		{"into todo, still blocked",
			standing(1, domain.Blocked, "", 2),
			standing(1, domain.Todo, "", 2), true},
		{"into current, blocked by words",
			standing(1, domain.Blocked, "the client"),
			standing(1, domain.Current, "the client"), true},
		{"into done, still blocked",
			standing(1, domain.Blocked, "", 2),
			standing(1, domain.Done, "", 2), false},
		{"into todo with nothing recorded",
			standing(1, domain.Blocked, ""),
			standing(1, domain.Todo, ""), false},
		{"into todo from somewhere that was never blocked",
			standing(1, domain.Backlog, "", 2),
			standing(1, domain.Todo, "", 2), false},
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
	err := standing(1, domain.Blocked, "the client's answer", 2, 3).
		MayLeaveFor(standing(1, domain.Todo, "the client's answer", 2, 3))
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
			standing(1, domain.Blocked, "", 2),
			standing(1, domain.Blocked, ""), true},
		{"the last words",
			standing(1, domain.Blocked, "the client"),
			standing(1, domain.Blocked, ""), true},
		{"one of two",
			standing(1, domain.Blocked, "", 2, 3),
			standing(1, domain.Blocked, "", 3), false},
		{"the blocker, with words left",
			standing(1, domain.Blocked, "the client", 2),
			standing(1, domain.Blocked, "the client"), false},
		{"a task that was not blocked",
			standing(1, domain.Todo, "", 2),
			standing(1, domain.Todo, ""), false},
		{"a task leaving the column itself",
			standing(1, domain.Blocked, "", 2),
			standing(1, domain.Done, ""), false},
	} {
		freed, put := tc.was.FreedBy(tc.becomes)
		if put != tc.freed {
			t.Errorf("%s: freed = %v, want %v", tc.what, put, tc.freed)
			continue
		}
		if put && freed != domain.Todo {
			t.Errorf("%s: it landed in %q, want todo", tc.what, freed)
		}
	}
}

func TestFreedBy_actsOnTheWordsBeingErasedAndNeverOnTheirMeaning(t *testing.T) {
	if _, put := standing(1, domain.Blocked, "the client").
		FreedBy(standing(1, domain.Blocked, "the client")); put {
		t.Error("a task moved with its words untouched")
	}
	if _, put := standing(1, domain.Blocked, "the client").
		FreedBy(standing(1, domain.Blocked, "")); !put {
		t.Error("erasing the last thing in its way left it blocked")
	}
}
