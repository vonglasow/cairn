package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"cairn/internal/domain"
	"cairn/internal/migration"
)

// What can be wrong with the shape of a store.
var (
	// ErrNewerStore: an older binary cannot know what a later one wrote, and
	// writing into it anyway is how a store gets quietly wrecked.
	ErrNewerStore = errors.New("the store comes from a newer cairn")

	// ErrStoreBehind: the store is missing a change this cairn needs. Applying
	// it on the way past would change somebody's data as a side effect of
	// listing their tasks, so it is said and left alone.
	ErrStoreBehind = errors.New("the store is behind this cairn")

	// ErrNoSuchChange: a number that names no change.
	ErrNoSuchChange = errors.New("there is no such change")

	// ErrNothingWaiting: everything this cairn knows is already applied. That is
	// an answer, not a refusal — `migration apply` on a store already up to date
	// has done what was asked of it.
	ErrNothingWaiting = errors.New("there is nothing waiting to be applied")

	// ErrAlreadyApplied: a change named by number that the store already has.
	// Kept apart from ErrNothingWaiting because it is a refusal: somebody asked
	// for a particular change and did not get it, and answering "nothing is
	// waiting" would be false the moment another one is.
	ErrAlreadyApplied = errors.New("the change is already applied")

	// ErrNothingToRollBack: only the shape a store starts with is left, and
	// that one is not a change on top of something else — undoing it would mean
	// throwing the store away, which is what deleting the file is for.
	ErrNothingToRollBack = errors.New("there is nothing left to roll back")

	// ErrUndoneMeanwhile: the change read as the last applied was undone by
	// another cairn between the reading and the rollback.
	ErrUndoneMeanwhile = errors.New("the change was rolled back by somebody else meanwhile")

	// ErrEarlierWaiting: a change was asked for while one before it is still
	// waiting. Changes are made in order — each was written on the shape the
	// ones before it leave — and one applied out of order gives a shape no
	// sequence of changes describes.
	ErrEarlierWaiting = errors.New("an earlier change is still waiting")
)

// ledger is where the store records what has been done to it.
const ledger = `
CREATE TABLE IF NOT EXISTS migrations (
	name       TEXT NOT NULL PRIMARY KEY,
	applied_at TEXT NOT NULL
);`

// changeFacts pulls a change's raw content out through Tell, in the shape
// this file needs it: the SQL to run, the name to record, and whether the
// ledger already has it. Nothing here reaches into a change's fields, and a
// change knows nothing of the store that reads it this way.
type changeFacts struct {
	number                    int
	name, does, undoes, loses string
	applied                   bool
	made                      time.Time
}

func (f *changeFacts) Numbered(n int)  { f.number = n }
func (f *changeFacts) Named(s string)  { f.name = s }
func (f *changeFacts) Does(s string)   { f.does = s }
func (f *changeFacts) Undoes(s string) { f.undoes = s }
func (f *changeFacts) NotUndoable()    { f.undoes = "" }
func (f *changeFacts) Loses(s string)  { f.loses = s }
func (f *changeFacts) NothingLost()    { f.loses = "" }
func (f *changeFacts) Applied(t time.Time) {
	f.applied = true
	f.made = t
}
func (f *changeFacts) NotApplied() { f.applied = false }

// factsOfChange reads back what this file needs from a change, without
// reaching into fields that stopped being exported.
func factsOfChange(m migration.Migration) changeFacts {
	var f changeFacts
	m.Tell(&f)
	return f
}

// settle prepares a freshly opened store, in the order a store needs it:
// first the ledger, without which nothing can be recorded, then whichever of
// the remaining steps its shape calls for.
//
// Opening a store is meant to read its shape, never lay it down — see "the
// shape of the store changes, and never on its own" in
// docs/architecture.md — and three of the steps below are the stated
// exceptions to that. Each is named for what it does, and says right above
// itself why it writes during what everywhere else is a mere opening.
func (s *Store) settle(mustBeCurrent bool) error {
	if err := s.makeTheLedgerWhenOpeningFindsNone(); err != nil {
		return err
	}
	empty, err := s.isEmpty()
	if err != nil {
		return err
	}
	if empty {
		// Whether it was this opening that made the store or somebody else's,
		// what is on disk afterwards is a store like any other — so it is read
		// like any other, and there is nothing to return early for.
		if err := s.layDownTheShapeWhenOpeningFindsNothing(); err != nil {
			return err
		}
	}
	if err := s.adoptTheUnrecordedShapeWhileOpening(); err != nil {
		return err
	}
	if err := s.refuseAnUnknownChange(); err != nil {
		return err
	}
	if !mustBeCurrent {
		return nil
	}
	return s.refuseIfBehind()
}

