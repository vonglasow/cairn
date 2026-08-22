package store

import (
	"errors"
	"fmt"
	"time"

	"cairn/internal/domain"
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
)

// migration is one step in the shape of the store. Each carries how to make it
// and how to unmake it, because a shape that cannot be undone is a shape you
// can only ever move away from.
type migration struct {
	name string
	up   string
	down string
	// loses is what rolling this one back drops, in the words of somebody who
	// has the tasks rather than the schema. Empty when nothing is lost. Writing
	// it is the point: a migration whose author cannot say what undoing it
	// costs has not finished being thought about.
	loses string
}

// migrations, oldest first. Appending one here is what a store then has to be
// told to apply; the name is what gets recorded, so it is written for somebody
// reading `cairn migration list` a year from now.
var migrations = []migration{
	{
		name: "the shape a store starts with",
		up: `
CREATE TABLE IF NOT EXISTS tasks (
	number      INTEGER PRIMARY KEY AUTOINCREMENT,
	title       TEXT    NOT NULL,
	description TEXT    NOT NULL DEFAULT '',
	state       TEXT    NOT NULL,
	priority    TEXT    NOT NULL,
	deadline    TEXT    NOT NULL DEFAULT '',
	estimate    INTEGER NOT NULL DEFAULT 0,
	time_spent  INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS labels (
	task_number INTEGER NOT NULL REFERENCES tasks(number) ON DELETE CASCADE,
	label       TEXT    NOT NULL,
	PRIMARY KEY (task_number, label)
);`,
		// Left empty on purpose: this one is the floor, see ErrNothingToRollBack.
		down: "",
	},
	{
		name:  "a task can hang under another",
		up:    `ALTER TABLE tasks ADD COLUMN parent INTEGER NOT NULL DEFAULT 0`,
		down:  `ALTER TABLE tasks DROP COLUMN parent`,
		loses: "which task each sub-task hangs under",
	},
	{
		name: "objectives, and the tasks that serve them",
		up: `
CREATE TABLE IF NOT EXISTS objectives (
	number      INTEGER PRIMARY KEY AUTOINCREMENT,
	title       TEXT    NOT NULL,
	description TEXT    NOT NULL DEFAULT '',
	state       TEXT    NOT NULL,
	priority    TEXT    NOT NULL,
	deadline    TEXT    NOT NULL DEFAULT '',
	estimate    INTEGER NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS objective_labels (
	objective_number INTEGER NOT NULL REFERENCES objectives(number) ON DELETE CASCADE,
	label            TEXT    NOT NULL,
	PRIMARY KEY (objective_number, label)
);
ALTER TABLE tasks ADD COLUMN objective INTEGER NOT NULL DEFAULT 0;`,
		down: `
DROP TABLE IF EXISTS objective_labels;
DROP TABLE IF EXISTS objectives;
ALTER TABLE tasks DROP COLUMN objective;`,
		loses: "every objective, and which one each task serves",
	},
}

// ledger is where the store records what has been done to it.
const ledger = `
CREATE TABLE IF NOT EXISTS migrations (
	name       TEXT NOT NULL PRIMARY KEY,
	applied_at TEXT NOT NULL
);`

// settle prepares a freshly opened store: it makes sure the ledger is there,
// creates the shape when the file holds nothing at all, and otherwise reports
// what the store is missing without touching it.
func (s *Store) settle(mustBeCurrent bool) error {
	// Looked for, not laid down. Opening a store reads its shape; `migration
	// apply` is what changes it. count(*) always answers, so there is one way
	// for this to fail rather than two.
	var kept int
	if err := s.db.QueryRow(
		`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'migrations'`,
	).Scan(&kept); err != nil {
		return fmt.Errorf("reading what was done to the store: %w", err)
	}
	if kept == 0 {
		if _, err := s.db.Exec(ledger); err != nil {
			return fmt.Errorf("making room to record what is done to the store: %w", err)
		}
	}
	empty, err := s.isEmpty()
	if err != nil {
		return err
	}
	if empty {
		// Making a store is not changing one. A file that held nothing a moment
		// ago has nothing to lose and nobody to warn, so it is created at the
		// shape this cairn knows.
		_, err := s.applyAll()
		return err
	}
	if err := s.adoptUnrecorded(); err != nil {
		return err
	}
	if err := s.refuseUnknown(); err != nil {
		return err
	}
	if !mustBeCurrent {
		return nil
	}
	waiting, err := s.Waiting()
	if err != nil {
		return err
	}
	if len(waiting) > 0 {
		return fmt.Errorf("%w by %d change(s)", ErrStoreBehind, len(waiting))
	}
	return nil
}

// isEmpty says whether the file holds nothing of ours at all.
func (s *Store) isEmpty() (bool, error) {
	var count int
	if err := s.db.QueryRow(
		`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name IN ('tasks', 'labels')`).
		Scan(&count); err != nil {
		return false, fmt.Errorf("looking at the shape of the store: %w", err)
	}
	if count > 0 {
		return false, nil
	}
	applied, err := s.appliedNames()
	if err != nil {
		return false, err
	}
	return len(applied) == 0, nil
}

