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
	"time"

	// The pure-Go SQLite driver. Registered for its side effect, as database/sql
	// drivers are. Pure Go on purpose: cgo would forfeit cross-compilation for
	// the surfaces that come later.
	_ "modernc.org/sqlite"

	"cairn/internal/domain"
)

// ErrNoSuchTask is returned when a number refers to no task.
var ErrNoSuchTask = errors.New("there is no such task")

// Store is a collection of tasks kept on disk.
type Store struct {
	db   *sql.DB
	path string
}

// Open opens a store that has to be up to date, and says so when it is not.
// Changing somebody's data as a side effect of listing their tasks is not this
// command's business — see OpenBehind for the ones whose business it is.
func Open(path string) (*Store, error) { return open(path, true) }

// OpenBehind opens a store whatever shape it is in, for the commands whose
// whole job is to change that shape.
func OpenBehind(path string) (*Store, error) { return open(path, false) }

// open opens the store at path, creating the folder and the file if they are
// not there yet. A file that is not a cairn store is reported, never replaced.
func open(path string, mustBeCurrent bool) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("making room for the store: %w", err)
	}
	db, err := sql.Open("sqlite", path)
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

// Keep records a task and hands it back carrying the number it was given.
func (s *Store) Keep(t domain.Task) (domain.Task, error) {
	res, err := s.db.Exec(
		`INSERT INTO tasks (parent, objective, title, description, state, priority, deadline, estimate, time_spent)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.Parent, t.Objective, t.Title, t.Description, string(t.State), string(t.Priority),
		domain.FormatDay(t.Deadline), int64(t.Estimate), int64(t.TimeSpent))
	if err != nil {
		return domain.Task{}, fmt.Errorf("keeping the task: %w", err)
	}
	number, err := res.LastInsertId()
	if err != nil {
		return domain.Task{}, fmt.Errorf("numbering the task: %w", err)
	}
	t.Number = int(number)
	if err := s.keepLabels(taskLabels, t.Number, t.Labels); err != nil {
		return domain.Task{}, err
	}
	return t, nil
}

// Update replaces what the store holds for a task. The labels given replace
// those that were there: a task carries a set, so one left out is one removed.
func (s *Store) Update(t domain.Task) error {
	res, err := s.db.Exec(
		`UPDATE tasks SET parent = ?, objective = ?, title = ?, description = ?, state = ?,
		 priority = ?, deadline = ?, estimate = ?, time_spent = ? WHERE number = ?`,
		t.Parent, t.Objective, t.Title, t.Description, string(t.State), string(t.Priority),
		domain.FormatDay(t.Deadline), int64(t.Estimate), int64(t.TimeSpent), t.Number)
	if err != nil {
		return fmt.Errorf("changing the task: %w", err)
	}
	changed, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("changing the task: %w", err)
	}
	if changed == 0 {
		return fmt.Errorf("task %d: %w", t.Number, ErrNoSuchTask)
	}
	if err := s.clearLabels(taskLabels, t.Number); err != nil {
		return err
	}
	return s.keepLabels(taskLabels, t.Number, t.Labels)
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
func (s *Store) Find(number int) (domain.Task, error) {
	row := s.db.QueryRow(
		`SELECT `+taskColumns+` FROM tasks WHERE number = ?`, number)
	t, err := scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Task{}, fmt.Errorf("task %d: %w", number, ErrNoSuchTask)
	}
	if err != nil {
		return domain.Task{}, fmt.Errorf("reading task %d: %w", number, err)
	}
	if t.Labels, err = s.labelsOf(taskLabels, number); err != nil {
		return domain.Task{}, err
	}
	return t, nil
}

// All hands back every task, in the order they were added.
func (s *Store) All() ([]domain.Task, error) {
	rows, err := s.db.Query(
		`SELECT ` + taskColumns + ` FROM tasks ORDER BY number`)
	if err != nil {
		return nil, fmt.Errorf("reading the store: %w", err)
	}
	defer func() { _ = rows.Close() }()

	tasks := []domain.Task{}
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("reading a task: %w", err)
		}
		if t.Labels, err = s.labelsOf(taskLabels, t.Number); err != nil {
			return nil, err
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
func (s *Store) Under(number int) ([]domain.Task, error) {
	rows, err := s.db.Query(
		`SELECT `+taskColumns+` FROM tasks WHERE parent = ? ORDER BY number`, number)
	if err != nil {
		return nil, fmt.Errorf("reading the sub-tasks of %d: %w", number, err)
	}
	defer func() { _ = rows.Close() }()

	tasks := []domain.Task{}
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("reading a sub-task: %w", err)
		}
		if t.Labels, err = s.labelsOf(taskLabels, t.Number); err != nil {
			return nil, err
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

// scanTask reads one task, without its labels.
func scanTask(row scanner) (domain.Task, error) {
	var (
		t               domain.Task
		state, priority string
		deadline        string
		estimate, spent int64
	)
	if err := row.Scan(&t.Number, &t.Parent, &t.Objective, &t.Title, &t.Description, &state, &priority,
		&deadline, &estimate, &spent); err != nil {
		return domain.Task{}, err
	}
	t.State = domain.State(state)
	t.Priority = domain.Priority(priority)
	t.Estimate = time.Duration(estimate)
	t.TimeSpent = time.Duration(spent)

	day, err := domain.ParseDeadline(deadline)
	if err != nil {
		return domain.Task{}, err
	}
	t.Deadline = day
	return t, nil
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
