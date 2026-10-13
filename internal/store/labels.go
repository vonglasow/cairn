package store

import "fmt"

type labelStore struct{ insert, byOwner, clear, carrier string }

var taskLabels = labelStore{
	insert:  `INSERT INTO labels (task_number, label) VALUES (?, ?)`,
	byOwner: `SELECT label FROM labels WHERE task_number = ? ORDER BY label`,
	clear:   `DELETE FROM labels WHERE task_number = ?`,
	carrier: `EXISTS (SELECT 1 FROM labels WHERE task_number = tasks.number AND label = ?)`,
}

func (s *Store) keepLabels(in labelStore, number int, labels []string) error {
	for _, label := range labels {
		if _, err := s.on.Exec(in.insert, number, label); err != nil {
			return fmt.Errorf("keeping the labels: %w", err)
		}
	}
	return nil
}

func (s *Store) clearLabels(in labelStore, number int) error {
	if _, err := s.on.Exec(in.clear, number); err != nil {
		return fmt.Errorf("clearing the labels: %w", err)
	}
	return nil
}

func (s *Store) labelsOf(in labelStore, number int) ([]string, error) {
	labels, err := rowsOf(s, one[string], in.byOwner, number)
	if err != nil {
		return nil, fmt.Errorf("reading the labels: %w", err)
	}
	return labels, nil
}
