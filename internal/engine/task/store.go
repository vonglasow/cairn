package task

import (
	"cairn/internal/engine/filter"
	"errors"
)

// Store is where the tasks are kept.
type Store interface {
	Keep(Task) (Task, error)
	Update(Task) error
	Delete(int) error
	Find(int) (Task, error)
	All() ([]Task, error)
	Narrowed(*filter.Narrowing) ([]Task, error)
	NothingBlocks() ([]int, error)
	Close() error
}

// Places is where the tasks sit in their column, and which are in focus.
type Places interface {
	MoveFirst(int) (bool, error)
	MoveLast(int) (bool, error)
	MoveAfter(int, int) (bool, error)
	Focus() ([]int, error)
	AddFocus(int) error
	DropFocus(int) error
	ClearFocus() error
}

// ErrNoSuchTask is returned when a number refers to no task.
var ErrNoSuchTask = errors.New("there is no such task")
