package engine

import (
	"cairn/internal/engine/backup"
	"errors"
	"fmt"

	"cairn/internal/engine/migration"
	"cairn/internal/store/sqlite"
)

// storeUnusable says whether an error is the store being out of reach rather than the request being wrong.
var storeUnusable = sqlite.Unusable

// backupsAt names the backups kept beside the store at path.
func backupsAt(path string) backup.Vault { return sqlite.NewVault(path) }

func (e *Engine) migrations() ([]migration.Migration, error) { return e.store.Migrations() }

func (e *Engine) change(number int) (migration.Migration, error) { return e.store.Change(number) }

func (e *Engine) held() (backup.Held, error) { return e.store.Held() }

// Version says whether the store has moved since it was last asked.
func (e *Engine) Version() (version int64, err error) {
	defer func() { noteVersion(version, err) }()
	defer e.taking()()
	if err := e.following(); err != nil {
		return 0, Told(err)
	}
	if e.refused != nil {
		return 0, nil
	}
	moved, err := e.store.Version()
	return moved + e.followed<<32, err
}

func (e *Engine) apply(path string, number int) (string, error) {
	if err := e.backupFirst(path); err != nil {
		return "", err
	}
	return e.store.Apply(number)
}

func (e *Engine) applyAll(path string) ([]string, error) {
	if err := e.backupFirst(path); err != nil {
		return nil, err
	}
	return e.store.ApplyAll()
}

func (e *Engine) rollBack(path string) (migration.Migration, error) {
	if err := e.backupFirst(path); err != nil {
		return migration.Migration{}, err
	}
	return e.store.RollBack()
}

func (e *Engine) backupFirst(path string) error {
	if _, err := backupsAt(path).Make(); err != nil {
		return fmt.Errorf("setting a backup aside before changing the store: %w", err)
	}
	return nil
}

// Told adds what somebody has to type about a store that cannot be used as it stands.
func Told(err error) error {
	if errors.Is(err, migration.ErrStoreBehind) {
		return typed{err, "cairn migration apply"}
	}
	return err
}

type typed struct {
	err     error
	command string
}

func (t typed) Error() string { return fmt.Sprintf("%s — run `%s`", t.err, t.command) }
func (t typed) Unwrap() error { return t.err }

// ToType splits a refusal that names a command to type into what it says and that command, so a person can be shown the command alone on its line.
func ToType(err error) (said, command string, ok bool) {
	var t typed
	if !errors.As(err, &t) {
		return err.Error(), "", false
	}
	return t.err.Error(), t.command, true
}
