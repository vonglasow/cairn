package store

import "fmt"

const (
	keptLabels      = `SELECT label FROM labels_kept ORDER BY label`
	keepLabel       = `INSERT OR IGNORE INTO labels_kept (label) VALUES (?)`
	forgetLabel     = `DELETE FROM labels_kept WHERE label = ?`
	renameKeptLabel = `UPDATE OR REPLACE labels_kept SET label = ? WHERE label = ?`
)

// KeptLabels is every label made on its own, whatever carries it or not, in alphabetical order.
func (s *Store) KeptLabels() ([]string, error) {
	kept, err := rowsOf(s, one[string], keptLabels)
	if err != nil {
		return nil, fmt.Errorf("reading the labels kept: %w", err)
	}
	return kept, nil
}

// KeepLabel keeps a label though nothing carries it; kept twice, it is kept once.
func (s *Store) KeepLabel(label string) error {
	if _, err := s.on.Exec(keepLabel, label); err != nil {
		return fmt.Errorf("keeping the label %q: %w", label, err)
	}
	return nil
}

// ForgetLabel lets a kept label go; one never kept is let go of without a word.
func (s *Store) ForgetLabel(label string) error {
	if _, err := s.on.Exec(forgetLabel, label); err != nil {
		return fmt.Errorf("letting the label %q go: %w", label, err)
	}
	return nil
}

// RenameKeptLabel writes another name for a kept label; renamed onto one already kept, the two are one.
func (s *Store) RenameKeptLabel(from, to string) error {
	if _, err := s.on.Exec(renameKeptLabel, to, from); err != nil {
		return fmt.Errorf("renaming the label %q: %w", from, err)
	}
	return nil
}
