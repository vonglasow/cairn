package engine

import (
	"errors"
	"fmt"

	"cairn/internal/domain"
	"cairn/internal/migration"
	"cairn/internal/store"
)

// What a surface has to be able to recognise about a store and its backups.
var (
	ErrNoSuchBackup      = store.ErrNoSuchBackup
	ErrNoBackup          = store.ErrNoBackup
	ErrDamagedBackup     = store.ErrDamagedBackup
	ErrNotAStore         = store.ErrNotAStore
	ErrNewerStore        = store.ErrNewerStore
	ErrStoreBehind       = store.ErrStoreBehind
	ErrNoSuchChange      = store.ErrNoSuchChange
	ErrNothingWaiting    = store.ErrNothingWaiting
	ErrAlreadyApplied    = store.ErrAlreadyApplied
	ErrNothingToRollBack = store.ErrNothingToRollBack
	ErrUndoneMeanwhile   = store.ErrUndoneMeanwhile
	ErrEarlierWaiting    = store.ErrEarlierWaiting
)

// StoreUnusable says whether an error is the store being out of reach rather than the request being wrong.
var StoreUnusable = store.Unusable

// Backups is the set of backups kept beside a store.
type Backups struct {
	vault *store.Vault
}

// BackupsAt names the backups kept beside the store at path.
func BackupsAt(path string) *Backups {
	return &Backups{vault: store.NewVault(path)}
}

// Make backups the store as it stands.
func (c *Backups) Make() (domain.Backup, error) { return c.vault.Make() }

// All hands back every backup, oldest number first.
func (c *Backups) All() ([]domain.Backup, error) { return c.vault.All() }

// Find hands back one backup by its number.
func (c *Backups) Find(number int) (domain.Backup, error) { return c.vault.Find(number) }

// Latest hands back the most recent backup.
func (c *Backups) Latest() (domain.Backup, error) { return c.vault.Latest() }

// Delete removes a backup.
func (c *Backups) Delete(number int) error { return c.vault.Delete(number) }

// Restore puts a backup back, backing up first what it replaces.
func (c *Backups) Restore(number int) error { return c.vault.Restore(number) }

// HoldsStore counts what the store holds, read as it lies.
func (c *Backups) HoldsStore() (domain.Held, error) { return c.vault.HoldsStore() }

// Holds counts what a backup holds, without changing it.
func (c *Backups) Holds(one domain.Backup) (domain.Held, error) { return c.vault.Holds(one) }

func (e *Engine) migrations() ([]migration.Migration, error) { return e.store.Migrations() }

func (e *Engine) change(number int) (migration.Migration, error) { return e.store.Change(number) }

func (e *Engine) held() (domain.Held, error) { return e.store.Held() }

// Version says whether the store has moved since it was last asked.
func (e *Engine) Version() (int64, error) {
	defer e.taking()()
	if err := e.following(); err != nil {
		return 0, Told(err)
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
	if _, err := BackupsAt(path).Make(); err != nil {
		return fmt.Errorf("setting a backup aside before changing the store: %w", err)
	}
	return nil
}

// Told adds what somebody has to type about a store that cannot be used as it stands.
func Told(err error) error {
	if errors.Is(err, ErrStoreBehind) {
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
