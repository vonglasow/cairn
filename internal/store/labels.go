package store

import "fmt"

// The labels a task and an objective carry: one table each, written the same
// way.

// labelStore says where one kind of thing keeps its labels. The queries are
// written out rather than built from a table name: a name interpolated into
// SQL is a habit that reads as harmless until the day the name comes from
// somewhere else. Three literals cost nothing and cannot be turned into an
// injection by a later edit.
type labelStore struct{ insert, byOwner, clear, carrier string }

var taskLabels = labelStore{
	insert:  `INSERT INTO labels (task_number, label) VALUES (?, ?)`,
	byOwner: `SELECT label FROM labels WHERE task_number = ? ORDER BY label`,
	clear:   `DELETE FROM labels WHERE task_number = ?`,
	carrier: `EXISTS (SELECT 1 FROM labels WHERE task_number = tasks.number AND label = ?)`,
}

// keepLabels records the labels one thing carries.
func (s *Store) keepLabels(in labelStore, number int, labels []string) error {
	for _, label := range labels {
		if _, err := s.on.Exec(in.insert, number, label); err != nil {
			return fmt.Errorf("keeping the labels: %w", err)
		}
	}
	return nil
}

// clearLabels takes away every label one thing carries, so a set given can
// replace the set that was there.
func (s *Store) clearLabels(in labelStore, number int) error {
	if _, err := s.on.Exec(in.clear, number); err != nil {
		return fmt.Errorf("clearing the labels: %w", err)
	}
	return nil
}

// labelsOf hands back the labels one thing carries, in alphabetical order.
func (s *Store) labelsOf(in labelStore, number int) ([]string, error) {
	rows, err := s.on.Query(in.byOwner, number)
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
