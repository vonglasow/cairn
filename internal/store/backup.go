package store

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"cairn/internal/domain"
)

// The reasons a backup cannot be found.
var (
	ErrNoSuchBackup = errors.New("there is no such backup")
	ErrNoBackup     = errors.New("there is no backup to put back")

	// ErrNotAStore: the file is there and readable, and holds no tasks table.
	// Backing it up anyway would answer "backup made, 0 tasks", which turns a
	// stranger's file into an empty store of yours.
	ErrNotAStore = errors.New("the store cannot be read")
)

// backupFolder sits beside the store rather than inside it, so a backup is never
// mistaken for the thing it backups, and deleting the folder by hand is enough
// to be rid of them all.
const backupFolder = "backups"

// counterFile holds the last number handed out. It is kept out of the store on
// purpose: putting a backup back would otherwise put the counter back with it,
// and numbers already used would start coming round again.
const counterFile = ".last-number"

// Vault is where the backups of one store are kept.
type Vault struct {
	store  string
	folder string
}

// NewVault names the place the backups of a store live.
func NewVault(store string) *Vault {
	return &Vault{store: store, folder: filepath.Join(filepath.Dir(store), backupFolder)}
}

// Make backups the store as it stands and hands back the backup it made.
func (v *Vault) Make() (domain.Backup, error) {
	if err := os.MkdirAll(v.folder, 0o700); err != nil {
		return domain.Backup{}, fmt.Errorf("making room for the backups: %w", err)
	}
	number, err := v.nextNumber()
	if err != nil {
		return domain.Backup{}, err
	}
	path := v.pathOf(number)
	if err := snapshot(v.store, path); err != nil {
		return domain.Backup{}, err
	}
	if err := v.rememberNumber(number); err != nil {
		return domain.Backup{}, err
	}
	made, err := os.Stat(path)
	if err != nil {
		return domain.Backup{}, fmt.Errorf("reading the backup just made: %w", err)
	}
	return domain.Backup{Number: number, Made: made.ModTime(), Path: path}, nil
}

// All hands back every backup, oldest number first.
func (v *Vault) All() ([]domain.Backup, error) {
	entries, err := os.ReadDir(v.folder)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading the backups: %w", err)
	}

	backups := []domain.Backup{}
	for _, entry := range entries {
		number, ok := numberOf(entry.Name())
		if !ok {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("reading a backup: %w", err)
		}
		backups = append(backups, domain.Backup{
			Number: number, Made: info.ModTime(), Path: filepath.Join(v.folder, entry.Name()),
		})
	}
	return backups, nil
}

// Find hands back one backup by its number.
func (v *Vault) Find(number int) (domain.Backup, error) {
	backups, err := v.All()
	if err != nil {
		return domain.Backup{}, err
	}
	for _, one := range backups {
		if one.Number == number {
			return one, nil
		}
	}
	return domain.Backup{}, fmt.Errorf("backup %d: %w", number, ErrNoSuchBackup)
}

// Latest hands back the most recent backup.
func (v *Vault) Latest() (domain.Backup, error) {
	backups, err := v.All()
	if err != nil {
		return domain.Backup{}, err
	}
	if len(backups) == 0 {
		return domain.Backup{}, ErrNoBackup
	}
	return backups[len(backups)-1], nil
}

// Delete removes a backup.
func (v *Vault) Delete(number int) error {
	one, err := v.Find(number)
	if err != nil {
		return err
	}
	if err := os.Remove(one.Path); err != nil {
		return fmt.Errorf("deleting backup %d: %w", number, err)
	}
	return nil
}

// Restore puts a backup back, backing up first what it is about to replace: picking
// the wrong backup should cost a command, not the work it overwrote.
func (v *Vault) Restore(number int) error {
	one, err := v.Find(number)
	if err != nil {
		return err
	}
	if _, err := v.Make(); err != nil {
		return err
	}
	// Copied beside the store, then put in its place in one step, so nothing
	// ever opens half a store. What sat beside the old one goes first: a log
	// written for one file would be read on top of the other.
	arriving := v.store + ".arriving"
	if err := copyFile(one.Path, arriving); err != nil {
		return err
	}
	for _, beside := range []string{v.store + "-wal", v.store + "-shm"} {
		if err := os.Remove(beside); err != nil && !errors.Is(err, os.ErrNotExist) {
			_ = os.Remove(arriving)
			return fmt.Errorf("clearing what sat beside the store: %w", err)
		}
	}
	if err := os.Rename(arriving, v.store); err != nil {
		_ = os.Remove(arriving)
		return fmt.Errorf("putting backup %d in place: %w", number, err)
	}
	return nil
}

// Holds counts what a backup holds, without bringing it up to date: a backup is
// read as it was written, never changed by being looked at.
func (v *Vault) Holds(one domain.Backup) (domain.Held, error) {
	held, err := heldIn(one.Path)
	if err != nil {
		return domain.Held{}, fmt.Errorf("reading backup %d: %w", one.Number, err)
	}
	return held, nil
}

