// Package store keeps tasks and hands them back. It is the only place that
// knows a database exists: it speaks in entities, so no *sql.Rows crosses its
// boundary and no caller needs to know the schema.
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	// The pure-Go SQLite driver. Registered for its side effect, as database/sql
	// drivers are. Pure Go on purpose: cgo would forfeit cross-compilation for
	// the surfaces that come later.
	sqlite "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

	"cairn/internal/domain"
	"cairn/internal/task"
)

// ErrNoSuchTask is returned when a number refers to no task.
var ErrNoSuchTask = errors.New("there is no such task")

// Store is a collection of tasks kept on disk.
type Store struct {
	db   *sql.DB
	path string
}

// Unusable says whether an error is the store itself being out of reach
// rather than a request cairn could not make sense of — a file nothing may
// write to, most often. Answered here because this is the only package
// allowed to know what SQLite returns, and asked by the surface that has to
// choose an exit code: a script fixes its arguments on one and looks at the
// machine on the other.
func Unusable(err error) bool {
	var problem *sqlite.Error
	return errors.As(err, &problem) && problem.Code() == sqlite3.SQLITE_READONLY
}

// heldFor is how long a command waits for a store somebody else is writing to
// before giving up. SQLite refuses a busy file straight away unless it is told
// otherwise, and straight away is wrong here: the work was accepted, and
// dropping it because another command held the file for a millisecond loses it
// without a word.
//
// Five seconds is far longer than any write cairn makes and far shorter than
// somebody's patience — long enough for a queue to clear, short enough that a
// truly stuck file is still reported.
// It is a plain number of milliseconds and not a duration, so that the two
// places needing it — the pragma SQLite reads and the wait Go does — each turn
// it into their own shape. One number, said once.
const heldFor = 5000

// waitedOut is the budget in the shape Go waits in — the same wait SQLite is
// given, said once and turned into each shape where it is needed.
//
// It is a function and not a declaration for a reason the mutation gate taught:
// Go instruments statements, so arithmetic on a declaration is arithmetic no
// test can be seen exercising, and the one number this all rests on would go
// unjudged.
func waitedOut() time.Duration { return heldFor * time.Millisecond }

// shared turns a path into one more than one cairn can hold at once.
//
// The wait alone is not enough, and that is measured rather than assumed: with
// the default journal a reader blocks a writer, so six mixed operations left
// the writers waiting out all five seconds and losing five of them. In WAL a
// reader and a writer hold the file together, and the same six take fifty
// milliseconds.
//
// WAL keeps recent writes in a file beside the store. Nothing is lost by it
// here because a command opens the store in its RunE and closes it on the way
// out, and closing checkpoints — so between two commands there is only ever
// the one file, which is what `backup add` copies. A surface that holds the
// store open, as a server would, has to checkpoint before a backup is taken.
func shared(path string) string {
	return "file:" + path + "?_pragma=busy_timeout(" + strconv.Itoa(heldFor) + ")&_pragma=journal_mode(WAL)"
}

// Open opens a store that has to be up to date, and says so when it is not.
// Changing somebody's data as a side effect of listing their tasks is not this
// command's business — see OpenBehind for the ones whose business it is.
func Open(path string) (*Store, error) { return open(path, true) }

// OpenBehind opens a store whatever shape it is in, for the commands whose
// whole job is to change that shape.
func OpenBehind(path string) (*Store, error) { return open(path, false) }

// open opens the store at path, waiting out a store somebody else is making,
// and reporting one that is held for longer than anybody should wait.
//
// The wait is here and not left to busy_timeout because busy_timeout does not
// cover it: switching a fresh file to WAL takes a lock of its own, and that is
// measured rather than assumed — four commands making a store together left two
// of them refused, every time. Nothing was lost, the losers said so and exited,
// but this store's rule is that a command waits its turn.
//
// Nothing is slept through between two tries. The lock is held only for as long
// as the winner takes to make the file, which is a fraction of a millisecond,
// and yielding rather than sleeping means a store made by somebody else costs
// the others nothing measurable.
func open(path string, mustBeCurrent bool) (*Store, error) {
	giveUpAt := time.Now().Add(waitedOut())
	for {
		s, err := tryToOpen(path, mustBeCurrent)
		// No `err == nil` here on purpose: nothing is what nobody is holding, so
		// heldByAnother already says no to it. Asking twice would be a condition
		// no test could ever be seen telling apart from this one.
		if !heldByAnother(err) || time.Now().After(giveUpAt) {
			return s, err
		}
		runtime.Gosched()
	}
}

// heldByAnother says whether an error is somebody else holding the file, which
// is a moment to wait through, rather than anything about the store itself.
func heldByAnother(err error) bool {
	var problem *sqlite.Error
	return errors.As(err, &problem) && beingHeld(problem.Code())
}

