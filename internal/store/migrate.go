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
	// ErrNewerStore: an older binary cannot know what a later one wrote, and writing into it anyway is how a store gets quietly wrecked.
	ErrNewerStore = errors.New("the store comes from a newer cairn")

	// ErrStoreBehind: the store is missing a change this cairn needs.
	ErrStoreBehind = errors.New("the store is behind this cairn")

	// ErrNoSuchChange: a number that names no change.
	ErrNoSuchChange = errors.New("there is no such change")

	// ErrNothingWaiting: everything this cairn knows is already applied.
	ErrNothingWaiting = errors.New("there is nothing waiting to be applied")

	// ErrAlreadyApplied: a change named by number that the store already has.
	ErrAlreadyApplied = errors.New("the change is already applied")

	// ErrNothingToRollBack: only the shape a store starts with is left, and that one is not a change on top of something else — undoing it would mean throwing the store away, which is what deleting the file is for.
	ErrNothingToRollBack = errors.New("there is nothing left to roll back")

	// ErrUndoneMeanwhile: the change read as the last applied was undone by another cairn between the reading and the rollback.
	ErrUndoneMeanwhile = errors.New("the change was rolled back by somebody else meanwhile")

	// ErrEarlierWaiting: a change was asked for while one before it is still waiting.
	ErrEarlierWaiting = errors.New("an earlier change is still waiting")
)

const ledger = `
CREATE TABLE IF NOT EXISTS migrations (
	name       TEXT NOT NULL PRIMARY KEY,
	applied_at TEXT NOT NULL
);`

func factsOfChange(m migration.Migration) migration.Carried { return m.Carried() }

func (s *Store) settle(mustBeCurrent bool) error {
	if err := s.makeTheLedgerWhenOpeningFindsNone(); err != nil {
		return err
	}
	empty, err := s.isEmpty()
	if err != nil {
		return err
	}
	if empty {
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
	return s.Current()
}

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

func (s *Store) layDownTheShapeWhenOpeningFindsNothing() error {
	making, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("making the store: %w", err)
	}
	defer func() { _ = making.Rollback() }()

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

func (s *Store) adoptTheUnrecordedShapeWhileOpening() error {
	applied, err := s.appliedNames()
	if err != nil {
		return err
	}
	if len(applied) > 0 {
		return nil
	}
	return s.record(factsOfChange(migration.Known[0]).Name)
}

func (s *Store) refuseAnUnknownChange() error {
	applied, err := s.appliedNames()
	if err != nil {
		return err
	}
	known := make(map[string]bool, len(migration.Known))
	for _, raw := range migration.Known {
		known[factsOfChange(raw).Name] = true
	}
	for name := range applied {
		if !known[name] {
			return fmt.Errorf("%w: it carries %d change(s) including %q, this cairn knows %d",
				ErrNewerStore, len(applied), name, len(migration.Known))
		}
	}
	return nil
}

// Current says whether the store has every change this cairn knows, and by how many it is behind when it has not.
func (s *Store) Current() error {
	waiting, err := s.Waiting()
	if err != nil {
		return err
	}
	if len(waiting) > 0 {
		return fmt.Errorf("%w by %d change(s)", ErrStoreBehind, len(waiting))
	}
	return nil
}

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

func (s *Store) record(name string) error { return recordOn(s.on, name) }

func recordOn(on runs, name string) error {
	if _, err := on.Exec(
		`INSERT INTO migrations (name, applied_at) VALUES (?, ?)`,
		name, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("recording %q: %w", name, err)
	}
	return nil
}

// Migrations hands back every change this cairn knows, in order, saying which have been applied and when.
func (s *Store) Migrations() ([]migration.Migration, error) {
	applied, err := s.appliedNames()
	if err != nil {
		return nil, err
	}
	all := make([]migration.Migration, 0, len(migration.Known))
	for i, raw := range migration.Known {
		facts := factsOfChange(raw)
		facts.Number = i + 1
		facts.Made, facts.Applied = applied[facts.Name]
		all = append(all, migration.Made(facts))
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
		if !factsOfChange(m).Applied {
			waiting = append(waiting, m)
		}
	}
	return waiting, nil
}

// Change hands back one change by its number, and says so when there is none.
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
	if facts.Applied {
		return "", fmt.Errorf("change %d: %w", number, ErrAlreadyApplied)
	}
	waiting, err := s.Waiting()
	if err != nil {
		return "", err
	}
	if first := factsOfChange(waiting[0]).Number; first < number {
		return "", fmt.Errorf("change %d: %w — apply %d first", number, ErrEarlierWaiting, first)
	}
	return facts.Name, s.make(migration.Known[number-1])
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

func (s *Store) applyAll() ([]string, error) {
	waiting, err := s.Waiting()
	if err != nil {
		return nil, err
	}
	made := []string{}
	for _, one := range waiting {
		facts := factsOfChange(one)
		if err := s.make(migration.Known[facts.Number-1]); err != nil {
			return nil, err
		}
		made = append(made, facts.Name)
	}
	return made, nil
}

func (s *Store) make(raw migration.Migration) error {
	return s.Within(func(kept *Store) error {
		applied, err := appliedNamesOn(kept.on)
		if err != nil {
			return err
		}
		name := factsOfChange(raw).Name
		if _, done := applied[name]; done {
			return fmt.Errorf("%q: %w", name, ErrAlreadyApplied)
		}
		return kept.makeOn(kept.on, raw)
	})
}

func (s *Store) makeOn(on runs, raw migration.Migration) error {
	facts := factsOfChange(raw)
	if _, err := on.Exec(facts.Does); err != nil {
		return fmt.Errorf("applying %q: %w", facts.Name, err)
	}
	return recordOn(on, facts.Name)
}

type runs interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// RollBack unmakes the last change applied, keeping the work the older shape can still hold.
func (s *Store) RollBack() (migration.Migration, error) {
	all, err := s.Migrations()
	if err != nil {
		return migration.Migration{}, err
	}
	return s.rollBackFrom(all)
}

func (s *Store) rollBackFrom(all []migration.Migration) (migration.Migration, error) {
	for i := len(all) - 1; i >= 1; i-- {
		if !factsOfChange(all[i]).Applied {
			continue
		}
		raw := factsOfChange(migration.Known[i])
		err := s.Within(func(kept *Store) error {
			applied, err := appliedNamesOn(kept.on)
			if err != nil {
				return err
			}
			if _, still := applied[raw.Name]; !still {
				return fmt.Errorf("%q: %w", raw.Name, ErrUndoneMeanwhile)
			}
			if _, err := kept.on.Exec(raw.Undoes); err != nil {
				return fmt.Errorf("rolling back %q: %w", raw.Name, err)
			}
			if _, err := kept.on.Exec(`DELETE FROM migrations WHERE name = ?`, raw.Name); err != nil {
				return fmt.Errorf("forgetting %q: %w", raw.Name, err)
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

// Held counts what a store holds, tasks and objectives apart, without reading any of it.
func (s *Store) Held() (domain.Held, error) {
	held, err := heldOn(s.on)
	if err != nil {
		return domain.Held{}, fmt.Errorf("counting what the store holds: %w", err)
	}
	return held, nil
}
