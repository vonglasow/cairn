package store

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// Within runs one act against this store as a single transaction: what it wrote is kept if it worked, and none of it is if it failed.
func (s *Store) Within(act func(*Store) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning an act on the store: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := act(&Store{db: s.db, on: tx, path: s.path}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("keeping what an act wrote: %w", err)
	}
	return nil
}

// Unusable says whether an error is the store itself being out of reach rather than a request cairn could not make sense of — a file nothing may write to, most often.
func Unusable(err error) bool {
	var problem *sqlite.Error
	return errors.As(err, &problem) && problem.Code() == sqlite3.SQLITE_READONLY
}

const heldFor = 5000

func waitedOut() time.Duration { return heldFor * time.Millisecond }

func shared(path string) string {
	return addressOf(path, "_pragma=busy_timeout("+strconv.Itoa(heldFor)+")&_pragma=journal_mode(WAL)&_txlock=immediate")
}

func addressOf(path, parameters string) string {
	return "file:" + (&url.URL{Path: path}).EscapedPath() + "?" + parameters
}

// Open opens a store that has to be up to date, and says so when it is not.
func Open(path string) (*Store, error) { return open(path, true) }

// OpenBehind opens a store whatever shape it is in, for the commands whose whole job is to change that shape.
func OpenBehind(path string) (*Store, error) { return open(path, false) }

func open(path string, mustBeCurrent bool) (*Store, error) {
	giveUpAt := time.Now().Add(waitedOut())
	for {
		s, err := tryToOpen(path, mustBeCurrent)
		if !heldByAnother(err) || time.Now().After(giveUpAt) {
			return s, err
		}
		runtime.Gosched()
	}
}

func heldByAnother(err error) bool {
	var problem *sqlite.Error
	return errors.As(err, &problem) && beingHeld(problem.Code())
}

func beingHeld(code int) bool { return code&0xff == sqlite3.SQLITE_BUSY }

func tryToOpen(path string, mustBeCurrent bool) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("making room for the store: %w", err)
	}
	if err := claim(path); err != nil && !errors.Is(err, os.ErrExist) {
		return nil, fmt.Errorf("making the store: %w", err)
	}
	if err := refuseAnotherDatabase(path); err != nil {
		return nil, fmt.Errorf("%w — store: %s", err, path)
	}
	db, err := sql.Open("sqlite", shared(path))
	if err != nil {
		return nil, fmt.Errorf("opening the store: %w", err)
	}
	s := &Store{db: db, on: db, path: path}
	if err := s.settle(mustBeCurrent); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("%w — store: %s", err, path)
	}
	opened, err := os.Stat(path)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("reading the store just opened: %w — store: %s", err, path)
	}
	s.opened = opened
	return s, nil
}

// Replaced says whether the file at this store's path is no longer the one it was opened on: a backup put back under a store held open puts another file in its place, and a connection still on the first one writes where nobody reads again.
func (s *Store) Replaced() (bool, error) {
	now, err := os.Stat(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("looking at the store: %w", err)
	}
	return !os.SameFile(s.opened, now), nil
}

func refuseAnotherDatabase(path string) error {
	db, err := readOnly(path)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	var theirs, ours int
	err = db.QueryRow(`SELECT
		count(*) FILTER (WHERE name NOT LIKE 'sqlite_%'),
		count(*) FILTER (WHERE name IN ('tasks', 'migrations'))
		FROM sqlite_master WHERE type = 'table'`).Scan(&theirs, &ours)
	if err != nil {
		return nil
	}
	if theirs > 0 && ours == 0 {
		return fmt.Errorf("%w: it holds tables, and none of a cairn store's", ErrNotAStore)
	}
	return nil
}

func claim(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	return f.Close()
}

// Close releases the store.
func (s *Store) Close() error {
	return s.db.Close()
}
