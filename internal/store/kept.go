package store

import "fmt"

const (
	keptLabels      = `SELECT label FROM labels_kept ORDER BY label`
	keepLabel       = `INSERT OR IGNORE INTO labels_kept (label) VALUES (?)`
	forgetLabel     = `DELETE FROM labels_kept WHERE label = ?`
	renameKeptLabel = `UPDATE OR REPLACE labels_kept SET label = ? WHERE label = ?`
	labelColours    = `SELECT label, colour FROM label_colours`
	colourLabel     = `INSERT INTO label_colours (label, colour) VALUES (?, ?) ON CONFLICT (label) DO UPDATE SET colour = excluded.colour`
	forgetColour    = `DELETE FROM label_colours WHERE label = ?`
	renameColour    = `UPDATE OR REPLACE label_colours SET label = ? WHERE label = ?`
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

// ForgetLabel lets a kept label go, and the colour chosen for it; one never kept is let go of without a word.
func (s *Store) ForgetLabel(label string) error {
	for _, forget := range []string{forgetLabel, forgetColour} {
		if _, err := s.on.Exec(forget, label); err != nil {
			return fmt.Errorf("letting the label %q go: %w", label, err)
		}
	}
	return nil
}

// RenameKeptLabel writes another name for a kept label and for the colour chosen for it; renamed onto one already kept, the two are one.
func (s *Store) RenameKeptLabel(from, to string) error {
	for _, rename := range []string{renameKeptLabel, renameColour} {
		if _, err := s.on.Exec(rename, to, from); err != nil {
			return fmt.Errorf("renaming the label %q: %w", from, err)
		}
	}
	return nil
}

// LabelColours is the colour chosen for each label one was chosen for.
func (s *Store) LabelColours() (map[string]string, error) {
	rows, err := s.on.Query(labelColours)
	if err != nil {
		return nil, fmt.Errorf("reading the labels' colours: %w", err)
	}
	defer func() { _ = rows.Close() }()
	colours := map[string]string{}
	for rows.Next() {
		var label, colour string
		if err := rows.Scan(&label, &colour); err != nil {
			return nil, fmt.Errorf("reading a label's colour: %w", err)
		}
		colours[label] = colour
	}
	return colours, rows.Err()
}

// ColourLabel chooses a label's colour, in place of the one it had.
func (s *Store) ColourLabel(label, colour string) error {
	if _, err := s.on.Exec(colourLabel, label, colour); err != nil {
		return fmt.Errorf("colouring the label %q: %w", label, err)
	}
	return nil
}
