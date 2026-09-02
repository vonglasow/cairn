package store

// Whether anything wrote to the store since the last look. A board stays open
// while a command line and an agent go on working, and it has to notice — but
// reading every task twice a second to find out that nothing happened is a
// price nobody should pay for it.

import "fmt"

// Version is a number SQLite raises whenever a connection other than this one
// commits. Reading it costs one integer, which is what makes it usable on a
// beat: the store is only read in full once the number has actually moved.
//
// It says nothing about what changed, and it is not a count of changes — two
// writes may raise it by more than one. Equal means nothing happened; different
// means something did, and the only honest answer is to look again.
func (s *Store) Version() (int64, error) {
	var version int64
	// PRAGMA data_version is not a setting: it is only ever read, and it is
	// local to this connection, which is exactly why it can be compared with
	// what this connection saw before.
	if err := s.db.QueryRow(`PRAGMA data_version`).Scan(&version); err != nil {
		return 0, fmt.Errorf("reading whether the store moved: %w", err)
	}
	return version, nil
}
