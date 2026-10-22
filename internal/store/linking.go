package store

import "fmt"

const (
	linksOf    = `SELECT high FROM links WHERE low = ? UNION SELECT low FROM links WHERE high = ? ORDER BY 1`
	clearLinks = `DELETE FROM links WHERE low = ? OR high = ?`
	keepLink   = `INSERT OR IGNORE INTO links (low, high) VALUES (?, ?)`
)

func (s *Store) linkedTo(number int) ([]int, error) {
	others, err := rowsOf(s, one[int], linksOf, number, number)
	if err != nil {
		return nil, fmt.Errorf("reading what task %d is linked to: %w", number, err)
	}
	return others, nil
}

func (s *Store) keepLinksOn(number int, others []int) error {
	if _, err := s.on.Exec(clearLinks, number, number); err != nil {
		return fmt.Errorf("clearing the links of task %d: %w", number, err)
	}
	for _, other := range others {
		if _, err := s.on.Exec(keepLink, min(number, other), max(number, other)); err != nil {
			return fmt.Errorf("keeping the links of task %d: %w", number, err)
		}
	}
	return nil
}

func (s *Store) releaseLinksOn(number int) error {
	if _, err := s.on.Exec(clearLinks, number, number); err != nil {
		return fmt.Errorf("releasing the links of task %d: %w", number, err)
	}
	return nil
}