// makeTheLedgerWhenOpeningFindsNone writes during a mere opening: without a
// ledger nothing that follows can record what has been done to the store,
// and a store opened for the first time has no other moment to be given one.
func (s *Store) makeTheLedgerWhenOpeningFindsNone() error {
	var kept int
	if err := s.on.QueryRow(
		`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'migrations'`,
	).Scan(&kept); err != nil {
		return fmt.Errorf("reading what was done to the store: %w", err)
	}
	if kept == 0 {
		if _, err := s.on.Exec(ledger); err != nil {
			return fmt.Errorf("making room to record what is done to the store: %w", err)
		}
	}
	return nil
}

// isEmpty says whether the file holds nothing of ours at all.
func (s *Store) isEmpty() (bool, error) { return isEmptyOn(s.on) }

func isEmptyOn(on runs) (bool, error) {
	var count int
	if err := on.QueryRow(
		`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name IN ('tasks', 'labels')`).
		Scan(&count); err != nil {
		return false, fmt.Errorf("looking at the shape of the store: %w", err)
	}
	if count > 0 {
		return false, nil
	}
	applied, err := appliedNamesOn(on)
	if err != nil {
		return false, err
	}
	return len(applied) == 0, nil
}

// layDownTheShapeWhenOpeningFindsNothing writes, during a mere opening, and
// that is said out loud in its name because it is the exception: making a store
// is not changing one. A file that held nothing a moment ago has nothing to
// lose and nobody to warn, so it is created at the shape this cairn knows rather
// than left for `migration apply` to catch up on later.
//
// It is one act, not one per change. Measured: laying the shape down change by
// change let another command open in the middle, read a ledger with one of the
// three recorded, and report the store as behind — which it was not, it was
// being made. A store half made is a store nobody may see.
func (s *Store) layDownTheShapeWhenOpeningFindsNothing() error {
	making, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("making the store: %w", err)
	}
	defer func() { _ = making.Rollback() }()

	// Looked at again, from inside. Emptiness was read before the transaction
	// began, and between those two moments somebody else may have made the whole
	// store — measured: two commands both found it empty, both applied the first
	// change, and the second was refused by the ledger's own uniqueness. Whoever
	// arrives second is not making anything; it reads what is there like any
	// other opening.
	empty, err := isEmptyOn(making)
	if err != nil {
		return err
	}
	if !empty {
		return nil
	}

	for _, one := range migration.Known {
		if err := s.makeOn(making, one); err != nil {
			return err
		}
	}
	if err := making.Commit(); err != nil {
		return fmt.Errorf("making the store: %w", err)
	}
	return nil
}

// adoptTheUnrecordedShapeWhileOpening writes during a mere opening: a store
// written before cairn kept a ledger holds the first shape with no record of
// it, and reading that silence as "nothing applied yet" would have the next
// step try to lay a shape down that is already there.
func (s *Store) adoptTheUnrecordedShapeWhileOpening() error {
	applied, err := s.appliedNames()
	if err != nil {
		return err
	}
	if len(applied) > 0 {
		return nil
	}
	return s.record(factsOfChange(migration.Known[0]).name)
}

// refuseAnUnknownChange stops on a store carrying a change this cairn does
// not know. It only reads the store; it changes nothing.
func (s *Store) refuseAnUnknownChange() error {
	applied, err := s.appliedNames()
	if err != nil {
		return err
	}
	known := make(map[string]bool, len(migration.Known))
	for _, raw := range migration.Known {
		known[factsOfChange(raw).name] = true
	}
	for name := range applied {
		if !known[name] {
			return fmt.Errorf("%w: it carries %d change(s) including %q, this cairn knows %d",
				ErrNewerStore, len(applied), name, len(migration.Known))
		}
	}
	return nil
}

