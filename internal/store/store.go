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

const taskColumns = `number, title, description, state, archived, priority, because, deadline, estimate`

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
	 WHERE sisters.state = ?)`

type boundTask struct {
	number int
	title, description,
	state, priority,
	deadline string
	labels     []string
	blockers   []int
	objectives []int
	steps      []task.Step
	linked     []int
	because    string
	estimate   int64
	archived   bool
}

func bind(t task.Task) *boundTask {
	c := t.Carried()
	return &boundTask{
		number: c.Number, title: c.Title, description: c.Description,
		state: string(c.State), priority: string(c.Priority), deadline: domain.FormatDay(c.Deadline),
		labels: c.Labels, blockers: c.Blockers, objectives: c.Objectives, steps: c.Steps, linked: c.Linked, because: c.Because,
		estimate: int64(c.Estimate), archived: c.Archived,
	}
}

func renumbered(t task.Task, number int) task.Task {
	c := t.Carried()
	c.Number = number
	return task.Made(c)
}

// Keep records a task and hands it back carrying the number it was given.
func (s *Store) Keep(t task.Task) (task.Task, error) {
	b := bind(t)
	res, err := s.on.Exec(
		`INSERT INTO tasks (title, description, state, archived, priority,
		 because, deadline, estimate, position)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, `+lastPlaceAmong+`)`,
		b.title, b.description, b.state, b.archived, b.priority,
		b.because, b.deadline, b.estimate, b.state)
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
	if err := s.keepSteps(int(number), b.steps); err != nil {
		return task.Task{}, err
	}
	if err := s.keepLinksOn(int(number), b.linked); err != nil {
		return task.Task{}, err
	}
	return renumbered(t, int(number)), nil
}

// Update replaces what the store holds for a task.
func (s *Store) Update(t task.Task) error {
	b := bind(t)
	wasIn, held, err := s.whereItSits(b.number)
	if err != nil {
		return err
	}

	place, leaving := held, wasIn != b.state
	if leaving {
		among, err := s.countSisters(b.state)
		if err != nil {
			return err
		}
		place = among + 1
	}

	if _, err := s.on.Exec(
		`UPDATE tasks SET title = ?, description = ?, state = ?,
		 archived = ?, priority = ?, because = ?, deadline = ?, estimate = ?,
		 position = ?
		 WHERE number = ?`,
		b.title, b.description, b.state, b.archived, b.priority,
		b.because, b.deadline, b.estimate, place, b.number); err != nil {
		return fmt.Errorf("changing the task: %w", err)
	}

	if leaving {
		if err := s.closeTheGapLeftIn(wasIn, held); err != nil {
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
	if err := s.keepSteps(b.number, b.steps); err != nil {
		return err
	}
	if err := s.keepLinksOn(b.number, b.linked); err != nil {
		return err
	}
	return s.keepServedBy(b.number, b.objectives)
}

// Delete removes a task, and says so when there was none to remove.
func (s *Store) Delete(number int) error {
	wasIn, held, err := s.whereItSits(number)
	if err != nil {
		return err
	}
	if _, err := s.on.Exec(`DELETE FROM tasks WHERE number = ?`, number); err != nil {
		return fmt.Errorf("deleting the task: %w", err)
	}
	if err := s.closeTheGapLeftIn(wasIn, held); err != nil {
		return err
	}
	if err := s.keepSteps(number, nil); err != nil {
		return err
	}
	if err := s.clearLabels(taskLabels, number); err != nil {
		return err
	}
	if err := s.releaseBlockersOn(number); err != nil {
		return err
	}
	if err := s.releaseLinksOn(number); err != nil {
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
		number             int
		title, description string
		state, priority    string
		because, deadline  string
		estimate           int64
		archived           bool
	)
	if err := row.Scan(&number, &title, &description, &state, &archived, &priority,
		&because, &deadline, &estimate); err != nil {
		return task.Task{}, err
	}
	day, err := domain.ParseDeadline(deadline)
	if err != nil {
		return task.Task{}, err
	}
	with, err := carried.of(number)
	if err != nil {
		return task.Task{}, err
	}

	return task.Made(task.Carried{
		Number: number, Title: title, Description: description,
		State: domain.State(state), Priority: domain.Priority(priority),
		Labels: noneIsNil(with.labels), Blockers: noneIsNil(with.blockers), Objectives: noneIsNil(with.objectives),
		Because: because, Deadline: day, Estimate: time.Duration(estimate),
		Archived: archived, Steps: noneIsNil(with.steps), Linked: noneIsNil(with.linked),
	}), nil
}

func noneIsNil[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	return s
}
