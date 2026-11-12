package engine

import (
	"errors"
	"strings"
	"testing"

	"cairn/internal/domain"
)

func TestBlocking_aBlockerFurtherAlongTheChainMakesACycle(t *testing.T) {
	further := map[int][]int{2: {3}, 3: {1}}
	err := mayBlock(1, []int{2}, func(one int) ([]int, error) { return further[one], nil })
	if !errors.Is(err, ErrBlockingCycle) {
		t.Errorf("err = %v, want ErrBlockingCycle", err)
	}
}

func TestBlocking_aCycleNamesTheBlockerThatClosesIt(t *testing.T) {
	further := map[int][]int{4: {1}}
	err := mayBlock(1, []int{4}, func(one int) ([]int, error) { return further[one], nil })
	if err == nil || !strings.Contains(err.Error(), "task 1 and task 4") {
		t.Errorf("err = %v, want it to name task 1 and task 4", err)
	}
}

func TestBlocking_anEditClosingACycleIsRefused(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")
	add(t, e, "Ship it")
	if _, err := e.edit(2, Changes{Blocker: of([]string{"1"})}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.edit(1, Changes{Blocker: of([]string{"2"})}); !errors.Is(err, ErrBlockingCycle) {
		t.Errorf("err = %v, want ErrBlockingCycle", err)
	}
}

func TestBlocking_aBlockedTaskLosingItsLastBlockerIsFreed(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")
	add(t, e, "Ship it")
	if _, err := e.edit(2, Changes{State: of("blocked"), Blocker: of([]string{"1"})}); err != nil {
		t.Fatal(err)
	}
	freed, err := e.edit(2, Changes{Blocker: of([]string{"-1"})})
	if err != nil {
		t.Fatal(err)
	}
	if f := factsOf(freed); f.Number != 2 || f.State != domain.Todo {
		t.Errorf("freed = task %d in %q, want task 2 in todo", f.Number, f.State)
	}
}

func TestMayBlock_aBlockerReachedTwiceIsNoCycle(t *testing.T) {
	blocking := map[int][]int{2: {4}, 3: {4}}
	err := mayBlock(1, []int{2, 3}, func(n int) ([]int, error) { return blocking[n], nil })
	if err != nil {
		t.Errorf("two blockers waiting on the same third: %v", err)
	}
}

func blockedBy(blocking map[int][]int) func(int) ([]int, error) {
	return func(n int) ([]int, error) { return blocking[n], nil }
}

func TestMayBlock_findsACycleBehindABlockerReachedTwice(t *testing.T) {
	if err := mayBlock(1, []int{2, 3}, blockedBy(map[int][]int{2: {4}, 3: {4}, 4: {1}})); !errors.Is(err, ErrBlockingCycle) {
		t.Errorf("1 waits on 4 through 2 and 3, and 4 on 1: %v", err)
	}
}

func TestMayBlock_namesTheBlockerTheCycleCameBackThrough(t *testing.T) {
	if err := mayBlock(1, []int{2}, blockedBy(map[int][]int{2: {3}, 3: {1}})); err == nil || !strings.HasPrefix(err.Error(), "task 1 and task 2:") {
		t.Errorf("1 waits on 2, 2 on 3, 3 on 1: %v", err)
	}
	if err := mayBlock(1, []int{2, 3}, blockedBy(map[int][]int{2: {1}, 3: {1}})); err == nil || !strings.HasPrefix(err.Error(), "task 1 and task 3:") {
		t.Errorf("1 waits on 2 and 3, each on 1: %v", err)
	}
}