// beingHeld reads a result code as "somebody else has the file".
//
// The primary code is what is compared, not the whole one: SQLite refines a
// result with an extended code in the bits above it, and being held comes back
// as more than one of them — measured, a plain lock reports 5 and a transaction
// that began before somebody else committed reports 517. Both are the same
// moment to wait through, so both are read as one.
//
// It takes the code and not the error because the driver's error cannot be
// built from outside its package, and a rule nothing can be seen exercising is
// a rule nothing holds.
func beingHeld(code int) bool { return code&0xff == sqlite3.SQLITE_BUSY }

// tryToOpen opens the store at path, creating the folder and the file if they
// are not there yet. A file that is not a cairn store is reported, never
// replaced.
func tryToOpen(path string, mustBeCurrent bool) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("making room for the store: %w", err)
	}
	db, err := sql.Open("sqlite", shared(path))
	if err != nil {
		return nil, fmt.Errorf("opening the store: %w", err)
	}
	s := &Store{db: db, path: path}
	if err := s.settle(mustBeCurrent); err != nil {
		_ = db.Close()
		// The path goes last, never first: fang capitalises the opening letter of
		// a message, and a path that came back as /Var/Folders/… is a path
		// nobody can paste back into a shell.
		return nil, fmt.Errorf("%w — store: %s", err, path)
	}
	return s, nil
}

// Close releases the store. One invocation opens it once and closes it once,
// so nothing is cached between two runs of the command.
func (s *Store) Close() error {
	return s.db.Close()
}

// taskColumns is the reading order every query below hands to scanTask,
// written once so Find, All and Under cannot drift apart. A column added to
// one of them and forgotten in the others comes back as a zero nobody notices.
const taskColumns = `number, parent, objective, title, description, state, priority, deadline, estimate, time_spent`

// boundTask holds a task's facts in the shape SQL will bind: strings and
// integers, nothing a driver has to guess at. A task tells them here rather
// than handing over its fields directly, which is what lets Keep and Update
// stay correct the day those fields stop being exported.
type boundTask struct {
	number, parent, objective int
	title, description,
	state, priority,
	deadline string
	labels          []string
	estimate, spent int64
}