// adoptUnrecorded records the first shape for a store written before cairn kept
// a ledger. Reading it as untouched would have it apply a shape already there.
func (s *Store) adoptUnrecorded() error {
	applied, err := s.appliedNames()
	if err != nil {
		return err
	}
	if len(applied) > 0 {
		return nil
	}
	return s.record(migrations[0].name)
}

// refuseUnknown stops on a store carrying a change this cairn does not know.
func (s *Store) refuseUnknown() error {
	applied, err := s.appliedNames()
	if err != nil {
		return err
	}
	known := make(map[string]bool, len(migrations))
	for _, m := range migrations {
		known[m.name] = true
	}
	for name := range applied {
		if !known[name] {
			return fmt.Errorf("%w: it carries %d change(s) including %q, this cairn knows %d",
				ErrNewerStore, len(applied), name, len(migrations))
		}
	}
	return nil
}

// appliedNames reads what the store says has been done to it.
func (s *Store) appliedNames() (map[string]time.Time, error) {
	rows, err := s.db.Query(`SELECT name, applied_at FROM migrations`)
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
		made, _ := time.Parse(time.RFC3339, at)
		applied[name] = made
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading what was done to the store: %w", err)
	}
	return applied, nil
}

// record notes that a change has been made, and when.
func (s *Store) record(name string) error {
	if _, err := s.db.Exec(
		`INSERT INTO migrations (name, applied_at) VALUES (?, ?)`,
		name, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("recording %q: %w", name, err)
	}
	return nil
}

// Migrations hands back every change this cairn knows, in order, saying which
// have been applied and when.
func (s *Store) Migrations() ([]domain.Migration, error) {
	applied, err := s.appliedNames()
	if err != nil {
		return nil, err
	}
	all := make([]domain.Migration, 0, len(migrations))
	for i, m := range migrations {
		made, done := applied[m.name]
		all = append(all, domain.Migration{
			Number: i + 1, Name: m.name, Does: m.up, Undoes: m.down, Loses: m.loses,
			Applied: done, Made: made,
		})
	}
	return all, nil
}

// Waiting hands back the changes this cairn knows and the store has not.
func (s *Store) Waiting() ([]domain.Migration, error) {
	all, err := s.Migrations()
	if err != nil {
		return nil, err
	}
	waiting := []domain.Migration{}
	for _, m := range all {
		if !m.Applied {
			waiting = append(waiting, m)
		}
	}
	return waiting, nil
}

// Change hands back one change by its number, and says so when there is none.
// The numbering is the store's business, so the question "is there a change 4"
// is answered here and nowhere else.
func (s *Store) Change(number int) (domain.Migration, error) {
	all, err := s.Migrations()
	if err != nil {
		return domain.Migration{}, err
	}
	if number < 1 || number > len(all) {
		return domain.Migration{}, fmt.Errorf("change %d: %w", number, ErrNoSuchChange)
	}
	return all[number-1], nil
}

// Apply makes one change, named by its number.
func (s *Store) Apply(number int) (string, error) {
	one, err := s.Change(number)
	if err != nil {
		return "", err
	}
	if one.Applied {
		return "", fmt.Errorf("change %d: %w", number, ErrAlreadyApplied)
	}
	return one.Name, s.make(migrations[number-1])
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
		if err := s.make(migrations[one.Number-1]); err != nil {
			return nil, err
		}
		made = append(made, one.Name)
	}
	return made, nil
}

// make runs one change and writes it down.
func (s *Store) make(m migration) error {
	if _, err := s.db.Exec(m.up); err != nil {
		return fmt.Errorf("applying %q: %w", m.name, err)
	}
	return s.record(m.name)
}

// RollBack unmakes the last change applied, keeping the work the older shape
// can still hold. It hands back the change it unmade, so a surface can say what
// that cost. It refuses to unmake the shape a store starts with.
func (s *Store) RollBack() (domain.Migration, error) {
	all, err := s.Migrations()
	if err != nil {
		return domain.Migration{}, err
	}
	for i := len(all) - 1; i >= 1; i-- {
		if !all[i].Applied {
			continue
		}
		m := migrations[i]
		if _, err := s.db.Exec(m.down); err != nil {
			return domain.Migration{}, fmt.Errorf("rolling back %q: %w", m.name, err)
		}
		if _, err := s.db.Exec(`DELETE FROM migrations WHERE name = ?`, m.name); err != nil {
			return domain.Migration{}, fmt.Errorf("forgetting %q: %w", m.name, err)
		}
		return all[i], nil
	}
	return domain.Migration{}, ErrNothingToRollBack
}

// Held counts what a store holds, tasks and objectives apart, without
// reading any of it.
func (s *Store) Held() (domain.Held, error) {
	held, err := heldOn(s.db)
	if err != nil {
		return domain.Held{}, fmt.Errorf("counting what the store holds: %w", err)
	}
	return held, nil
}