// HoldsStore counts what the store holds, read as it lies. A store one change
// behind still answers this: the columns differ, the table does not — which is
// why a backup of an older store can still say how much it took.
func (v *Vault) HoldsStore() (domain.Held, error) {
	held, err := heldIn(v.store)
	if err != nil {
		return domain.Held{}, fmt.Errorf("%w: %w — store: %s", ErrNotAStore, err, v.store)
	}
	return held, nil
}

// heldIn counts what a file holds, opening it as it lies — no shape brought up
// to date, no version checked. Looking at something must not change it.
//
// A backup taken before objectives existed has no table to count, and that is
// not a damaged backup: it is an older shape, and it holds no objectives. The
// table is looked for rather than the error swallowed, so a real failure to
// read one still comes back as a failure.
func heldIn(path string) (domain.Held, error) {
	db, err := sql.Open("sqlite", shared(path))
	if err != nil {
		return domain.Held{}, err
	}
	defer func() { _ = db.Close() }()

	return heldOn(db)
}

// heldOn counts what one open database holds. Shared by a backup read as it
// lies and by the store itself, because both can be one change behind: a
// backup taken before objectives existed has no table to count, and neither
// does a store somebody has just rolled back. That is not damage, it is an
// older shape holding no objectives.
//
// The table is looked for rather than the error swallowed, so a real failure
// to read one still comes back as a failure.
func heldOn(db runs) (domain.Held, error) {
	var held domain.Held

	// Both counts at once. SQLite compiles the whole statement, so a store
	// one change behind fails it outright rather than answering half of it.
	if err := db.QueryRow(
		`SELECT (SELECT count(*) FROM tasks), (SELECT count(*) FROM objectives)`,
	).Scan(&held.Tasks, &held.Objectives); err == nil {
		return held, nil
	}

	// The older shape, then: tasks and no objectives. Asked rather than
	// assumed — if the file is unreadable for any other reason, this fails
	// too, and the failure is what comes back.
	if err := db.QueryRow(`SELECT count(*) FROM tasks`).Scan(&held.Tasks); err != nil {
		return domain.Held{}, err
	}
	return domain.Held{Tasks: held.Tasks}, nil
}

// pathOf is where the backup carrying number lives.
func (v *Vault) pathOf(number int) string {
	return filepath.Join(v.folder, fmt.Sprintf("cairn-%04d.db", number))
}

// numberOf reads the number out of a backup's name, and says so when the name is
// not one of ours.
func numberOf(name string) (int, bool) {
	trimmed, ok := strings.CutPrefix(name, "cairn-")
	if !ok {
		return 0, false
	}
	trimmed, ok = strings.CutSuffix(trimmed, ".db")
	if !ok {
		return 0, false
	}
	number, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, false
	}
	return number, true
}

// nextNumber is the number the next backup gets: one past the last handed out,
// never one that was handed out before.
func (v *Vault) nextNumber() (int, error) {
	held, err := os.ReadFile(filepath.Join(v.folder, counterFile))
	if errors.Is(err, os.ErrNotExist) {
		return 1, nil
	}
	if err != nil {
		return 0, fmt.Errorf("reading the last backup number: %w", err)
	}
	last, err := strconv.Atoi(strings.TrimSpace(string(held)))
	if err != nil {
		return 0, fmt.Errorf("reading the last backup number: %w", err)
	}
	return last + 1, nil
}

// rememberNumber writes down the number just handed out.
func (v *Vault) rememberNumber(number int) error {
	path := filepath.Join(v.folder, counterFile)
	if err := os.WriteFile(path, []byte(strconv.Itoa(number)), 0o600); err != nil {
		return fmt.Errorf("remembering the last backup number: %w", err)
	}
	return nil
}

// snapshot writes the store as it stands to a new file: what is written back
// to it and what is still held aside in its log alike, read in one go so a
// write landing halfway is either all in or not at all. A store held open — a
// board on screen — keeps its latest writes in the log, and a copy of the file
// alone missed them.
//
// A file SQLite cannot read as a database is copied as it lies instead: a
// backup is most needed exactly when the store is damaged, and putting a
// backup back first sets the damaged store aside the same way.
//
// The store is opened to be read and nothing else: opened the usual way, a
// store that is not there would be made by being backed up.
func snapshot(from, to string) error {
	if _, err := os.Stat(from); err != nil {
		return fmt.Errorf("reading %s: %w", from, err)
	}
	db, err := sql.Open("sqlite", "file:"+from+"?mode=ro&_pragma=busy_timeout("+strconv.Itoa(heldFor)+")")
	if err != nil {
		return fmt.Errorf("reading %s: %w", from, err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(`VACUUM INTO ?`, to); err != nil {
		_ = os.Remove(to)
		return copyFile(from, to)
	}
	return nil
}

// copyFile copies from one path to another, replacing what is there.
func copyFile(from, to string) error {
	source, err := os.Open(from)
	if err != nil {
		return fmt.Errorf("reading %s: %w", from, err)
	}
	defer func() { _ = source.Close() }()

	target, err := os.Create(to)
	if err != nil {
		return fmt.Errorf("writing %s: %w", to, err)
	}
	if _, err := io.Copy(target, source); err != nil {
		_ = target.Close()
		return fmt.Errorf("backing up to %s: %w", to, err)
	}
	return target.Close()
}
