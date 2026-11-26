// Package engine holds the operations.
package engine

import (
	"cairn/internal/engine/filter"
	"cairn/internal/engine/item"
	"cairn/internal/engine/label"
	"cairn/internal/engine/objective"
	"errors"
	"fmt"
	"slices"
	"sync"

	"cairn/internal/engine/act"
	"cairn/internal/engine/migration"
	"cairn/internal/engine/task"
	"cairn/internal/store/files"
	"cairn/internal/store/sqlite"
)

// errNothingToChange is returned when an edit names nothing to change.
var errNothingToChange = errors.New("nothing was named to change")

// errStore is returned when the store itself cannot be used — as opposed to a request the store understood and refused.
var errStore = errors.New("the store cannot be read")

// errLabelSignAtCreation is returned when a task is created with a label carrying a + or a -.
var errLabelSignAtCreation = errors.New("a label is added or taken away only when changing a task")

// Engine carries out the operations.
type Engine struct {
	tasks      task.Store
	places     task.Places
	history    act.Store
	objectives objective.Store
	store      *sqlite.Store
	acting     Acting
	notes      string
	home       Home
	path       string
	known      []migration.Migration
	turn       *sync.Mutex
	followed   int64
	sweepOwed  bool
	refused    error
}

// Open opens a store that has to be up to date, and refuses one that is not: a command about tasks does not change the shape of the store on the way past.
func Open(path string) (*Engine, error) { return openWith(path, migration.Known, sqlite.OpenKnowing) }

// OpenAt opens the store the way OpenBehind does, whatever shape it is in, and keeps where the notes are and what they were found from, for a surface that holds an engine open — a board — rather than one per call.
func OpenAt(at Places) (*Engine, error) {
	e, err := OpenBehind(at.Store)
	if err != nil {
		return nil, Told(err)
	}
	e.notes, e.home = at.Notes, at.Home
	return e, nil
}

// OpenBehind opens a store whatever shape it is in, for the commands whose job is to change that shape.
func OpenBehind(path string) (*Engine, error) {
	return openWith(path, migration.Known, sqlite.OpenBehindKnowing)
}

// OpenBehindKnowing opens a store the way OpenBehind does, for a cairn that knows the changes given rather than this one's.
func OpenBehindKnowing(path string, known []migration.Migration) (*Engine, error) {
	return openWith(path, known, sqlite.OpenBehindKnowing)
}

func openWith(path string, known []migration.Migration, open func(string, []migration.Migration) (*sqlite.Store, error)) (*Engine, error) {
	kept, err := open(path, known)
	if err != nil {
		return nil, storeProblem(err)
	}
	return (&Engine{path: path, known: known, turn: &sync.Mutex{}}).keptBy(kept), nil
}

func (e *Engine) taking() func() {
	e.turn.Lock()
	return e.turn.Unlock
}

func (e *Engine) following() error {
	if e.refused != nil {
		return nil
	}
	replaced, err := e.store.Replaced()
	if err != nil || !replaced {
		return err
	}
	fresh, err := sqlite.OpenKnowing(e.path, e.known)
	if err != nil {
		return storeProblem(err)
	}
	stale := e.store
	*e = *e.keptBy(fresh)
	e.followed++
	_ = stale.Close()
	return nil
}

func (e Engine) keptBy(s *sqlite.Store) *Engine {
	e.tasks, e.places, e.history, e.objectives, e.store = spentTasks{s, s}, s, s, s, s
	return &e
}

func storeProblem(err error) error {
	if errors.Is(err, migration.ErrStoreBehind) || errors.Is(err, migration.ErrNewerStore) {
		return err
	}
	return fmt.Errorf("%w: %w", errStore, err)
}

// Close releases the store.
func (e *Engine) Close() error {
	defer e.taking()()
	if e.refused != nil {
		return nil
	}
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
	if err := e.inScope(nil, t.Carried().Labels, label.OnTasks); err != nil {
		return task.Task{}, err
	}
	if err := e.canServe(t); err != nil {
		return task.Task{}, err
	}
	if err := task.CanBeLinked(e.tasks, t); err != nil {
		return task.Task{}, err
	}
	kept, err := e.tasks.Keep(t)
	if err != nil {
		return task.Task{}, err
	}
	if err := e.recordCreated(kept); err != nil {
		return task.Task{}, err
	}
	if err := e.spendOn(kept.Carried().Number, 0, t.Carried().Spent); err != nil {
		return task.Task{}, err
	}
	return e.tasks.Find(kept.Carried().Number)
}

