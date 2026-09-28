// Package engine holds the operations. A surface parses input, calls one thing
// here, and prints what comes back — it decides nothing. The engine is also the
// only caller of the store, whatever the store is made of.
package engine

import (
	"errors"
	"fmt"
	"slices"
	"time"

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

// The reasons a task cannot hang where it was asked to, or be deleted.
// Re-exported from domain, which is where a task's own rules live: a surface
// asks the engine and may not import domain, so the name it recognises has to
// live here too.
var (
	ErrParentIsSelf    = task.ErrParentIsSelf
	ErrParentIsSubTask = task.ErrParentIsSubTask
	ErrCarriesSubTasks = task.ErrCarriesSubTasks
)

// tasks is what the engine needs from whatever keeps its tasks. Declared here,
// by the consumer, so the rules stay testable and a second backing can arrive
// without the engine noticing.
type tasks interface {
	Keep(task.Task) (task.Task, error)
	Update(task.Task) error
	Delete(int) error
	Find(int) (task.Task, error)
	All() ([]task.Task, error)
	Narrowed(*store.Narrowing) ([]task.Task, error)
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
	// path is where the store was opened. Kept because an operation is handed
	// what sits beside the store as well as the store itself, and a surface
	// that stays open runs operations long after whoever opened it has gone.
	path string
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
	Blocker     *[]string
	Because     *string
	Deadline    *string
	Estimate    *string
	Spent       *string
	Parent      *int
	Objective   *[]string
}

// named says whether anything at all was given to change.
func (c Changes) named() bool {
	return c.Title != nil || c.Description != nil || c.State != nil ||
		c.Priority != nil || c.Labels != nil || c.Deadline != nil ||
		c.Estimate != nil || c.Spent != nil || c.Parent != nil || c.Objective != nil ||
		c.Blocker != nil || c.Because != nil
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
	return &Engine{tasks: kept, places: kept, history: kept, objectives: kept, store: kept, path: path}, nil
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
func (e *Engine) Add(title string, c Changes) (task.Task, error) {
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
	if err := e.canHang(t); err != nil {
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

// List hands back every task, in the order they were added.
func (e *Engine) List() ([]task.Task, error) {
	return e.tasks.All()
}

// Narrowed hands back the tasks a narrowing keeps. Told nothing, it keeps
// everything, so this is List with something to say.
func (e *Engine) Narrowed(kept *store.Narrowing) ([]task.Task, error) {
	return e.tasks.Narrowed(kept)
}

// Show hands back the task carrying number.
func (e *Engine) Show(number int) (task.Task, error) {
	return e.tasks.Find(number)
}

// Archive takes a task off the board, keeping everything it carries.
func (e *Engine) Archive(number int) error { return e.puttingAway(number, true) }

// Unarchive brings one back.
func (e *Engine) Unarchive(number int) error { return e.puttingAway(number, false) }

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
	return e.tasks.Update(a.Task())
}

// ArchiveEvery takes off the board every task a narrowing reaches, and says
// how many. A sweep says what it did rather than refusing what it cannot take.
func (e *Engine) ArchiveEvery(kept *store.Narrowing) (int, error) {
	found, err := e.tasks.Narrowed(kept)
	if err != nil {
		return 0, err
	}
	swept := 0
	for _, one := range found {
		if err := e.Archive(factsOf(one).number); err != nil {
			// A sweep passes over what it cannot take: one already off the
			// board, and a part that goes with its parent.
			if errors.Is(err, task.ErrAlreadyArchived) || errors.Is(err, task.ErrPartGoesWithItsWork) {
				continue
			}
			return swept, err
		}
		swept++
	}
	return swept, nil
}

// LabelEvery puts one label on every task a narrowing reaches, keeping what
// each already carries, and says how many.
func (e *Engine) LabelEvery(kept *store.Narrowing, label string) (int, error) {
	found, err := e.tasks.Narrowed(kept)
	if err != nil {
		return 0, err
	}
	put := 0
	adding := []string{"+" + label}
	for _, one := range found {
		if _, err := e.Edit(factsOf(one).number, Changes{Labels: &adding}); err != nil {
			return put, err
		}
		put++
	}
	return put, nil
}

// Delete removes the task carrying number, refusing while something still
// hangs under it: what depends on the task would go with it, and a deletion
// that takes more than the number typed is a deletion nobody asked for.
func (e *Engine) Delete(number int) error {
	hanging, err := e.tasks.Under(number)
	if err != nil {
		return err
	}
	if err := task.MayDelete(number, hanging); err != nil {
		return err
	}
	// Which task was in focus is read before the deletion, since the store lets
	// go of it on the way and there would be nothing left to ask afterwards.
	held, err := e.places.Focus()
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
	return e.record(number, trace.Deleted, "", "", "")
}

// Under hands back what hangs directly under a task.
func (e *Engine) Under(number int) ([]task.Task, error) {
	return e.tasks.Under(number)
}

// canHang says whether a task may hang where it was told to. A parent has to
// exist, cannot be the task itself, and cannot already hang under something —
// what hangs under a task is never something with more hanging under it.
func (e *Engine) canHang(t task.Task) error {
	parent := factsOf(t).parent
	if parent == task.NoParent {
		return nil
	}
	if err := t.MayCarryItsColumn(); err != nil {
		return err
	}
	if err := t.MayServeAnObjective(); err != nil {
		return err
	}
	if err := t.MayHangUnderItsParent(); err != nil {
		return err
	}
	found, err := e.tasks.Find(parent)
	if err != nil {
		return err
	}
	return found.MayCarrySubTasks()
}

// Edit changes what was named on a task and leaves the rest alone.
func (e *Engine) Edit(number int, c Changes) (task.Task, error) {
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
	if c.Parent != nil && factsOf(was).parent != factsOf(t).parent {
		t = t.Settled()
	}
	if err := e.canHang(t); err != nil {
		return task.Task{}, err
	}
	if err := e.canServe(t); err != nil {
		return task.Task{}, err
	}
	if err := e.canBeBlocked(t); err != nil {
		return task.Task{}, err
	}
	if err := mayLeave(factsOf(was), factsOf(t)); err != nil {
		return task.Task{}, err
	}
	if err := e.canFinish(was, t); err != nil {
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
	freed, put := freedFrom(factsOf(was), factsOf(found))
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

// taskFacts is a task's membrane for what the engine, or a test on it, has to
// read back off it — everything else is a rule the task already owns, or a
// fact only a builder writing a new task needs told to it.
type taskFacts struct {
	number, parent      int
	title, description  string
	state               domain.State
	priority            domain.Priority
	labels              []string
	blockers            []int
	objectives          []int
	because             string
	deadline            time.Time
	estimate, timeSpent time.Duration
	archived            bool
}

func (f *taskFacts) Numbered(n int)            { f.number = n }
func (f *taskFacts) Titled(s string)           { f.title = s }
func (f *taskFacts) InColumn(s domain.State)   { f.state = s }
func (f *taskFacts) Archived()                 { f.archived = true }
func (f *taskFacts) NotArchived()              { f.archived = false }
func (f *taskFacts) Carries(p domain.Priority) { f.priority = p }
func (f *taskFacts) Described(s string)        { f.description = s }
func (f *taskFacts) NotDescribed()             { f.description = "" }
func (f *taskFacts) Labelled(l string)         { f.labels = append(f.labels, l) }
func (f *taskFacts) NotLabelled()              { f.labels = nil }
func (f *taskFacts) LabelsTold()               {}
func (f *taskFacts) BlockedBy(n int)           { f.blockers = append(f.blockers, n) }
func (f *taskFacts) NotBlocked()               { f.blockers = nil }
func (f *taskFacts) BlockersTold()             {}
func (f *taskFacts) Because(s string)          { f.because = s }
func (f *taskFacts) NotBecause()               { f.because = "" }
func (f *taskFacts) Due(t time.Time)           { f.deadline = t }
func (f *taskFacts) NotDue()                   { f.deadline = time.Time{} }
func (f *taskFacts) Estimated(d time.Duration) { f.estimate = d }
func (f *taskFacts) NotEstimated()             { f.estimate = 0 }
func (f *taskFacts) Spent(d time.Duration)     { f.timeSpent = d }
func (f *taskFacts) NothingSpent()             { f.timeSpent = 0 }
func (f *taskFacts) Under(n int)               { f.parent = n }
func (f *taskFacts) UnderNothing()             { f.parent = task.NoParent }
func (f *taskFacts) Serves(n int)              { f.objectives = append(f.objectives, n) }
func (f *taskFacts) ServesNothing()            { f.objectives = nil }
func (f *taskFacts) ObjectivesTold()           {}

// factsOf reads back what the engine needs from a task, without reaching
// into fields that stopped being exported.
func factsOf(t task.Task) taskFacts {
	var f taskFacts
	t.Tell(&f)
	return f
}

// apply lays the changes over a task, refusing any value it cannot make sense
// of. A field that was not named is left exactly as it was.
func apply(t task.Task, c Changes, atCreation bool) (task.Task, error) {
	current := factsOf(t)
	a := &task.AsTold{}
	t.Tell(a)

	if c.Title != nil {
		retitled, err := task.NewTask(*c.Title)
		if err != nil {
			return task.Task{}, err
		}
		a.Titled(factsOf(retitled).title)
	}
	if c.Description != nil {
		a.Described(*c.Description)
	}
	if c.State != nil {
		state, err := domain.ParseState(*c.State)
		if err != nil {
			return task.Task{}, err
		}
		a.InColumn(state)
	}
	if c.Priority != nil {
		priority, err := domain.ParsePriority(*c.Priority)
		if err != nil {
			return task.Task{}, err
		}
		a.Carries(priority)
	}
	if c.Labels != nil {
		labels, err := changedLabels(*c.Labels, current.labels, atCreation)
		if err != nil {
			return task.Task{}, err
		}
		a.NotLabelled()
		for _, label := range labels {
			a.Labelled(label)
		}
		a.LabelsTold()
	}
	if c.Blocker != nil {
		blockers, err := changedBlockers(*c.Blocker, current.blockers, atCreation)
		if err != nil {
			return task.Task{}, err
		}
		a.NotBlocked()
		for _, blocker := range blockers {
			a.BlockedBy(blocker)
		}
		a.BlockersTold()
	}
	if c.Because != nil {
		a.Because(*c.Because)
	}
	if c.Deadline != nil {
		deadline, err := domain.ParseDeadline(*c.Deadline)
		if err != nil {
			return task.Task{}, err
		}
		a.Due(deadline)
	}
	if c.Spent != nil {
		spent, err := domain.ParseSpent(*c.Spent)
		if err != nil {
			return task.Task{}, err
		}
		a.Spent(spent.Apply(current.timeSpent))
	}
	if c.Parent != nil {
		a.Under(*c.Parent)
	}
	if c.Objective != nil {
		served, err := changedNumbers(*c.Objective, current.objectives, atCreation, ErrObjectiveRequired,
			ErrObjectivesBothWays, ErrObjectiveSignAtBirth)
		if err != nil {
			return task.Task{}, err
		}
		a.ServesNothing()
		for _, one := range served {
			a.Serves(one)
		}
		a.ObjectivesTold()
	}
	if c.Estimate != nil {
		estimate, err := domain.ParseEstimate(*c.Estimate)
		if err != nil {
			return task.Task{}, err
		}
		a.Estimated(estimate)
	}
	return a.Task(), nil
}

// canFinish refuses a task moved to done while something hanging under it is
// not. Asked only when done is where it lands and is not where it was: a task
// already finished, edited for something else, has nothing to answer for, and
// asking anyway would make it fail for a reason nobody caused today.
func (e *Engine) canFinish(was, now task.Task) error {
	if factsOf(now).state != domain.Done || factsOf(was).state == domain.Done {
		return nil
	}
	number := factsOf(now).number
	hanging, err := e.tasks.Under(number)
	if err != nil {
		return err
	}
	return task.MayFinish(number, hanging)
}
