package engine

import (
	"cairn/internal/engine/item"
	"cairn/internal/engine/task"
	"errors"
	"testing"
)

func TestBlocking_anEditClosingACycleIsRefused(t *testing.T) {
	e := openEngine(t)
	add(t, e, "Write the spec")
	add(t, e, "Ship it")
	if _, err := e.edit(2, Changes{Blocker: of([]string{"1"})}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.edit(1, Changes{Blocker: of([]string{"2"})}); !errors.Is(err, task.ErrBlockingCycle) {
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
	if f := freed.Carried(); f.Number != 2 || f.State != item.Todo {
		t.Errorf("freed = task %d in %q, want task 2 in todo", f.Number, f.State)
	}
}

func TestMayBlock_aBlockerReachedTwiceIsNoCycle(t *testing.T) {
	blocking := map[int][]int{2: {4}, 3: {4}}
	err := task.MayBlock(1, []int{2, 3}, func(n int) ([]int, error) { return blocking[n], nil })
	if err != nil {
		t.Errorf("two blockers waiting on the same third: %v", err)
	}
}
