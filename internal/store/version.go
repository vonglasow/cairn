package store

import "fmt"

// Version is a number SQLite raises whenever a connection other than this one commits.
func (s *Store) Version() (int64, error) {
	var version int64
	if err := s.on.QueryRow(`PRAGMA data_version`).Scan(&version); err != nil {
		return 0, fmt.Errorf("reading whether the store moved: %w", err)
	}
	return version, nil
}