// refuseIfBehind says whether the store is missing a change this cairn
// knows. It only reads the store; it changes nothing.
func (s *Store) refuseIfBehind() error {
	waiting, err := s.Waiting()
	if err != nil {
		return err
	}
	if len(waiting) > 0 {
		return fmt.Errorf("%w by %d change(s)", ErrStoreBehind, len(waiting))
	}
	return nil
}

// appliedNames reads what the store says has been done to it.
func (s *Store) appliedNames() (map[string]time.Time, error) { return appliedNamesOn(s.on) }

func appliedNamesOn(on runs) (map[string]time.Time, error) {
	rows, err := on.Query(`SELECT name, applied_at FROM migrations`)
	if err != nil {
		return nil, fmt.Errorf("reading what was done to the store: %w", err)
	}
	defer func() { _ = rows.Close() }()

	applied := map[string]time.Time{}
	for rows.Next() {
		var name, at string
		if err := rows.Scan(&name, &at); err != nil {
			return nil, fmt.Errorf("reading what was done to the store: %w", err)
		}
		made, err := time.Parse(time.RFC3339, at)
		if err != nil {
			return nil, fmt.Errorf("reading when %q was applied: %q is not a moment: %w", name, at, err)
		}
		applied[name] = made
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading what was done to the store: %w", err)
	}
	return applied, nil
}

// record notes that a change has been made, and when.
func (s *Store) record(name string) error { return recordOn(s.on, name) }

func recordOn(on runs, name string) error {
	if _, err := on.Exec(
		`INSERT INTO migrations (name, applied_at) VALUES (?, ?)`,
		name, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("recording %q: %w", name, err)
	}
	return nil
}

// Migrations hands back every change this cairn knows, in order, saying which
// have been applied and when.
func (s *Store) Migrations() ([]migration.Migration, error) {
	applied, err := s.appliedNames()
	if err != nil {
		return nil, err
	}
	all := make([]migration.Migration, 0, len(migration.Known))
	for i, raw := range migration.Known {
		facts := factsOfChange(raw)
		var a migration.AsTold
		a.Numbered(i + 1)
		a.Named(facts.name)
		a.Does(facts.does)
		if facts.undoes == "" {
			a.NotUndoable()
		} else {
			a.Undoes(facts.undoes)
		}
		if facts.loses == "" {
			a.NothingLost()
		} else {
			a.Loses(facts.loses)
		}
		if made, done := applied[facts.name]; done {
			a.Applied(made)
		} else {
			a.NotApplied()
		}
		all = append(all, a.Migration())
	}
	return all, nil
}

// Waiting hands back the changes this cairn knows and the store has not.
func (s *Store) Waiting() ([]migration.Migration, error) {
	all, err := s.Migrations()
	if err != nil {
		return nil, err
	}
	waiting := []migration.Migration{}
	for _, m := range all {
		if !factsOfChange(m).applied {
			waiting = append(waiting, m)
		}
	}
	return waiting, nil
}

// Change hands back one change by its number, and says so when there is none.
// The numbering is the store's business, so the question "is there a change 4"
// is answered here and nowhere else.
func (s *Store) Change(number int) (migration.Migration, error) {
	all, err := s.Migrations()
	if err != nil {
		return migration.Migration{}, err
	}
	if number < 1 || number > len(all) {
		return migration.Migration{}, fmt.Errorf("change %d: %w", number, ErrNoSuchChange)
	}
	return all[number-1], nil
}

// Apply makes one change, named by its number.
func (s *Store) Apply(number int) (string, error) {
	one, err := s.Change(number)
	if err != nil {
		return "", err
	}
	facts := factsOfChange(one)
	if facts.applied {
		return "", fmt.Errorf("change %d: %w", number, ErrAlreadyApplied)
	}
	waiting, err := s.Waiting()
	if err != nil {
		return "", err
	}
	if first := factsOfChange(waiting[0]).number; first < number {
		return "", fmt.Errorf("change %d: %w — apply %d first", number, ErrEarlierWaiting, first)
	}
	return facts.name, s.make(migration.Known[number-1])
}

// ApplyAll makes every change the store is missing, oldest first.
func (s *Store) ApplyAll() ([]string, error) {
	waiting, err := s.Waiting()
	if err != nil {
		return nil, err
	}
	if len(waiting) == 0 {
		return nil, ErrNothingWaiting
	}
	return s.applyAll()
}

