package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"cairn/internal/domain"
	"cairn/internal/objective"
	"cairn/internal/task"
)

// ErrNoSuchObjective is returned when an objective is asked for by a number nothing carries.
var ErrNoSuchObjective = errors.New("there is no such objective")

var objectiveLabels = labelStore{
	insert:  `INSERT INTO objective_labels (objective_number, label) VALUES (?, ?)`,
	byOwner: `SELECT label FROM objective_labels WHERE objective_number = ? ORDER BY label`,
	clear:   `DELETE FROM objective_labels WHERE objective_number = ?`,
	carrier: `EXISTS (SELECT 1 FROM objective_labels WHERE objective_number = objectives.number AND label = ?)`,
}

const objectiveColumns = `number, title, description, state, priority, deadline, estimate`

type boundObjective struct {
	number                    int
	title, description        string
	state, priority, deadline string
	labels                    []string
	estimate                  int64
}

func (b *boundObjective) Numbered(n int)             { b.number = n }
func (b *boundObjective) Titled(s string)            { b.title = s }
func (b *boundObjective) Standing(s objective.State) { b.state = string(s) }
func (b *boundObjective) Carries(p domain.Priority)  { b.priority = string(p) }
func (b *boundObjective) Described(s string)         { b.description = s }
func (b *boundObjective) NotDescribed()              { b.description = "" }
func (b *boundObjective) Labelled(s string)          { b.labels = append(b.labels, s) }
func (b *boundObjective) NotLabelled()               { b.labels = nil }
func (b *boundObjective) LabelsTold()                {}
func (b *boundObjective) Due(t time.Time)            { b.deadline = domain.FormatDay(t) }
func (b *boundObjective) NotDue()                    { b.deadline = "" }
func (b *boundObjective) Estimated(d time.Duration)  { b.estimate = int64(d) }
func (b *boundObjective) NotEstimated()              { b.estimate = 0 }

func bindObjective(o objective.Objective) *boundObjective {
	b := &boundObjective{}
	o.Tell(b)
	return b
}

func renumberedObjective(o objective.Objective, number int) objective.Objective {
	a := &objective.AsTold{}
	o.Tell(a)
	a.Numbered(number)
	return a.Objective()
}

