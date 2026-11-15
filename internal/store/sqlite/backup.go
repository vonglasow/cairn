package sqlite

import (
	"cairn/internal/engine/backup"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const backupFolder = "backups"

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
func (v *Vault) Make() (backup.Backup, error) { return v.take(v.store) }

func (v *Vault) take(from string) (backup.Backup, error) {
	if err := os.MkdirAll(v.folder, 0o700); err != nil {
		return backup.Backup{}, fmt.Errorf("making room for the backups: %w", err)
	}
	number, err := v.nextNumber()
	if err != nil {
		return backup.Backup{}, err
	}
	err = claim(v.pathOf(number))
	for errors.Is(err, os.ErrExist) {
		number++
		err = claim(v.pathOf(number))
	}
	if err != nil {
		return backup.Backup{}, fmt.Errorf("making backup %d: %w", number, err)
	}
	path := v.pathOf(number)
	if err := snapshot(from, path); err != nil {
		_ = os.Remove(path)
		return backup.Backup{}, err
	}
	if err := v.rememberNumber(number); err != nil {
		return backup.Backup{}, err
	}
	made, err := os.Stat(path)
	if err != nil {
		return backup.Backup{}, fmt.Errorf("reading the backup just made: %w", err)
	}
	return backup.Backup{Number: number, Made: made.ModTime(), Path: path, Size: made.Size()}, nil
}

// All hands back every backup, oldest number first.
func (v *Vault) All() ([]backup.Backup, error) {
	entries, err := os.ReadDir(v.folder)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading the backups: %w", err)
	}

	backups := []backup.Backup{}
	for _, entry := range entries {
		number, ok := numberOf(entry.Name())
		if !ok {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("reading a backup: %w", err)
		}
		backups = append(backups, backup.Backup{
			Number: number, Made: info.ModTime(), Path: filepath.Join(v.folder, entry.Name()), Size: info.Size(),
		})
	}
	return backups, nil
}

// Find hands back one backup by its number.
func (v *Vault) Find(number int) (backup.Backup, error) {
	backups, err := v.All()
	if err != nil {
		return backup.Backup{}, err
	}
	for _, one := range backups {
		if one.Number == number {
			return one, nil
		}
	}
	return backup.Backup{}, fmt.Errorf("backup %d: %w", number, backup.ErrNoSuchBackup)
}

// Latest hands back the most recent backup.
func (v *Vault) Latest() (backup.Backup, error) {
	backups, err := v.All()
	if err != nil {
		return backup.Backup{}, err
	}
	if len(backups) == 0 {
		return backup.Backup{}, backup.ErrNoBackup
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

// Restore puts a backup back, backing up first what it is about to replace: picking the wrong backup should cost a command, not the work it overwrote.
func (v *Vault) Restore(number int) error {
	one, err := v.Find(number)
	if err != nil {
		return err
	}
	if err := whole(one.Path); err != nil {
		return fmt.Errorf("backup %d: %w: %w — backup: %s", number, backup.ErrDamagedBackup, err, one.Path)
	}
	if _, err := v.Make(); err != nil {
		return err
	}
	arriving, err := os.CreateTemp(filepath.Dir(v.store), filepath.Base(v.store)+".arriving-*")
	if err != nil {
		return fmt.Errorf("making room for backup %d: %w", number, err)
	}
	defer func() { _ = os.Remove(arriving.Name()) }()
	if err := copyInto(one.Path, arriving); err != nil {
		return err
	}
	for _, beside := range []string{v.store + "-wal", v.store + "-shm"} {
		if err := os.Remove(beside); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("clearing what sat beside the store: %w", err)
		}
	}
	if err := os.Rename(arriving.Name(), v.store); err != nil {
		return fmt.Errorf("putting backup %d in place: %w", number, err)
	}
	return nil
}

// Holds counts what a backup holds, without bringing it up to date: a backup is read as it was written, never changed by being looked at.
func (v *Vault) Holds(one backup.Backup) (backup.Held, error) {
	held, err := heldIn(one.Path)
	if err != nil {
		return backup.Held{}, fmt.Errorf("reading backup %d: %w", one.Number, err)
	}
	return held, nil
}

// HoldsStore counts what the store holds, read as it lies.
func (v *Vault) HoldsStore() (backup.Held, error) {
	held, err := heldIn(v.store)
	if err != nil {
		return backup.Held{}, fmt.Errorf("%w: %w — store: %s", backup.ErrNotAStore, err, v.store)
	}
	return held, nil
}

func heldIn(path string) (backup.Held, error) {
	db, err := readOnly(path)
	if err != nil {
		return backup.Held{}, err
	}
	defer func() { _ = db.Close() }()

	return heldOn(db)
}

func heldOn(db runs) (backup.Held, error) {
	var held backup.Held

	if err := db.QueryRow(
		`SELECT (SELECT count(*) FROM tasks), (SELECT count(*) FROM objectives)`,
	).Scan(&held.Tasks, &held.Objectives); err == nil {
		return held, nil
	}

	if err := db.QueryRow(`SELECT count(*) FROM tasks`).Scan(&held.Tasks); err != nil {
		return backup.Held{}, err
	}
	return backup.Held{Tasks: held.Tasks}, nil
}

func (v *Vault) pathOf(number int) string {
	return filepath.Join(v.folder, fmt.Sprintf("cairn-%04d.db", number))
}

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

func (v *Vault) rememberNumber(number int) error {
	path := filepath.Join(v.folder, counterFile)
	if err := os.WriteFile(path, []byte(strconv.Itoa(number)), 0o600); err != nil {
		return fmt.Errorf("remembering the last backup number: %w", err)
	}
	return nil
}

func snapshot(from, to string) error {
	if _, err := os.Stat(from); err != nil {
		return fmt.Errorf("reading %s: %w", from, err)
	}
	db, err := readOnly(from)
	if err != nil {
		return fmt.Errorf("reading %s: %w", from, err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(`VACUUM INTO ?`, to); err != nil {
		if whole(from) == nil {
			return fmt.Errorf("backing up %s: %w", from, err)
		}
		return copyFile(from, to)
	}
	return nil
}

func readOnly(path string) (*sql.DB, error) {
	return sql.Open("sqlite", addressOf(path, "mode=ro&_pragma=busy_timeout("+strconv.Itoa(heldFor)+")"))
}

func whole(path string) error {
	db, err := readOnly(path)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	said := "it is not a database SQLite can read"
	_ = db.QueryRow(`PRAGMA quick_check`).Scan(&said)
	if said != "ok" {
		return errors.New(said)
	}
	if _, err := heldOn(db); err != nil {
		return fmt.Errorf("it holds no tasks: %w", err)
	}
	return nil
}

func copyFile(from, to string) error {
	target, err := os.OpenFile(to, os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("writing %s: %w", to, err)
	}
	return copyInto(from, target)
}

func copyInto(from string, target *os.File) error {
	source, err := os.Open(from)
	if err != nil {
		_ = target.Close()
		return fmt.Errorf("reading %s: %w", from, err)
	}
	defer func() { _ = source.Close() }()
	if _, err := io.Copy(target, source); err != nil {
		_ = target.Close()
		return fmt.Errorf("copying to %s: %w", target.Name(), err)
	}
	if err := target.Sync(); err != nil {
		_ = target.Close()
		return fmt.Errorf("copying to %s: %w", target.Name(), err)
	}
	return target.Close()
}