// applyAll makes everything missing without complaining when nothing is.
func (s *Store) applyAll() ([]string, error) {
	waiting, err := s.Waiting()
	if err != nil {
		return nil, err
	}
	made := []string{}
	for _, one := range waiting {
		facts := factsOfChange(one)
		if err := s.make(migration.Known[facts.number-1]); err != nil {
			return nil, err
		}
		made = append(made, facts.name)
	}
	return made, nil
}

// make runs one change and writes it down, as one transaction: a change is
// often several statements, and one failing half way used to leave the first
// ones made and the ledger silent about them — a store the next `apply` could
// only refuse, with nothing saying where it had stopped.
//
// Whether it is already applied is asked again inside: read before, it may
// have been applied since by another cairn, which then answered with the
// SQL refusing a column made twice rather than with what happened.
func (s *Store) make(raw migration.Migration) error {
	return s.Within(func(kept *Store) error {
		applied, err := appliedNamesOn(kept.on)
		if err != nil {
			return err
		}
		name := factsOfChange(raw).name
		if _, done := applied[name]; done {
			return fmt.Errorf("%q: %w", name, ErrAlreadyApplied)
		}
		return kept.makeOn(kept.on, raw)
	})
}

// makeOn applies a change against whatever it is given to run on: the store
// itself when a change is applied on its own, or the one transaction that makes
// a store when the whole shape goes down at once.
func (s *Store) makeOn(on runs, raw migration.Migration) error {
	facts := factsOfChange(raw)
	if _, err := on.Exec(facts.does); err != nil {
		return fmt.Errorf("applying %q: %w", facts.name, err)
	}
	return recordOn(on, facts.name)
}

// runs is what a change is applied against, and read from. The store and a
// transaction both answer it, which is what lets one act be one act — and what
// lets the act check, from inside itself, that it is still the one making the
// store.
type runs interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// RollBack unmakes the last change applied, keeping the work the older shape
// can still hold. It hands back the change it unmade, so a surface can say what
// that cost. It refuses to unmake the shape a store starts with.
func (s *Store) RollBack() (migration.Migration, error) {
	all, err := s.Migrations()
	if err != nil {
		return migration.Migration{}, err
	}
	return s.rollBackFrom(all)
}

// rollBackFrom unmakes the last change of all that is applied. What was read
// as applied is asked again inside the transaction: read before, it may have
// been undone since by another cairn, and two rollbacks then both said they
// had undone the same change.
func (s *Store) rollBackFrom(all []migration.Migration) (migration.Migration, error) {
	for i := len(all) - 1; i >= 1; i-- {
		if !factsOfChange(all[i]).applied {
			continue
		}
		raw := factsOfChange(migration.Known[i])
		// A change with no undoing changed data and not the shape: rolling it
		// back forgets it in the ledger and nothing else, which is what lets a
		// store go back past it to an older shape at all. What that leaves as
		// it was is said by its Loses, which the answer carries; applying it
		// again later runs an update that changes nothing the second time.
		//
		// Undone and forgotten together: a ledger that still lists a change
		// whose shape is gone, or forgets one whose shape is still there, is a
		// store no later command can reason about.
		err := s.Within(func(kept *Store) error {
			applied, err := appliedNamesOn(kept.on)
			if err != nil {
				return err
			}
			if _, still := applied[raw.name]; !still {
				return fmt.Errorf("%q: %w", raw.name, ErrUndoneMeanwhile)
			}
			if _, err := kept.on.Exec(raw.undoes); err != nil {
				return fmt.Errorf("rolling back %q: %w", raw.name, err)
			}
			if _, err := kept.on.Exec(`DELETE FROM migrations WHERE name = ?`, raw.name); err != nil {
				return fmt.Errorf("forgetting %q: %w", raw.name, err)
			}
			return nil
		})
		if err != nil {
			return migration.Migration{}, err
		}
		return all[i], nil
	}
	return migration.Migration{}, ErrNothingToRollBack
}

// Held counts what a store holds, tasks and objectives apart, without
// reading any of it.
func (s *Store) Held() (domain.Held, error) {
	held, err := heldOn(s.on)
	if err != nil {
		return domain.Held{}, fmt.Errorf("counting what the store holds: %w", err)
	}
	return held, nil
}