func (e *Engine) narrowed(kept *filter.Narrowing) ([]task.Task, error) {
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

func (e *Engine) sweeping(kept *filter.Narrowing) ([]task.Task, error) {
	found, err := e.tasks.Narrowed(kept)
	if err != nil {
		return nil, err
	}
	return found, nil
}

func (e *Engine) archiveEvery(kept *filter.Narrowing) (int, error) {
	found, err := e.sweeping(kept)
	if err != nil {
		return 0, err
	}
	swept := 0
	for _, one := range found {
		if err := e.archive(one.Carried().Number); err != nil {
			if errors.Is(err, task.ErrAlreadyArchived) {
				continue
			}
			return swept, err
		}
		swept++
	}
	return swept, nil
}

func (e *Engine) labelEvery(kept *filter.Narrowing, name string) (int, error) {
	found, err := e.sweeping(kept)
	if err != nil {
		return 0, err
	}
	put := 0
	adding := []string{fmt.Sprintf("+%s", name)}
	for _, one := range found {
		if _, err := e.edit(one.Carried().Number, Changes{Labels: &adding}); err != nil {
			return put, err
		}
		put++
	}
	return put, nil
}

func (e *Engine) delete(number int) error {
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
	if err := e.recordOtherEnds(number, gone.Carried().Linked, nil); err != nil {
		return err
	}
	if slices.Contains(held, number) {
		if err := e.record(number, act.FocusReleased, "", "", ""); err != nil {
			return err
		}
	}
	went := act.WhatItHeld(gone)
	if e.timeGoesWithTasks() {
		went += "; " + act.WentWithItsTime
	}
	return e.record(number, act.Deleted, "", went, "")
}

func (e *Engine) timeGoesWithTasks() bool {
	if e.home.Dir == "" {
		return false
	}
	read, err := e.configured()
	return err == nil && read.TimeDeleted() == files.TimeTaken
}

func (e *Engine) configured() (files.Places, error) { return configuredAt(e.home, e.path) }

func (e *Engine) waitingOn(number int) ([]int, error) {
	every := filter.Narrowing{}
	every.WithArchived()
	all, err := e.tasks.Narrowed(&every)
	if err != nil {
		return nil, err
	}
	var waiting []int
	for _, t := range all {
		f := t.Carried()
		if slices.Contains(f.Blockers, number) {
			waiting = append(waiting, f.Number)
		}
	}
	slices.Sort(waiting)
	return waiting, nil
}

func (e *Engine) edit(number int, c Changes) (task.Task, error) {
	if !c.named() {
		return task.Task{}, errNothingToChange
	}
	was, err := e.tasks.Find(number)
	if err != nil {
		return task.Task{}, err
	}
	t := was
	if t, err = apply(t, c, changing); err != nil {
		return task.Task{}, err
	}
	if err := e.inScope(was.Carried().Labels, t.Carried().Labels, label.OnTasks); err != nil {
		return task.Task{}, err
	}
	if err := e.canServe(t); err != nil {
		return task.Task{}, err
	}
	if err := task.CanBeBlocked(e.tasks, t); err != nil {
		return task.Task{}, err
	}
	if err := task.CanBeLinked(e.tasks, t); err != nil {
		return task.Task{}, err
	}
	if err := was.MayLeaveFor(t); err != nil {
		return task.Task{}, err
	}
	if c.Finishing && t.Carried().State == item.Done {
		t = t.Ticked()
	}
	if err := was.MayFinishAs(t); errors.Is(err, ErrStepStillOpen) {
		return task.Task{}, fmt.Errorf("%w; --yes puts it in done too", err)
	} else if err != nil {
		return task.Task{}, err
	}
	if err := e.tasks.Update(t); err != nil {
		return task.Task{}, err
	}
	if err := e.spendOn(number, was.Carried().Spent, t.Carried().Spent); err != nil {
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
	after, err := e.tasks.Find(found.Carried().Number)
	if err != nil {
		return task.Task{}, err
	}
	if err := e.recordChanges(found, after); err != nil {
		return task.Task{}, err
	}
	return after, nil
}
