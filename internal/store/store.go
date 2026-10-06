// Package store keeps tasks and hands them back. It is the only place that
// knows a database exists: it speaks in entities, so no *sql.Rows crosses its
// boundary and no caller needs to know the schema.
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	// The pure-Go SQLite driver. Registered for its side effect, as database/sql
	// drivers are. Pure Go on purpose: cgo would forfeit cross-compilation for
	// the surfaces that come later.

	"cairn/internal/domain"
	"cairn/internal/task"
)

// ErrNoSuchTask is returned when a number refers to no task.
var ErrNoSuchTask = errors.New("there is no such task")

// Store is a collection of tasks kept on disk.
type Store struct {
	db *sql.DB
	// on is what every read and write goes through: the database itself, or
	// the one transaction an act is being kept whole in.
	on   runs
	path string
	// opened is the file this store was opened on, to tell it apart from one
	// put in its place since.
	opened os.FileInfo
}

// taskColumns is the reading order every query below hands to scanTask,
// written once so Find, All and Under cannot drift apart. A column added to
// one of them and forgotten in the others comes back as a zero nobody notices.
const taskColumns = `number, parent, title, description, state, archived, priority, because, deadline, estimate, time_spent`

// taskOrder is how every listing reads a task's place: the column first, in the
// order the columns are read, then the place set by hand inside it.
//
// The number breaks a tie, so a store whose tasks predate the places still
// comes back in one settled order rather than whichever the engine happened to
// return. A column this cairn does not know sorts last: an unknown column is
// not the next thing to do.
//
// Written out rather than built from domain.States, and that is a trade made
// with both gates in view. Built, it could not be a constant — and a query
// assembled from a function call cannot be told from one assembled from user
// input, which is a warning worth keeping rather than silencing. Constant, the
// two can drift, so a test compares this to what domain.States says: derived by
// a test instead of by construction, which fails loudly rather than quietly.
const taskOrder = `ORDER BY CASE state
	WHEN 'backlog'  THEN 0
	WHEN 'todo'     THEN 1
	WHEN 'current'  THEN 2
	WHEN 'blocked'  THEN 3
	WHEN 'watching' THEN 4
	WHEN 'paused'   THEN 5
	WHEN 'done'     THEN 6
	ELSE 7 END, position, number`

// lastPlaceAmong is the place a task takes when it arrives in a column: one
// past the last of the tasks it will sit with. Its sisters are the tasks in
// the same column under the same parent, which is what makes a sub-task rank
// among its sisters and a root task among the root tasks of its column.
const lastPlaceAmong = `
	(SELECT COALESCE(MAX(sisters.position), 0) + 1 FROM tasks AS sisters
	 WHERE sisters.state = ? AND sisters.parent = ?)`

// boundTask holds a task's facts in the shape SQL will bind: strings and
// integers, nothing a driver has to guess at. A task tells them here rather
// than handing over its fields directly, which is what lets Keep and Update
// stay correct the day those fields stop being exported.
type boundTask struct {
	number, parent int
	title, description,
	state, priority,
	deadline string
	labels          []string
	blockers        []int
	objectives      []int
	because         string
	estimate, spent int64
	archived        bool
}

func (b *boundTask) Numbered(n int)            { b.number = n }
func (b *boundTask) Titled(s string)           { b.title = s }
func (b *boundTask) InColumn(s domain.State)   { b.state = string(s) }
func (b *boundTask) Archived()                 { b.archived = true }
func (b *boundTask) NotArchived()              { b.archived = false }
func (b *boundTask) Carries(p domain.Priority) { b.priority = string(p) }
func (b *boundTask) Described(s string)        { b.description = s }
func (b *boundTask) NotDescribed()             { b.description = "" }
func (b *boundTask) Labelled(s string)         { b.labels = append(b.labels, s) }
func (b *boundTask) NotLabelled()              { b.labels = nil }
func (b *boundTask) LabelsTold()               {}
func (b *boundTask) BlockedBy(n int)           { b.blockers = append(b.blockers, n) }
func (b *boundTask) NotBlocked()               { b.blockers = nil }
func (b *boundTask) BlockersTold()             {}
func (b *boundTask) Because(s string)          { b.because = s }
func (b *boundTask) NotBecause()               { b.because = "" }
func (b *boundTask) Due(t time.Time)           { b.deadline = domain.FormatDay(t) }
func (b *boundTask) NotDue()                   { b.deadline = "" }
func (b *boundTask) Estimated(d time.Duration) { b.estimate = int64(d) }
func (b *boundTask) NotEstimated()             { b.estimate = 0 }
func (b *boundTask) Spent(d time.Duration)     { b.spent = int64(d) }
func (b *boundTask) NothingSpent()             { b.spent = 0 }
func (b *boundTask) Under(n int)               { b.parent = n }
func (b *boundTask) UnderNothing()             { b.parent = task.NoParent }
func (b *boundTask) Serves(n int)              { b.objectives = append(b.objectives, n) }
func (b *boundTask) ServesNothing()            { b.objectives = nil }
func (b *boundTask) ObjectivesTold()           {}

