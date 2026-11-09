package store

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
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

const objectiveColumns = `number, title, description, state, priority, deadline, estimate, parent`

type boundObjective struct {
	number                    int
	title, description        string
	state, priority, deadline string
	labels                    []string
	estimate                  int64
	parent                    sql.NullInt64
}

func bindObjective(o objective.Objective) *boundObjective {
	c := o.Carried()
	return &boundObjective{
		number: c.Number, title: c.Title, description: c.Description, state: string(c.State),
		priority: string(c.Priority), deadline: domain.FormatDay(c.Deadline), labels: c.Labels,
		estimate: int64(c.Estimate), parent: sql.NullInt64{Int64: int64(c.Parent), Valid: c.Parent != 0},
	}
}

func renumberedObjective(o objective.Objective, number int) objective.Objective {
	c := o.Carried()
	c.Number = number
	return objective.Made(c)
}

// KeepObjective records an objective and hands it back carrying the number it was given.
func (s *Store) KeepObjective(o objective.Objective) (objective.Objective, error) {
	b := bindObjective(o)
	res, err := s.on.Exec(
		`INSERT INTO objectives (title, description, state, priority, deadline, estimate, parent)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		b.title, b.description, b.state, b.priority, b.deadline, b.estimate, b.parent)
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
		 deadline = ?, estimate = ?, parent = ? WHERE number = ?`,
		b.title, b.description, b.state, b.priority, b.deadline, b.estimate, b.parent, b.number)
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
	labelled := func(n int) ([]string, error) { return labels[n], nil }
	objectives, err := rowsOf(s, func(r scanner) (objective.Objective, error) { return s.scanObjective(r, labelled) },
		fmt.Sprintf(`SELECT %s FROM objectives%s ORDER BY number`, objectiveColumns, where), args...)
	if err != nil {
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
		parent             sql.NullInt64
	)
	if err := row.Scan(&number, &title, &description, &state, &priority,
		&deadline, &estimate, &parent); err != nil {
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

	return objective.Made(objective.Carried{
		Number: number, Title: title, Description: description, State: objective.State(state),
		Priority: domain.Priority(priority), Labels: noneIsNil(labels), Deadline: day,
		Estimate: time.Duration(estimate), Parent: int(parent.Int64),
	}), nil
}

// TasksServing hands back the tasks put directly under any of the objectives named, each once, in the order they were added.
func (s *Store) TasksServing(numbers ...int) ([]task.Task, error) {
	numbers = slices.DeleteFunc(numbers, func(n int) bool { return n == task.NoObjective })
	if len(numbers) == 0 {
		return []task.Task{}, nil
	}
	every, err := s.everyLink()
	if err != nil {
		return nil, err
	}
	args := make([]any, 0, len(numbers))
	for _, n := range numbers {
		args = append(args, n)
	}
	tasks, err := rowsOf(s, func(r scanner) (task.Task, error) { return s.scanTask(r, every) },
		fmt.Sprintf(`SELECT %s %s%s`, taskColumns, fmt.Sprintf(servingAny, placeholders(len(numbers))), taskOrder), args...)
	if err != nil {
		return nil, fmt.Errorf("reading what serves objective %v: %w", numbers, err)
	}
	return tasks, nil
}
