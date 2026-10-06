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

// Opening a store: where SQLite is pointed, how long a busy file is waited
// out, and the one transaction an act is kept whole in.

// Within runs one act against this store as a single transaction: what it
// wrote is kept if it worked, and none of it is if it failed.
//
// An act writes several rows — a task, then its labels, what blocks it, what
// it serves, the place it leaves and the place it takes, then what was done to
// it — and a failure between two of them used to leave the first ones written.
// Two cairns acting at once read and write the same places, and the
// transaction begins as a write (see shared) so the second waits for the
// first rather than both reading the same last place.
func (s *Store) Within(act func(*Store) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning an act on the store: %w", err)
	}
	// Undone however the act ends — an error, or a panic that would
	// otherwise leave the write lock held for as long as the process lives.
	// After a commit, there is nothing left to undo and this does nothing.
	defer func() { _ = tx.Rollback() }()
	if err := act(&Store{db: s.db, on: tx, path: s.path}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("keeping what an act wrote: %w", err)
	}
	return nil
}

// Unusable says whether an error is the store itself being out of reach
// rather than a request cairn could not make sense of — a file nothing may
// write to, most often. Answered here because this is the only package
// allowed to know what SQLite returns, and asked by the surface that has to
// choose an exit code: a script fixes its arguments on one and looks at the
// machine on the other.
func Unusable(err error) bool {
	var problem *sqlite.Error
	return errors.As(err, &problem) && problem.Code() == sqlite3.SQLITE_READONLY
}

// heldFor is how long a command waits for a store somebody else is writing to
// before giving up. SQLite refuses a busy file straight away unless it is told
// otherwise, and straight away is wrong here: the work was accepted, and
// dropping it because another command held the file for a millisecond loses it
// without a word.
//
// Five seconds is far longer than any write cairn makes and far shorter than
// somebody's patience — long enough for a queue to clear, short enough that a
// truly stuck file is still reported.
// It is a plain number of milliseconds and not a duration, so that the two
// places needing it — the pragma SQLite reads and the wait Go does — each turn
// it into their own shape. One number, said once.
const heldFor = 5000

// waitedOut is the budget in the shape Go waits in — the same wait SQLite is
// given, said once and turned into each shape where it is needed.
//
// It is a function and not a declaration for a reason the mutation gate taught:
// Go instruments statements, so arithmetic on a declaration is arithmetic no
// test can be seen exercising, and the one number this all rests on would go
// unjudged.
func waitedOut() time.Duration { return heldFor * time.Millisecond }

// shared turns a path into one more than one cairn can hold at once.
//
// The wait alone is not enough, and that is measured rather than assumed: with
// the default journal a reader blocks a writer, so six mixed operations left
// the writers waiting out all five seconds and losing five of them. In WAL a
// reader and a writer hold the file together, and the same six take fifty
// milliseconds.
//
// A transaction begins as a write — `_txlock=immediate` — rather than as a read
// that becomes one. Begun as a read, two acts each read the last place in a
// column and then both tried to write the next: SQLite refuses the second
// straight away rather than waiting, since waiting could not help it.
//
// WAL keeps recent writes in a file beside the store until it is closed, which
// is why a backup reads the store rather than copying its file — see snapshot.
func shared(path string) string {
	return addressOf(path, "_pragma=busy_timeout("+strconv.Itoa(heldFor)+")&_pragma=journal_mode(WAL)&_txlock=immediate")
}

// addressOf is how SQLite is told which file to open and how: the path, then
// what follows the question mark. The path is escaped, since it is read as an
// address — a `?` or a `#` in a folder's name started the parameters early,
// and could turn a store opened to be read into one opened to be written.
func addressOf(path, parameters string) string {
	return "file:" + (&url.URL{Path: path}).EscapedPath() + "?" + parameters
}

// Open opens a store that has to be up to date, and says so when it is not.
// Changing somebody's data as a side effect of listing their tasks is not this
// command's business — see OpenBehind for the ones whose business it is.
func Open(path string) (*Store, error) { return open(path, true) }

