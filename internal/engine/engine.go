// Package engine holds the operations. A surface parses input, calls one thing
// here, and prints what comes back — it decides nothing. The engine is also the
// only caller of the store, whatever the store is made of.
package engine

import (
	"errors"
	"fmt"
	"slices"

	"cairn/internal/domain"
	"cairn/internal/store"
	"cairn/internal/task"
	"cairn/internal/trace"
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

// The reasons a task cannot hang where it was asked to. Re-exported from task,
// which is where a task's own rules live, so a surface can recognise every
// refusal by asking the engine for its name.
var (
	ErrParentIsSelf    = task.ErrParentIsSelf
	ErrParentIsSubTask = task.ErrParentIsSubTask
)

// tasks is what the engine needs from whatever keeps its tasks. Declared here,
// by the consumer, so the rules stay testable against a fake and the engine
// names only what it uses. It is not what would let a second backing arrive
// unnoticed: transactions, migrations and backups still reach the store itself.
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

// Engine carries out the operations. Every change of state goes through it.
type Engine struct {
	tasks      tasks
	places     places
	history    history
	objectives objectives
	store      *store.Store
	// through is the surface that asked, written into every act recorded. Empty
	// is the terminal: a history says who acted only when it was not the person
	// reading it.
	through string
	// notes is the root every task's outside context sits under, when the
	// surface that opened this engine said where; empty otherwise.
	notes string
	// home is what the places were found from, for the one operation that
	// answers how they were found.
	home Home
	// path is where the store was opened. Kept because an operation is handed
	// what sits beside the store as well as the store itself, and a surface
	// that stays open runs operations long after whoever opened it has gone.
	path string
}

// Open opens a store that has to be up to date, and refuses one that is not:
// a command about tasks does not change the shape of the store on the way past.
func Open(path string) (*Engine, error) { return openWith(path, store.Open) }

// OpenAt opens the store the way Open does, and keeps where the notes are and
// what they were found from, for a surface that holds an engine open — a
// board — rather than one per call. A store that cannot be used is Told, as an
// operation's would be.
func OpenAt(at Places) (*Engine, error) {
	e, err := Open(at.Store)
	if err != nil {
		return nil, Told(err)
	}
	e.notes, e.home = at.Notes, at.Home
	return e, nil
}

// OpenBehind opens a store whatever shape it is in, for the commands whose job
// is to change that shape.
func OpenBehind(path string) (*Engine, error) { return openWith(path, store.OpenBehind) }

// openWith builds the engine over whichever way the store was opened.
func openWith(path string, open func(string) (*store.Store, error)) (*Engine, error) {
	kept, err := open(path)
	if err != nil {
		return nil, storeProblem(err)
	}
	return (&Engine{path: path}).keptBy(kept), nil
}

// keptBy is this engine with every door it reaches the store through handed
// the one given — the store itself when it is opened, the one transaction an
// act runs in when it writes. Written once, for both: a door added and left out
// of one of them would write beside the transaction, and nothing would say so.
func (e Engine) keptBy(s *store.Store) *Engine {
	e.tasks, e.places, e.history, e.objectives, e.store = s, s, s, s, s
	return &e
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

// add records a new task carrying what it was given, and nothing else.
func (e *Engine) add(title string, c Changes) (task.Task, error) {
	t, err := task.NewTask(title)
	if err != nil {
		return task.Task{}, err
	}
	if t, err = apply(t, c, creating); err != nil {
		return task.Task{}, err
	}
	// A task added under a parent is a part from the start, so the column
	// every new task begins in is not one it may keep.
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
	return kept, nil
}

// narrowed hands back the tasks a narrowing keeps. Told nothing, it keeps
// everything, so this is List with something to say.
func (e *Engine) narrowed(kept *domain.Narrowing) ([]task.Task, error) {
	return e.tasks.Narrowed(kept)
}

// show hands back the task carrying number.
func (e *Engine) show(number int) (task.Task, error) {
	return e.tasks.Find(number)
}

// archive takes a task off the board, keeping everything it carries.
func (e *Engine) archive(number int) error { return e.puttingAway(number, true) }

// unarchive brings one back.
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
	var a task.AsTold
	found.Tell(&a)
	if away {
		a.Archived()
	} else {
		a.NotArchived()
	}
	put := a.Task()
	if err := e.tasks.Update(put); err != nil {
		return err
	}
	return e.recordChanges(found, put)
}

// sweeping is what a narrowing reaches with its parts left out. A sweep acts
// on the work, not on the pieces it is made of, and both sweeps read the same
// world — one counting parts while the other does not is a pair answering two
// different boards.
func (e *Engine) sweeping(kept *domain.Narrowing) ([]task.Task, error) {
	found, err := e.tasks.Narrowed(kept)
	if err != nil {
		return nil, err
	}
	return task.Work(found), nil
}

// archiveEvery takes off the board every task a narrowing reaches, and says
// how many. A sweep says what it did rather than refusing what it cannot take.
func (e *Engine) archiveEvery(kept *domain.Narrowing) (int, error) {
	found, err := e.sweeping(kept)
	if err != nil {
		return 0, err
	}
	swept := 0
	for _, one := range found {
		if err := e.archive(factsOf(one).number); err != nil {
			// A narrowing asked to keep both reaches what is already off the
			// board; a sweep passes over it rather than refusing.
			if errors.Is(err, task.ErrAlreadyArchived) {
				continue
			}
			return swept, err
		}
		swept++
	}
	return swept, nil
}

// labelEvery puts one label on every task a narrowing reaches, keeping what
// each already carries, and says how many.
func (e *Engine) labelEvery(kept *domain.Narrowing, label string) (int, error) {
	found, err := e.sweeping(kept)
	if err != nil {
		return 0, err
	}
	put := 0
	adding := []string{"+" + label}
	for _, one := range found {
		if _, err := e.edit(factsOf(one).number, Changes{Labels: &adding}); err != nil {
			return put, err
		}
		put++
	}
	return put, nil
}

// delete removes the task carrying number, and everything hanging under it.
func (e *Engine) delete(number int) error {
	hanging, err := e.tasks.Under(number)
	if err != nil {
		return err
	}
	// What hangs under it goes with it. A part is not work in its own right —
	// it is a line of a list its task owns, reachable only in that task's own
	// form — so a task refused while a part was there could not be deleted at
	// all. It is the rule archiving already goes by.
	for _, part := range hanging {
		if err := e.delete(factsOf(part).number); err != nil {
			return err
		}
	}
	// Which task was in focus is read before the deletion, since the store lets
	// go of it on the way and there would be nothing left to ask afterwards.
	held, err := e.places.Focus()
	if err != nil {
		return err
	}
	// What the task was, read before it goes. The trace recorded that a task
	// had been deleted and nothing of what it held — no title, no description
	// — so even putting it back by hand was impossible. Reading and undoing
	// are two different asks and cairn answers only the first; a trace that
	// kept no content made the second unanswerable by anybody.
	gone, err := e.tasks.Find(number)
	if err != nil {
		return err
	}
	if err := e.tasks.Delete(number); err != nil {
		return err
	}
	// A task being worked on that goes away closes its stretch: without this the
	// last stretch of it never ends, and the time keeps running on a task that
	// no longer exists.
	if slices.Contains(held, number) {
		if err := e.record(number, trace.FocusReleased, "", "", ""); err != nil {
			return err
		}
	}
	return e.record(number, trace.Deleted, "", whatItHeld(gone), "")
}

// under hands back what hangs directly under a task.
func (e *Engine) under(number int) ([]task.Task, error) {
	return e.tasks.Under(number)
}

// waitingOn names the tasks a task is in the way of, in the order their
// numbers run.
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
		if slices.Contains(f.blockers, number) {
			waiting = append(waiting, f.number)
		}
	}
	slices.Sort(waiting)
	return waiting, nil
}

