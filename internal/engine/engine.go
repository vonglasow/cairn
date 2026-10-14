// Package engine holds the operations.
package engine

import (
	"errors"
	"fmt"
	"slices"
	"sync"

	"cairn/internal/domain"
	"cairn/internal/store"
	"cairn/internal/task"
	"cairn/internal/trace"
)

// ErrNothingToChange is returned when an edit names nothing to change.
var ErrNothingToChange = errors.New("nothing was named to change")

// ErrStore is returned when the store itself cannot be used — as opposed to a request the store understood and refused.
var ErrStore = errors.New("the store cannot be read")

// ErrLabelSignAtCreation is returned when a task is created with a label carrying a + or a -.
var ErrLabelSignAtCreation = errors.New("a label is added or taken away only when changing a task")

// ErrNoSuchTask is re-exported so a surface can recognise it: a surface asks the engine and may not import the store, so the engine is where the name it answers with has to live.
var ErrNoSuchTask = store.ErrNoSuchTask

// The reasons a task cannot hang where it was asked to.
var (
	ErrParentIsSelf    = task.ErrParentIsSelf
	ErrParentIsSubTask = task.ErrParentIsSubTask
)

type tasks interface {
	Keep(task.Task) (task.Task, error)
	Update(task.Task) error
	Delete(int) error
	Find(int) (task.Task, error)
	All() ([]task.Task, error)
	Narrowed(*domain.Narrowing) ([]task.Task, error)
	NothingBlocks() ([]int, error)
	Under(int) ([]task.Task, error)
	Close() error
}

// Engine carries out the operations.
type Engine struct {
	tasks      tasks
	places     places
	history    history
	objectives objectives
	store      *store.Store
	through    string
	notes      string
	home       Home
	path       string
	turn       *sync.Mutex
	followed   int64
}

// Open opens a store that has to be up to date, and refuses one that is not: a command about tasks does not change the shape of the store on the way past.
func Open(path string) (*Engine, error) { return openWith(path, store.Open) }

// OpenAt opens the store the way Open does, and keeps where the notes are and what they were found from, for a surface that holds an engine open — a board — rather than one per call.
func OpenAt(at Places) (*Engine, error) {
	e, err := Open(at.Store)
	if err != nil {
		return nil, Told(err)
	}
	e.notes, e.home = at.Notes, at.Home
	return e, nil
}

// OpenBehind opens a store whatever shape it is in, for the commands whose job is to change that shape.
func OpenBehind(path string) (*Engine, error) { return openWith(path, store.OpenBehind) }

func openWith(path string, open func(string) (*store.Store, error)) (*Engine, error) {
	kept, err := open(path)
	if err != nil {
		return nil, storeProblem(err)
	}
	return (&Engine{path: path, turn: &sync.Mutex{}}).keptBy(kept), nil
}

func (e *Engine) taking() func() {
	e.turn.Lock()
	return e.turn.Unlock
}

func (e *Engine) following() error {
	replaced, err := e.store.Replaced()
	if err != nil || !replaced {
		return err
	}
	fresh, err := store.Open(e.path)
	if err != nil {
		return storeProblem(err)
	}
	stale := e.store
	*e = *e.keptBy(fresh)
	e.followed++
	_ = stale.Close()
	return nil
}

func (e Engine) keptBy(s *store.Store) *Engine {
	e.tasks, e.places, e.history, e.objectives, e.store = spentTasks{s, s}, s, s, s, s
	return &e
}

func storeProblem(err error) error {
	if errors.Is(err, ErrStoreBehind) || errors.Is(err, ErrNewerStore) {
		return err
	}
	return fmt.Errorf("%w: %w", ErrStore, err)
}

// Close releases the store.
func (e *Engine) Close() error {
	defer e.taking()()
	return e.tasks.Close()
}

func (e *Engine) add(title string, c Changes) (task.Task, error) {
	t, err := task.NewTask(title)
	if err != nil {
		return task.Task{}, err
	}
	if t, err = apply(t, c, creating); err != nil {
		return task.Task{}, err
	}
	t = t.Settled()
	if err := t.MayHang(e.tasks.Find); err != nil {
		return task.Task{}, err
	}
	if err := e.canServe(t); err != nil {
		return task.Task{}, err
	}
	kept, err := e.tasks.Keep(t)
	if err != nil {
		return task.Task{}, err
	}
	if err := e.recordCreated(kept); err != nil {
		return task.Task{}, err
	}
	if err := e.spendOn(factsOf(kept).Number, 0, factsOf(t).Spent); err != nil {
		return task.Task{}, err
	}
	return e.tasks.Find(factsOf(kept).Number)
}

func (e *Engine) narrowed(kept *domain.Narrowing) ([]task.Task, error) {
	return e.tasks.Narrowed(kept)
}

func (e *Engine) show(number int) (task.Task, error) {
	return e.tasks.Find(number)
}

func (e *Engine) archive(number int) error { return e.puttingAway(number, true) }