// KeepObjective records an objective and hands it back carrying the number it was given.
func (s *Store) KeepObjective(o objective.Objective) (objective.Objective, error) {
	b := bindObjective(o)
	res, err := s.on.Exec(
		`INSERT INTO objectives (title, description, state, priority, deadline, estimate)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		b.title, b.description, b.state, b.priority, b.deadline, b.estimate)
	if err != nil {
		return objective.Objective{}, fmt.Errorf("keeping the objective: %w", err)
	}
	number, err := res.LastInsertId()
	if err != nil {
		return objective.Objective{}, fmt.Errorf("numbering the objective: %w", err)
	}
	if err := s.keepLabels(objectiveLabels, int(number), b.labels); err != nil {
		return objective.Objective{}, err
	}
	return renumberedObjective(o, int(number)), nil
}

// UpdateObjective replaces what the store holds for an objective.
func (s *Store) UpdateObjective(o objective.Objective) error {
	b := bindObjective(o)
	res, err := s.on.Exec(
		`UPDATE objectives SET title = ?, description = ?, state = ?, priority = ?,
		 deadline = ?, estimate = ? WHERE number = ?`,
		b.title, b.description, b.state, b.priority, b.deadline, b.estimate, b.number)
	if err != nil {
		return fmt.Errorf("changing the objective: %w", err)
	}
	changed, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("changing the objective: %w", err)
	}
	if changed == 0 {
		return fmt.Errorf("objective %d: %w", b.number, ErrNoSuchObjective)
	}
	if err := s.clearLabels(objectiveLabels, b.number); err != nil {
		return err
	}
	return s.keepLabels(objectiveLabels, b.number, b.labels)
}

// DeleteObjective removes an objective, and says so when there was none to remove.
func (s *Store) DeleteObjective(number int) error {
	res, err := s.on.Exec(`DELETE FROM objectives WHERE number = ?`, number)
	if err != nil {
		return fmt.Errorf("deleting the objective: %w", err)
	}
	changed, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("deleting the objective: %w", err)
	}
	if changed == 0 {
		return fmt.Errorf("objective %d: %w", number, ErrNoSuchObjective)
	}
	return s.clearLabels(objectiveLabels, number)
}

// FindObjective hands back the objective carrying number, and says so when there is none.
func (s *Store) FindObjective(number int) (objective.Objective, error) {
	row := s.on.QueryRow(
		`SELECT `+objectiveColumns+` FROM objectives WHERE number = ?`, number)
	o, err := s.scanObjective(row, func(n int) ([]string, error) { return s.labelsOf(objectiveLabels, n) })
	if errors.Is(err, sql.ErrNoRows) {
		return objective.Objective{}, fmt.Errorf("objective %d: %w", number, ErrNoSuchObjective)
	}
	if err != nil {
		return objective.Objective{}, fmt.Errorf("reading objective %d: %w", number, err)
	}
	return o, nil
}

// AllObjectives hands back every objective, in the order they were added.
func (s *Store) AllObjectives() ([]objective.Objective, error) {
	return s.NarrowedObjectives(&domain.Narrowing{})
}

// NarrowedObjectives hands back the objectives a narrowing keeps, in the order they were added.
func (s *Store) NarrowedObjectives(kept *domain.Narrowing) ([]objective.Objective, error) {
	labels, err := pairs[string](s, `SELECT objective_number, label FROM objective_labels ORDER BY objective_number, label`)
	if err != nil {
		return nil, fmt.Errorf("reading the labels: %w", err)
	}
	where, args := readBy(kept).where(objectiveLabels)
	rows, err := s.on.Query(
		`SELECT `+objectiveColumns+` FROM objectives`+where+` ORDER BY number`, args...)
	if err != nil {
		return nil, fmt.Errorf("reading the objectives: %w", err)
	}
	defer func() { _ = rows.Close() }()

	objectives := []objective.Objective{}
	for rows.Next() {
		o, err := s.scanObjective(rows, func(n int) ([]string, error) { return labels[n], nil })
		if err != nil {
			return nil, fmt.Errorf("reading an objective: %w", err)
		}
		objectives = append(objectives, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading the objectives: %w", err)
	}
	return objectives, nil
}

func (s *Store) scanObjective(row scanner, labelled func(int) ([]string, error)) (objective.Objective, error) {
	var (
		number             int
		title, description string
		state, priority    string
		deadline           string
		estimate           int64
	)
	if err := row.Scan(&number, &title, &description, &state, &priority,
		&deadline, &estimate); err != nil {
		return objective.Objective{}, err
	}
	day, err := domain.ParseDeadline(deadline)
	if err != nil {
		return objective.Objective{}, err
	}
	labels, err := labelled(number)
	if err != nil {
		return objective.Objective{}, err
	}

	a := &objective.AsTold{}
	a.Numbered(number)
	a.Titled(title)
	a.Standing(objective.State(state))
	a.Carries(domain.Priority(priority))
	domain.TellOrNot(description, a.Described, a.NotDescribed)
	domain.TellAll(labels, a.Labelled, a.NotLabelled, a.LabelsTold)
	domain.TellOrNot(day, a.Due, a.NotDue)
	domain.TellOrNot(time.Duration(estimate), a.Estimated, a.NotEstimated)
	return a.Objective(), nil
}

// TasksServing hands back the tasks put directly under an objective, in the order they were added.
func (s *Store) TasksServing(number int) ([]task.Task, error) {
	if number == task.NoObjective {
		return []task.Task{}, nil
	}
	every, err := s.everyLink()
	if err != nil {
		return nil, err
	}
	rows, err := s.on.Query(`SELECT `+taskColumns+` `+servingOne+taskOrder, number)
	if err != nil {
		return nil, fmt.Errorf("reading what serves objective %d: %w", number, err)
	}
	defer func() { _ = rows.Close() }()

	tasks := []task.Task{}
	for rows.Next() {
		t, err := s.scanTask(rows, every)
		if err != nil {
			return nil, fmt.Errorf("reading a task: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading what serves objective %d: %w", number, err)
	}
	return tasks, nil
}