func (b *boundTask) Numbered(n int)            { b.number = n }
func (b *boundTask) Titled(s string)           { b.title = s }
func (b *boundTask) InColumn(s domain.State)   { b.state = string(s) }
func (b *boundTask) Carries(p domain.Priority) { b.priority = string(p) }
func (b *boundTask) Described(s string)        { b.description = s }
func (b *boundTask) NotDescribed()             { b.description = "" }
func (b *boundTask) Labelled(s string)         { b.labels = append(b.labels, s) }
func (b *boundTask) NotLabelled()              { b.labels = nil }
func (b *boundTask) LabelsTold()               {}
func (b *boundTask) Due(t time.Time)           { b.deadline = domain.FormatDay(t) }
func (b *boundTask) NotDue()                   { b.deadline = "" }
func (b *boundTask) Estimated(d time.Duration) { b.estimate = int64(d) }
func (b *boundTask) NotEstimated()             { b.estimate = 0 }
func (b *boundTask) Spent(d time.Duration)     { b.spent = int64(d) }
func (b *boundTask) NothingSpent()             { b.spent = 0 }
func (b *boundTask) Under(n int)               { b.parent = n }
func (b *boundTask) UnderNothing()             { b.parent = task.NoParent }
func (b *boundTask) Serves(n int)              { b.objective = n }
func (b *boundTask) ServesNothing()            { b.objective = task.NoObjective }

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
	res, err := s.db.Exec(
		`INSERT INTO tasks (parent, objective, title, description, state, priority, deadline, estimate, time_spent)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		b.parent, b.objective, b.title, b.description, b.state, b.priority,
		b.deadline, b.estimate, b.spent)
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
	return renumbered(t, int(number)), nil
}

// Update replaces what the store holds for a task. The labels given replace
// those that were there: a task carries a set, so one left out is one removed.
func (s *Store) Update(t task.Task) error {
	b := bind(t)
	res, err := s.db.Exec(
		`UPDATE tasks SET parent = ?, objective = ?, title = ?, description = ?, state = ?,
		 priority = ?, deadline = ?, estimate = ?, time_spent = ? WHERE number = ?`,
		b.parent, b.objective, b.title, b.description, b.state, b.priority,
		b.deadline, b.estimate, b.spent, b.number)
	if err != nil {
		return fmt.Errorf("changing the task: %w", err)
	}
	changed, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("changing the task: %w", err)
	}
	if changed == 0 {
		return fmt.Errorf("task %d: %w", b.number, ErrNoSuchTask)
	}
	if err := s.clearLabels(taskLabels, b.number); err != nil {
		return err
	}
	return s.keepLabels(taskLabels, b.number, b.labels)
}

// Delete removes a task, and says so when there was none to remove.
func (s *Store) Delete(number int) error {
	res, err := s.db.Exec(`DELETE FROM tasks WHERE number = ?`, number)
	if err != nil {
		return fmt.Errorf("deleting the task: %w", err)
	}
	changed, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("deleting the task: %w", err)
	}
	if changed == 0 {
		return fmt.Errorf("task %d: %w", number, ErrNoSuchTask)
	}
	return s.clearLabels(taskLabels, number)
}

// Find hands back the task carrying number, and says so when there is none.
func (s *Store) Find(number int) (task.Task, error) {
	row := s.db.QueryRow(
		`SELECT `+taskColumns+` FROM tasks WHERE number = ?`, number)
	t, err := s.scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return task.Task{}, fmt.Errorf("task %d: %w", number, ErrNoSuchTask)
	}
	if err != nil {
		return task.Task{}, fmt.Errorf("reading task %d: %w", number, err)
	}
	return t, nil
}

// All hands back every task, in the order they were added.
func (s *Store) All() ([]task.Task, error) {
	rows, err := s.db.Query(
		`SELECT ` + taskColumns + ` FROM tasks ORDER BY number`)
	if err != nil {
		return nil, fmt.Errorf("reading the store: %w", err)
	}
	defer func() { _ = rows.Close() }()

	tasks := []task.Task{}
	for rows.Next() {
		t, err := s.scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("reading a task: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading the store: %w", err)
	}
	return tasks, nil
}

// Under hands back the tasks hanging directly under a task, in the order they
// were added.
func (s *Store) Under(number int) ([]task.Task, error) {
	rows, err := s.db.Query(
		`SELECT `+taskColumns+` FROM tasks WHERE parent = ? ORDER BY number`, number)
	if err != nil {
		return nil, fmt.Errorf("reading the sub-tasks of %d: %w", number, err)
	}
	defer func() { _ = rows.Close() }()

	tasks := []task.Task{}
	for rows.Next() {
		t, err := s.scanTask(rows)
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
func (s *Store) scanTask(row scanner) (task.Task, error) {
	var (
		number, parent, objective int
		title, description        string
		state, priority           string
		deadline                  string
		estimate, spent           int64
	)
	if err := row.Scan(&number, &parent, &objective, &title, &description, &state, &priority,
		&deadline, &estimate, &spent); err != nil {
		return task.Task{}, err
	}
	day, err := domain.ParseDeadline(deadline)
	if err != nil {
		return task.Task{}, err
	}
	labels, err := s.labelsOf(taskLabels, number)
	if err != nil {
		return task.Task{}, err
	}

	a := &task.AsTold{}
	a.Numbered(number)
	a.Titled(title)
	a.InColumn(domain.State(state))
	a.Carries(domain.Priority(priority))
	if description == "" {
		a.NotDescribed()
	} else {
		a.Described(description)
	}
	if len(labels) == 0 {
		a.NotLabelled()
	} else {
		for _, label := range labels {
			a.Labelled(label)
		}
	}
	a.LabelsTold()
	if day.IsZero() {
		a.NotDue()
	} else {
		a.Due(day)
	}
	if estimate == 0 {
		a.NotEstimated()
	} else {
		a.Estimated(time.Duration(estimate))
	}
	if spent == 0 {
		a.NothingSpent()
	} else {
		a.Spent(time.Duration(spent))
	}
	if parent == task.NoParent {
		a.UnderNothing()
	} else {
		a.Under(parent)
	}
	if objective == task.NoObjective {
		a.ServesNothing()
	} else {
		a.Serves(objective)
	}
	return a.Task(), nil
}

// labelStore says where one kind of thing keeps its labels. The queries are
// written out rather than built from a table name: a name interpolated into
// SQL is a habit that reads as harmless until the day the name comes from
// somewhere else. Three literals cost nothing and cannot be turned into an
// injection by a later edit.
type labelStore struct{ insert, byOwner, clear string }

var taskLabels = labelStore{
	insert:  `INSERT INTO labels (task_number, label) VALUES (?, ?)`,
	byOwner: `SELECT label FROM labels WHERE task_number = ? ORDER BY label`,
	clear:   `DELETE FROM labels WHERE task_number = ?`,
}

// keepLabels records the labels one thing carries.
func (s *Store) keepLabels(in labelStore, number int, labels []string) error {
	for _, label := range labels {
		if _, err := s.db.Exec(in.insert, number, label); err != nil {
			return fmt.Errorf("keeping the labels: %w", err)
		}
	}
	return nil
}

// clearLabels takes away every label one thing carries, so a set given can
// replace the set that was there.
func (s *Store) clearLabels(in labelStore, number int) error {
	if _, err := s.db.Exec(in.clear, number); err != nil {
		return fmt.Errorf("clearing the labels: %w", err)
	}
	return nil
}

// labelsOf hands back the labels one thing carries, in alphabetical order.
func (s *Store) labelsOf(in labelStore, number int) ([]string, error) {
	rows, err := s.db.Query(in.byOwner, number)
	if err != nil {
		return nil, fmt.Errorf("reading the labels: %w", err)
	}
	defer func() { _ = rows.Close() }()

	labels := []string{}
	for rows.Next() {
		var label string
		if err := rows.Scan(&label); err != nil {
			return nil, fmt.Errorf("reading a label: %w", err)
		}
		labels = append(labels, label)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading the labels: %w", err)
	}
	return labels, nil
}