// bind tells a task's facts to a fresh boundTask, ready for the parameters an
// INSERT or an UPDATE needs.
func bind(t task.Task) *boundTask {
	b := &boundTask{}
	t.Tell(b)
	return b
}

// renumbered hands back a task carrying a new number and everything else t
// already told. Told a second time here rather than assigned: the number is
// not a field this package may reach into once it stops being exported.
func renumbered(t task.Task, number int) task.Task {
	a := &task.AsTold{}
	t.Tell(a)
	a.Numbered(number)
	return a.Task()
}

// Keep records a task and hands it back carrying the number it was given.
func (s *Store) Keep(t task.Task) (task.Task, error) {
	b := bind(t)
	res, err := s.on.Exec(
		`INSERT INTO tasks (parent, title, description, state, archived, priority,
		 because, deadline, estimate, time_spent, position)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, `+lastPlaceAmong+`)`,
		b.parent, b.title, b.description, b.state, b.archived, b.priority,
		b.because, b.deadline, b.estimate, b.spent, b.state, b.parent)
	if err != nil {
		return task.Task{}, fmt.Errorf("keeping the task: %w", err)
	}
	number, err := res.LastInsertId()
	if err != nil {
		return task.Task{}, fmt.Errorf("numbering the task: %w", err)
	}
	if err := s.keepLabels(taskLabels, int(number), b.labels); err != nil {
		return task.Task{}, err
	}
	if err := s.keepBlockersOn(int(number), b.blockers); err != nil {
		return task.Task{}, err
	}
	if err := s.keepServedBy(int(number), b.objectives); err != nil {
		return task.Task{}, err
	}
	return renumbered(t, int(number)), nil
}

// Update replaces what the store holds for a task. The labels given replace
// those that were there: a task carries a set, so one left out is one removed.
//
// A task changes place only when it leaves the tasks it sat with — a different
// column, or a different parent. Editing a title or a priority moves nothing.
func (s *Store) Update(t task.Task) error {
	b := bind(t)
	wasIn, wasUnder, held, err := s.whereItSits(b.number)
	if err != nil {
		return err
	}

	place, leaving := held, wasIn != b.state || wasUnder != b.parent
	if leaving {
		among, err := s.countSisters(b.state, b.parent)
		if err != nil {
			return err
		}
		// The task still carries its old group at this point, so it is not
		// among the ones just counted: one past them is the last place there.
		place = among + 1
	}

	if _, err := s.on.Exec(
		`UPDATE tasks SET parent = ?, title = ?, description = ?, state = ?,
		 archived = ?, priority = ?, because = ?, deadline = ?, estimate = ?,
		 time_spent = ?, position = ?
		 WHERE number = ?`,
		b.parent, b.title, b.description, b.state, b.archived, b.priority,
		b.because, b.deadline, b.estimate, b.spent, place, b.number); err != nil {
		return fmt.Errorf("changing the task: %w", err)
	}

	// The group it left closes up behind it. Without this the places there stop
	// running 1..N — the order still reads correctly, so nothing looks wrong,
	// and every departure leaves one more hole than the last.
	if leaving {
		if err := s.closeTheGapLeftIn(wasIn, wasUnder, held); err != nil {
			return err
		}
	}

	if err := s.clearLabels(taskLabels, b.number); err != nil {
		return err
	}
	if err := s.keepLabels(taskLabels, b.number, b.labels); err != nil {
		return err
	}
	if err := s.keepBlockersOn(b.number, b.blockers); err != nil {
		return err
	}
	return s.keepServedBy(b.number, b.objectives)
}