// OpenBehind opens a store whatever shape it is in, for the commands whose
// whole job is to change that shape.
func OpenBehind(path string) (*Store, error) { return open(path, false) }

// open opens the store at path, waiting out a store somebody else is making,
// and reporting one that is held for longer than anybody should wait.
//
// The wait is here and not left to busy_timeout because busy_timeout does not
// cover it: switching a fresh file to WAL takes a lock of its own, and that is
// measured rather than assumed — four commands making a store together left two
// of them refused, every time. Nothing was lost, the losers said so and exited,
// but this store's rule is that a command waits its turn.
//
// Nothing is slept through between two tries. The lock is held only for as long
// as the winner takes to make the file, which is a fraction of a millisecond,
// and yielding rather than sleeping means a store made by somebody else costs
// the others nothing measurable.
func open(path string, mustBeCurrent bool) (*Store, error) {
	giveUpAt := time.Now().Add(waitedOut())
	for {
		s, err := tryToOpen(path, mustBeCurrent)
		// No `err == nil` here on purpose: nothing is what nobody is holding, so
		// heldByAnother already says no to it. Asking twice would be a condition
		// no test could ever be seen telling apart from this one.
		if !heldByAnother(err) || time.Now().After(giveUpAt) {
			return s, err
		}
		runtime.Gosched()
	}
}

// heldByAnother says whether an error is somebody else holding the file, which
// is a moment to wait through, rather than anything about the store itself.
func heldByAnother(err error) bool {
	var problem *sqlite.Error
	return errors.As(err, &problem) && beingHeld(problem.Code())
}

// beingHeld reads a result code as "somebody else has the file".
//
// The primary code is what is compared, not the whole one: SQLite refines a
// result with an extended code in the bits above it, and being held comes back
// as more than one of them — measured, a plain lock reports 5 and a transaction
// that began before somebody else committed reports 517. Both are the same
// moment to wait through, so both are read as one.
//
// It takes the code and not the error because the driver's error cannot be
// built from outside its package, and a rule nothing can be seen exercising is
// a rule nothing holds.
func beingHeld(code int) bool { return code&0xff == sqlite3.SQLITE_BUSY }

// tryToOpen opens the store at path, creating the folder and the file if they
// are not there yet. A file that is not a cairn store is reported, never
// replaced.
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
		// The path goes last: a message says what went wrong first, and where
		// after, so the path is whole at the end of the line, ready to paste.
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

// Replaced says whether the file at this store's path is no longer the one it
// was opened on: a backup put back under a store held open puts another file
// in its place, and a connection still on the first one writes where nobody
// reads again. A store whose file is gone has been replaced by nothing.
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

// refuseAnotherDatabase refuses a database that holds tables and none of a
// store's: another application's, named as the store by mistake. Asked before
// the store is opened the way it is used, and asked read-only — opening it the
// usual way switches its journal and lays cairn's ledger down, which is
// writing into somebody else's file before saying it is not ours. An empty
// file is a store with nothing in it yet, and is left to be made.
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
		// What SQLite cannot read at all is refused further on, the way it
		// always was; this only answers the question it can.
		return nil
	}
	if theirs > 0 && ours == 0 {
		return fmt.Errorf("%w: it holds tables, and none of a cairn store's", ErrNotAStore)
	}
	return nil
}

// claim makes an empty file at path that its owner alone may read, and says
// os.ErrExist when one is there already. A store holds somebody's work, and
// left for SQLite to make it came out readable by anybody on the machine,
// in a folder that is not always one cairn made. An empty file is a database
// with nothing in it yet, so what opens it next lays the shape down as usual;
// the logs SQLite keeps beside it take the same permissions.
func claim(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	return f.Close()
}

// Close releases the store. One invocation opens it once and closes it once,
// so nothing is cached between two runs of the command.
func (s *Store) Close() error {
	return s.db.Close()
}