// edit changes what was named on a task and leaves the rest alone.
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
	// A task told to hang under another stops being what it was doing on its
	// own — except finished, which stays finished. Settled before the rules
	// are asked, so that hanging a task under one never fails for the column
	// it happened to be in a moment ago.
	// Asked of the parent itself rather than of whether one was named: one
	// named as the same is no change, and one not named cannot have changed.
	if factsOf(was).parent != factsOf(t).parent {
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
	// Written down after the change and not before: an edit that was refused
	// changed nothing, and a history that recorded it would say something
	// happened that did not.
	found, err := e.tasks.Find(number)
	if err != nil {
		return task.Task{}, err
	}
	if err := e.recordChanges(was, found); err != nil {
		return task.Task{}, err
	}
	return e.freeIfNothingBlocksIt(was, found)
}

// freeIfNothingBlocksIt puts a task back in todo when the change just made took
// the last thing out of its way, and hands back the task as it ends up.
//
// A change of its own rather than part of the one asked for, and written down
// after it. Two reasons, and the second is why it is worth the second write:
// nobody asked for this column, so folding it into their change would put their
// name on cairn's decision. And a history reads top to bottom as a sentence —
// the blocker taken away, then the column that followed from it — which the one
// pass could not give, since it lays attributes out in the order a card does and
// a card shows the column first.
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
	after, err := e.tasks.Find(factsOf(found).number)
	if err != nil {
		return task.Task{}, err
	}
	if err := e.recordChanges(found, after); err != nil {
		return task.Task{}, err
	}
	return after, nil
}