func (e *Engine) unarchive(number int) error { return e.puttingAway(number, false) }

func (e *Engine) puttingAway(number int, away bool) error {
	found, err := e.tasks.Find(number)
	if err != nil {
		return err
	}
	may := found.MayUnarchive
	if away {
		may = found.MayArchive
	}
	if err := may(); err != nil {
		return err
	}
	a := found.Carried()
	a.Archived = away
	put := task.Made(a)
	if err := e.tasks.Update(put); err != nil {
		return err
	}
	return e.recordChanges(found, put)
}

func (e *Engine) sweeping(kept *domain.Narrowing) ([]task.Task, error) {
	found, err := e.tasks.Narrowed(kept)
	if err != nil {
		return nil, err
	}
	return task.Work(found), nil
}

func (e *Engine) archiveEvery(kept *domain.Narrowing) (int, error) {
	found, err := e.sweeping(kept)
	if err != nil {
		return 0, err
	}
	swept := 0
	for _, one := range found {
		if err := e.archive(factsOf(one).Number); err != nil {
			if errors.Is(err, task.ErrAlreadyArchived) {
				continue
			}
			return swept, err
		}
		swept++
	}
	return swept, nil
}

func (e *Engine) labelEvery(kept *domain.Narrowing, label string) (int, error) {
	found, err := e.sweeping(kept)
	if err != nil {
		return 0, err
	}
	put := 0
	adding := []string{"+" + label}
	for _, one := range found {
		if _, err := e.edit(factsOf(one).Number, Changes{Labels: &adding}); err != nil {
			return put, err
		}
		put++
	}
	return put, nil
}

func (e *Engine) delete(number int) error {
	hanging, err := e.tasks.Under(number)
	if err != nil {
		return err
	}
	for _, part := range hanging {
		if err := e.delete(factsOf(part).Number); err != nil {
			return err
		}
	}
	held, err := e.places.Focus()
	if err != nil {
		return err
	}
	gone, err := e.tasks.Find(number)
	if err != nil {
		return err
	}
	if err := e.tasks.Delete(number); err != nil {
		return err
	}
	if slices.Contains(held, number) {
		if err := e.record(number, trace.FocusReleased, "", "", ""); err != nil {
			return err
		}
	}
	return e.record(number, trace.Deleted, "", whatItHeld(gone), "")
}

func (e *Engine) under(number int) ([]task.Task, error) {
	return e.tasks.Under(number)
}

func (e *Engine) waitingOn(number int) ([]int, error) {
	every := domain.Narrowing{}
	every.WithArchived()
	all, err := e.tasks.Narrowed(&every)
	if err != nil {
		return nil, err
	}
	var waiting []int
	for _, t := range all {
		f := factsOf(t)
		if slices.Contains(f.Blockers, number) {
			waiting = append(waiting, f.Number)
		}
	}
	slices.Sort(waiting)
	return waiting, nil
}

func (e *Engine) edit(number int, c Changes) (task.Task, error) {
	if !c.named() {
		return task.Task{}, ErrNothingToChange
	}
	was, err := e.tasks.Find(number)
	if err != nil {
		return task.Task{}, err
	}
	t := was
	if t, err = apply(t, c, changing); err != nil {
		return task.Task{}, err
	}
	if factsOf(was).Parent != factsOf(t).Parent {
		t = t.Settled()
	}
	if err := t.MayHang(e.tasks.Find); err != nil {
		return task.Task{}, err
	}
	if err := e.canServe(t); err != nil {
		return task.Task{}, err
	}
	if err := e.canBeBlocked(t); err != nil {
		return task.Task{}, err
	}
	if err := was.MayLeaveFor(t); err != nil {
		return task.Task{}, err
	}
	if err := was.MayFinishAs(t, e.tasks.Under); err != nil {
		return task.Task{}, err
	}
	if err := e.tasks.Update(t); err != nil {
		return task.Task{}, err
	}
	if err := e.spendOn(number, factsOf(was).Spent, factsOf(t).Spent); err != nil {
		return task.Task{}, err
	}
	found, err := e.tasks.Find(number)
	if err != nil {
		return task.Task{}, err
	}
	if err := e.recordChanges(was, found); err != nil {
		return task.Task{}, err
	}
	return e.freeIfNothingBlocksIt(was, found)
}

func (e *Engine) freeIfNothingBlocksIt(was, found task.Task) (task.Task, error) {
	freed, put := was.FreedBy(found)
	if !put {
		return found, nil
	}
	into := string(freed)
	free, err := apply(found, Changes{State: &into}, changing)
	if err != nil {
		return task.Task{}, err
	}
	if err := e.tasks.Update(free); err != nil {
		return task.Task{}, err
	}
	after, err := e.tasks.Find(factsOf(found).Number)
	if err != nil {
		return task.Task{}, err
	}
	if err := e.recordChanges(found, after); err != nil {
		return task.Task{}, err
	}
	return after, nil
}
