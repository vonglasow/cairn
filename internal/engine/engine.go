package engine

import (
	"errors"
	"fmt"

	"cairn/internal/domain"
	"cairn/internal/store"
)

// ErrNothingToChange is returned when an edit names nothing to change. Doing
// nothing quietly would look like success to whoever asked.
var ErrNothingToChange = errors.New("nothing was named to change")

// ErrStore is returned when the store itself cannot be used — as opposed to a
// request the store understood and refused. A surface tells the two apart to
// answer with the right exit code, so a script knows whether to retry or to
// fix its arguments.
var ErrStore = errors.New("the store cannot be read")

// ErrLabelSignAtCreation is returned when a task is created with a label
// carrying a + or a -. Those say how an existing set changes, and a task being
// created has no set yet — accepting them would mean `--label -spec` quietly
// doing nothing at all.
var ErrLabelSignAtCreation = errors.New("a label is added or taken away only when changing a task")

// ErrNoSuchTask is re-exported so a surface can recognise it: a surface asks
// the engine and may not import the store, so the engine is where the name it
// answers with has to live.
var ErrNoSuchTask = store.ErrNoSuchTask

// The reasons a task cannot hang where it was asked to, or be deleted.
var (
	ErrParentIsSelf    = errors.New("a task cannot hang under itself")
	ErrParentIsSubTask = errors.New("a sub-task cannot carry sub-tasks")
	ErrCarriesSubTasks = errors.New("the task carries sub-tasks")
)

// tasks is what the engine needs from whatever keeps its tasks. Declared here,
// by the consumer, so the rules stay testable and a second backing can arrive
// without the engine noticing.
type tasks interface {
	Keep(domain.Task) (domain.Task, error)
	Update(domain.Task) error
	Delete(int) error
	Find(int) (domain.Task, error)
	All() ([]domain.Task, error)
	Under(int) ([]domain.Task, error)
	Close() error
}

// Engine carries out the operations. Every change of state goes through it.
type Engine struct {
	tasks      tasks
	objectives objectives
	store      *store.Store
}

// Changes are the attributes an operation was given. A nil field was not
// named at all; a field pointing at an empty value was named and emptied —
// which is how an attribute is taken away.
type Changes struct {
	Title       *string
	Description *string
	State       *string
	Priority    *string
	Labels      *[]string
	Deadline    *string
	Estimate    *string
	Spent       *string
	Parent      *int
	Objective   *int
}

// named says whether anything at all was given to change.
func (c Changes) named() bool {
	return c.Title != nil || c.Description != nil || c.State != nil ||
		c.Priority != nil || c.Labels != nil || c.Deadline != nil ||
		c.Estimate != nil || c.Spent != nil || c.Parent != nil || c.Objective != nil
}

// Open opens a store that has to be up to date, and refuses one that is not:
// a command about tasks does not change the shape of the store on the way past.
func Open(path string) (*Engine, error) { return openWith(path, store.Open) }

// OpenBehind opens a store whatever shape it is in, for the commands whose job
// is to change that shape.
func OpenBehind(path string) (*Engine, error) { return openWith(path, store.OpenBehind) }

// openWith builds the engine over whichever way the store was opened.
func openWith(path string, open func(string) (*store.Store, error)) (*Engine, error) {
	kept, err := open(path)
	if err != nil {
		return nil, storeProblem(err)
	}
	return &Engine{tasks: kept, objectives: kept, store: kept}, nil
}

// storeProblem marks an error as being about the store itself rather than about
// the request. An error that already says which kind of problem it is passes
// through untouched: wrapping "the store is behind" in "the store cannot be
// read" would say two things, one of them false.
func storeProblem(err error) error {
	if errors.Is(err, ErrStoreBehind) || errors.Is(err, ErrNewerStore) {
		return err
	}
	return fmt.Errorf("%w: %w", ErrStore, err)
}

// Close releases the store.
func (e *Engine) Close() error {
	return e.tasks.Close()
}

// Add records a new task carrying what it was given, and nothing else.
func (e *Engine) Add(title string, c Changes) (domain.Task, error) {
	task, err := domain.NewTask(title)
	if err != nil {
		return domain.Task{}, err
	}
	if task, err = apply(task, c, creating); err != nil {
		return domain.Task{}, err
	}
	if err := e.canHang(task); err != nil {
		return domain.Task{}, err
	}
	if err := e.canServe(task); err != nil {
		return domain.Task{}, err
	}
	return e.tasks.Keep(task)
}

// List hands back every task, in the order they were added.
func (e *Engine) List() ([]domain.Task, error) {
	return e.tasks.All()
}

