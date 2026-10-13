// Package store keeps tasks and hands them back.
package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"cairn/internal/domain"
	"cairn/internal/task"
)

// ErrNoSuchTask is returned when a number refers to no task.
var ErrNoSuchTask = errors.New("there is no such task")

// Store is a collection of tasks kept on disk.
type Store struct {
	db     *sql.DB
	on     runs
	path   string
	opened os.FileInfo
}

const taskColumns = `number, parent, title, description, state, archived, priority, because, deadline, estimate, time_spent`

const taskOrder = `ORDER BY CASE state
	WHEN 'backlog'  THEN 0
	WHEN 'todo'     THEN 1
	WHEN 'current'  THEN 2
	WHEN 'blocked'  THEN 3
	WHEN 'watching' THEN 4
	WHEN 'paused'   THEN 5
	WHEN 'done'     THEN 6
	ELSE 7 END, position, number`

const lastPlaceAmong = `
	(SELECT COALESCE(MAX(sisters.position), 0) + 1 FROM tasks AS sisters
	 WHERE sisters.state = ? AND sisters.parent = ?)`

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

func bind(t task.Task) *boundTask {
	b := &boundTask{}
	t.Tell(b)
	return b
}

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

// Update replaces what the store holds for a task.
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
	if err := s.clearLabels(taskLabels, number); err != nil {
		return err
	}
	if err := s.releaseBlockersOn(number); err != nil {
		return err
	}
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

// All hands back every task in the one order there is: the column first, then the place set by hand inside it.
func (s *Store) All() ([]task.Task, error) {
	return s.Narrowed(&domain.Narrowing{})
}

// Under hands back the tasks hanging directly under a task, each in the place it takes among its sisters.
func (s *Store) Under(number int) ([]task.Task, error) {
	every, err := s.everyLink()
	if err != nil {
		return nil, err
	}
	tasks, err := rowsOf(s, func(r scanner) (task.Task, error) { return s.scanTask(r, every) },
		`SELECT `+taskColumns+` FROM tasks WHERE parent = ? `+taskOrder, number)
	if err != nil {
		return nil, fmt.Errorf("reading the sub-tasks of %d: %w", number, err)
	}
	return tasks, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func rowsOf[T any](s *Store, scan func(scanner) (T, error), query string, args ...any) ([]T, error) {
	rows, err := s.on.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []T{}
	for rows.Next() {
		one, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, one)
	}
	return out, rows.Err()
}

func one[T any](row scanner) (T, error) {
	var v T
	err := row.Scan(&v)
	return v, err
}

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
