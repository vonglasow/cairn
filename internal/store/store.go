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

// schema is created on every open. Numbers come from AUTOINCREMENT, which
// never hands out a rowid it has already used — so a number freed by a
// deletion stays freed, which is what the scenarios ask for.
const schema = `
CREATE TABLE IF NOT EXISTS tasks (
	number      INTEGER PRIMARY KEY AUTOINCREMENT,
	title       TEXT    NOT NULL,
	description TEXT    NOT NULL DEFAULT '',
	state       TEXT    NOT NULL,
	priority    TEXT    NOT NULL,
	deadline    TEXT    NOT NULL DEFAULT '',
	estimate    INTEGER NOT NULL DEFAULT 0,
	time_spent  INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS labels (
	task_number INTEGER NOT NULL REFERENCES tasks(number) ON DELETE CASCADE,
	label       TEXT    NOT NULL,
	PRIMARY KEY (task_number, label)
);`

// Store is a collection of tasks kept on disk.
type Store struct {
	db *sql.DB
}

// Open opens the store at path, creating the folder and the file if they are
// not there yet. A file that is not a cairn store is reported, never replaced.
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("making room for the store: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening the store: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("reading the store at %s: %w", path, err)
	}
	return &Store{db: db}, nil
}

// Close releases the store. One invocation opens it once and closes it once,
// so nothing is cached between two runs of the command.
func (s *Store) Close() error {
	return s.db.Close()
}

// Keep records a task and hands it back carrying the number it was given.
func (s *Store) Keep(t domain.Task) (domain.Task, error) {
	res, err := s.db.Exec(
		`INSERT INTO tasks (title, description, state, priority, deadline, estimate, time_spent)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		t.Title, t.Description, string(t.State), string(t.Priority),
		domain.FormatDay(t.Deadline), int64(t.Estimate), int64(t.TimeSpent))
	if err != nil {
		return domain.Task{}, fmt.Errorf("keeping the task: %w", err)
	}
	number, err := res.LastInsertId()
	if err != nil {
		return domain.Task{}, fmt.Errorf("numbering the task: %w", err)
	}
	t.Number = int(number)
	if err := s.keepLabels(t); err != nil {
		return domain.Task{}, err
	}
	return t, nil
}

// Update replaces what the store holds for a task. The labels given replace
// those that were there: a task carries a set, so one left out is one removed.
func (s *Store) Update(t domain.Task) error {
	res, err := s.db.Exec(
		`UPDATE tasks SET title = ?, description = ?, state = ?, priority = ?,
		 deadline = ?, estimate = ?, time_spent = ? WHERE number = ?`,
		t.Title, t.Description, string(t.State), string(t.Priority),
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
	if _, err := s.db.Exec(`DELETE FROM labels WHERE task_number = ?`, t.Number); err != nil {
		return fmt.Errorf("clearing the labels: %w", err)
	}
	return s.keepLabels(t)
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
	if _, err := s.db.Exec(`DELETE FROM labels WHERE task_number = ?`, number); err != nil {
		return fmt.Errorf("clearing the labels: %w", err)
	}
	return nil
}

// Find hands back the task carrying number, and says so when there is none.
func (s *Store) Find(number int) (domain.Task, error) {
	row := s.db.QueryRow(
		`SELECT number, title, description, state, priority, deadline, estimate, time_spent
		 FROM tasks WHERE number = ?`, number)
	t, err := scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Task{}, fmt.Errorf("task %d: %w", number, ErrNoSuchTask)
	}
	if err != nil {
		return domain.Task{}, fmt.Errorf("reading task %d: %w", number, err)
	}
	if t.Labels, err = s.labelsOf(number); err != nil {
		return domain.Task{}, err
	}
	return t, nil
}

// All hands back every task, in the order they were added.
func (s *Store) All() ([]domain.Task, error) {
	rows, err := s.db.Query(
		`SELECT number, title, description, state, priority, deadline, estimate, time_spent
		 FROM tasks ORDER BY number`)
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
		if t.Labels, err = s.labelsOf(t.Number); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading the store: %w", err)
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
	if err := row.Scan(&t.Number, &t.Title, &t.Description, &state, &priority,
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

// keepLabels records the labels a task carries.
func (s *Store) keepLabels(t domain.Task) error {
	for _, label := range t.Labels {
		if _, err := s.db.Exec(
			`INSERT INTO labels (task_number, label) VALUES (?, ?)`, t.Number, label); err != nil {
			return fmt.Errorf("keeping the labels: %w", err)
		}
	}
	return nil
}

// labelsOf hands back the labels a task carries, in alphabetical order.
func (s *Store) labelsOf(number int) ([]string, error) {
	rows, err := s.db.Query(`SELECT label FROM labels WHERE task_number = ? ORDER BY label`, number)
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