// Show hands back the task carrying number.
func (e *Engine) Show(number int) (domain.Task, error) {
	return e.tasks.Find(number)
}

// Delete removes the task carrying number, refusing while something still
// hangs under it: what depends on the task would go with it, and a deletion
// that takes more than the number typed is a deletion nobody asked for.
func (e *Engine) Delete(number int) error {
	hanging, err := e.tasks.Under(number)
	if err != nil {
		return err
	}
	if len(hanging) > 0 {
		return fmt.Errorf("task %d: %w", number, ErrCarriesSubTasks)
	}
	return e.tasks.Delete(number)
}

// Under hands back what hangs directly under a task.
func (e *Engine) Under(number int) ([]domain.Task, error) {
	return e.tasks.Under(number)
}

// canHang says whether a task may hang where it was told to. A parent has to
// exist, cannot be the task itself, and cannot already hang under something —
// what hangs under a task is never something with more hanging under it.
func (e *Engine) canHang(task domain.Task) error {
	if task.Parent == domain.NoParent {
		return nil
	}
	if task.Objective != domain.NoObjective {
		return ErrSubTaskServesNoObjective
	}
	if task.Parent == task.Number {
		return fmt.Errorf("task %d: %w", task.Number, ErrParentIsSelf)
	}
	parent, err := e.tasks.Find(task.Parent)
	if err != nil {
		return err
	}
	if parent.Parent != domain.NoParent {
		return fmt.Errorf("task %d: %w", parent.Number, ErrParentIsSubTask)
	}
	return nil
}

// Edit changes what was named on a task and leaves the rest alone.
func (e *Engine) Edit(number int, c Changes) (domain.Task, error) {
	if !c.named() {
		return domain.Task{}, ErrNothingToChange
	}
	task, err := e.tasks.Find(number)
	if err != nil {
		return domain.Task{}, err
	}
	if task, err = apply(task, c, changing); err != nil {
		return domain.Task{}, err
	}
	if err := e.canHang(task); err != nil {
		return domain.Task{}, err
	}
	if err := e.canServe(task); err != nil {
		return domain.Task{}, err
	}
	if err := e.tasks.Update(task); err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

// Whether a task is being created or changed. Labels are the only attribute
// that cares: a sign says how an existing set moves, and there is no set to
// move at creation.
const (
	creating = true
	changing = false
)

// changedLabels works out the set that a change leaves behind. A sign says how
// an existing set moves, and there is no set to move at creation, so one given
// there is refused rather than read as a label named "+urgent".
//
// Shared by tasks and objectives: both carry a set and both move it the same
// way, and this is the one part of applying a change that is more than a call
// to a parser.
func changedLabels(given, current []string, atCreation bool) ([]string, error) {
	change, err := domain.ParseLabelChange(given)
	if err != nil {
		return nil, err
	}
	if atCreation && change.Adjusts() {
		return nil, ErrLabelSignAtCreation
	}
	return change.Apply(current)
}

// apply lays the changes over a task, refusing any value it cannot make sense
// of. A field that was not named is left exactly as it was.
func apply(task domain.Task, c Changes, atCreation bool) (domain.Task, error) {
	if c.Title != nil {
		retitled, err := domain.NewTask(*c.Title)
		if err != nil {
			return domain.Task{}, err
		}
		task.Title = retitled.Title
	}
	if c.Description != nil {
		task.Description = *c.Description
	}
	if c.State != nil {
		state, err := domain.ParseState(*c.State)
		if err != nil {
			return domain.Task{}, err
		}
		task.State = state
	}
	if c.Priority != nil {
		priority, err := domain.ParsePriority(*c.Priority)
		if err != nil {
			return domain.Task{}, err
		}
		task.Priority = priority
	}
	if c.Labels != nil {
		labels, err := changedLabels(*c.Labels, task.Labels, atCreation)
		if err != nil {
			return domain.Task{}, err
		}
		task.Labels = labels
	}
	if c.Deadline != nil {
		deadline, err := domain.ParseDeadline(*c.Deadline)
		if err != nil {
			return domain.Task{}, err
		}
		task.Deadline = deadline
	}
	if c.Spent != nil {
		spent, err := domain.ParseSpent(*c.Spent)
		if err != nil {
			return domain.Task{}, err
		}
		task.TimeSpent = spent.Apply(task.TimeSpent)
	}
	if c.Parent != nil {
		task.Parent = *c.Parent
	}
	if c.Objective != nil {
		task.Objective = *c.Objective
	}
	if c.Estimate != nil {
		estimate, err := domain.ParseEstimate(*c.Estimate)
		if err != nil {
			return domain.Task{}, err
		}
		task.Estimate = estimate
	}
	return task, nil
}
