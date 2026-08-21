package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"cairn/internal/domain"
)

// ErrNoSuchObjective is returned when an objective is asked for by a number
// nothing carries.
var ErrNoSuchObjective = errors.New("there is no such objective")

// objectiveLabels says where an objective keeps its labels. A table of its
// own, so a task and an objective that happen to share a number never share
// what they carry.
var objectiveLabels = labelStore{
	insert:  `INSERT INTO objective_labels (objective_number, label) VALUES (?, ?)`,
	byOwner: `SELECT label FROM objective_labels WHERE objective_number = ? ORDER BY label`,
	clear:   `DELETE FROM objective_labels WHERE objective_number = ?`,
}

// objectiveColumns is the reading order every query below hands to
// scanObjective, written once so the two cannot drift apart.
const objectiveColumns = `number, title, description, state, priority, deadline, estimate`

// KeepObjective records an objective and hands it back carrying the number it
// was given. Objectives are numbered apart from tasks: they sit in a table of
// their own, so the first of each is number one.
func (s *Store) KeepObjective(o domain.Objective) (domain.Objective, error) {
	res, err := s.db.Exec(
		`INSERT INTO objectives (title, description, state, priority, deadline, estimate)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		o.Title, o.Description, string(o.State), string(o.Priority),
		domain.FormatDay(o.Deadline), int64(o.Estimate))
	if err != nil {
		return domain.Objective{}, fmt.Errorf("keeping the objective: %w", err)
	}
	number, err := res.LastInsertId()
	if err != nil {
		return domain.Objective{}, fmt.Errorf("numbering the objective: %w", err)
	}
	o.Number = int(number)
	if err := s.keepLabels(objectiveLabels, o.Number, o.Labels); err != nil {
		return domain.Objective{}, err
	}
	return o, nil
}

// UpdateObjective replaces what the store holds for an objective. The labels
// given replace those that were there: an objective carries a set, so one left
// out is one removed.
func (s *Store) UpdateObjective(o domain.Objective) error {
	res, err := s.db.Exec(
		`UPDATE objectives SET title = ?, description = ?, state = ?, priority = ?,
		 deadline = ?, estimate = ? WHERE number = ?`,
		o.Title, o.Description, string(o.State), string(o.Priority),
		domain.FormatDay(o.Deadline), int64(o.Estimate), o.Number)
	if err != nil {
		return fmt.Errorf("changing the objective: %w", err)
	}
	changed, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("changing the objective: %w", err)
	}
	if changed == 0 {
		return fmt.Errorf("objective %d: %w", o.Number, ErrNoSuchObjective)
	}
	if err := s.clearLabels(objectiveLabels, o.Number); err != nil {
		return err
	}
	return s.keepLabels(objectiveLabels, o.Number, o.Labels)
}

// DeleteObjective removes an objective, and says so when there was none to
// remove. What was put under it is the engine's business, not this one's.
func (s *Store) DeleteObjective(number int) error {
	res, err := s.db.Exec(`DELETE FROM objectives WHERE number = ?`, number)
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

// FindObjective hands back the objective carrying number, and says so when
// there is none.
func (s *Store) FindObjective(number int) (domain.Objective, error) {
	row := s.db.QueryRow(
		`SELECT `+objectiveColumns+` FROM objectives WHERE number = ?`, number)
	o, err := scanObjective(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Objective{}, fmt.Errorf("objective %d: %w", number, ErrNoSuchObjective)
	}
	if err != nil {
		return domain.Objective{}, fmt.Errorf("reading objective %d: %w", number, err)
	}
	if o.Labels, err = s.labelsOf(objectiveLabels, number); err != nil {
		return domain.Objective{}, err
	}
	return o, nil
}

// AllObjectives hands back every objective, in the order they were added.
func (s *Store) AllObjectives() ([]domain.Objective, error) {
	rows, err := s.db.Query(`SELECT ` + objectiveColumns + ` FROM objectives ORDER BY number`)
	if err != nil {
		return nil, fmt.Errorf("reading the objectives: %w", err)
	}
	defer func() { _ = rows.Close() }()

	objectives := []domain.Objective{}
	for rows.Next() {
		o, err := scanObjective(rows)
		if err != nil {
			return nil, fmt.Errorf("reading an objective: %w", err)
		}
		if o.Labels, err = s.labelsOf(objectiveLabels, o.Number); err != nil {
			return nil, err
		}
		objectives = append(objectives, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading the objectives: %w", err)
	}
	return objectives, nil
}

// scanObjective reads one objective, without its labels.
func scanObjective(row scanner) (domain.Objective, error) {
	var (
		o               domain.Objective
		state, priority string
		deadline        string
		estimate        int64
	)
	if err := row.Scan(&o.Number, &o.Title, &o.Description, &state, &priority,
		&deadline, &estimate); err != nil {
		return domain.Objective{}, err
	}
	o.State = domain.ObjectiveState(state)
	o.Priority = domain.Priority(priority)
	o.Estimate = time.Duration(estimate)

	day, err := domain.ParseDeadline(deadline)
	if err != nil {
		return domain.Objective{}, err
	}
	o.Deadline = day
	return o, nil
}

// TasksServing hands back the tasks put directly under an objective, in the
// order they were added. Sub-tasks are not among them: they serve no objective
// of their own, and are reached through the task they hang under.
func (s *Store) TasksServing(number int) ([]domain.Task, error) {
	if number == domain.NoObjective {
		return []domain.Task{}, nil
	}
	rows, err := s.db.Query(
		`SELECT `+taskColumns+` FROM tasks WHERE objective = ? ORDER BY number`, number)
	if err != nil {
		return nil, fmt.Errorf("reading what serves objective %d: %w", number, err)
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
		return nil, fmt.Errorf("reading what serves objective %d: %w", number, err)
	}
	return tasks, nil
}