// Delete removes a task, and says so when there was none to remove.
func (s *Store) Delete(number int) error {
	// Where it sat is read before it goes, since the group it leaves has to
	// close up behind it: a task removed from the middle would otherwise leave
	// the places there running 1, 3, 4.
	wasIn, wasUnder, held, err := s.whereItSits(number)
	if err != nil {
		return err
	}
	if _, err := s.on.Exec(`DELETE FROM tasks WHERE number = ?`, number); err != nil {
		return fmt.Errorf("deleting the task: %w", err)
	}
	if err := s.closeTheGapLeftIn(wasIn, wasUnder, held); err != nil {
		return err
	}
	// What referred to the task goes with it, and it goes by hand. The schema
	// declares ON DELETE CASCADE and ON DELETE SET NULL, and neither fires:
	// SQLite leaves foreign keys off unless a connection turns them on, and
	// this one does not. Measured, after believing otherwise — the labels
	// disappearing on a delete looked like the cascade working and was this
	// call all along.
	if err := s.clearLabels(taskLabels, number); err != nil {
		return err
	}
	// Both sides of every blocking link, since a task that is gone neither
	// blocks anything nor is blocked by anything.
	if err := s.releaseBlockersOn(number); err != nil {
		return err
	}
	// What it served goes with it too, for the same reason.
	if _, err := s.on.Exec(clearServing, number); err != nil {
		return fmt.Errorf("releasing what task %d served: %w", number, err)
	}
	return s.releaseFocusOn(number)
}

// Find hands back the task carrying number, and says so when there is none.
func (s *Store) Find(number int) (task.Task, error) {
	row := s.on.QueryRow(
		`SELECT `+taskColumns+` FROM tasks WHERE number = ?`, number)
	t, err := s.scanTask(row, askedEach{s})
	if errors.Is(err, sql.ErrNoRows) {
		return task.Task{}, fmt.Errorf("task %d: %w", number, ErrNoSuchTask)
	}
	if err != nil {
		return task.Task{}, fmt.Errorf("reading task %d: %w", number, err)
	}
	return t, nil
}

// All hands back every task in the one order there is: the column first,
// then the place set by hand inside it.
//
// A read told to keep nothing is a read of everything, so this is Narrowed
// with nothing named rather than a second loop over the same rows.
func (s *Store) All() ([]task.Task, error) {
	return s.Narrowed(&domain.Narrowing{})
}

// Under hands back the tasks hanging directly under a task, each in the place
// it takes among its sisters.
func (s *Store) Under(number int) ([]task.Task, error) {
	every, err := s.everyLink()
	if err != nil {
		return nil, err
	}
	rows, err := s.on.Query(
		`SELECT `+taskColumns+` FROM tasks WHERE parent = ? `+taskOrder, number)
	if err != nil {
		return nil, fmt.Errorf("reading the sub-tasks of %d: %w", number, err)
	}
	defer func() { _ = rows.Close() }()

	tasks := []task.Task{}
	for rows.Next() {
		t, err := s.scanTask(rows, every)
		if err != nil {
			return nil, fmt.Errorf("reading a sub-task: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading the sub-tasks of %d: %w", number, err)
	}
	return tasks, nil
}

// scanner is what both a single row and a row of a result set can do. Declared
// here so Find and All share one reading of a task.
type scanner interface {
	Scan(dest ...any) error
}

// scanTask reads one task, labels included, by scanning a row into plain
// variables and telling them to a builder rather than into a task's fields
// directly — the membrane Tell uses, the other way round, so this keeps
// working the day those fields stop being exported.
func (s *Store) scanTask(row scanner, carried links) (task.Task, error) {
	var (
		number, parent     int
		title, description string
		state, priority    string
		because, deadline  string
		estimate, spent    int64
		archived           bool
	)
	if err := row.Scan(&number, &parent, &title, &description, &state, &archived, &priority,
		&because, &deadline, &estimate, &spent); err != nil {
		return task.Task{}, err
	}
	day, err := domain.ParseDeadline(deadline)
	if err != nil {
		return task.Task{}, err
	}
	labels, blockers, objectives, err := carried.of(number)
	if err != nil {
		return task.Task{}, err
	}

	a := &task.AsTold{}
	a.Numbered(number)
	a.Titled(title)
	a.InColumn(domain.State(state))
	if archived {
		a.Archived()
	} else {
		a.NotArchived()
	}
	a.Carries(domain.Priority(priority))
	domain.TellOrNot(description, a.Described, a.NotDescribed)
	domain.TellAll(labels, a.Labelled, a.NotLabelled, a.LabelsTold)
	domain.TellAll(blockers, a.BlockedBy, a.NotBlocked, a.BlockersTold)
	domain.TellOrNot(because, a.Because, a.NotBecause)
	domain.TellOrNot(day, a.Due, a.NotDue)
	domain.TellOrNot(time.Duration(estimate), a.Estimated, a.NotEstimated)
	domain.TellOrNot(time.Duration(spent), a.Spent, a.NothingSpent)
	if parent == task.NoParent {
		a.UnderNothing()
	} else {
		a.Under(parent)
	}
	domain.TellAll(objectives, a.Serves, a.ServesNothing, a.ObjectivesTold)
	return a.Task(), nil
}
