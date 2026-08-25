package engine

import (
	"fmt"

	"cairn/internal/domain"
	"cairn/internal/migration"
	"cairn/internal/store"
)

// What a surface has to be able to recognise about a store and its backups.
// Re-exported here because a surface asks the engine and may not import the
// store, so the engine is where the names it answers with have to live.
var (
	ErrNoSuchBackup      = store.ErrNoSuchBackup
	ErrNoBackup          = store.ErrNoBackup
	ErrNotAStore         = store.ErrNotAStore
	ErrNewerStore        = store.ErrNewerStore
	ErrStoreBehind       = store.ErrStoreBehind
	ErrNoSuchChange      = store.ErrNoSuchChange
	ErrNothingWaiting    = store.ErrNothingWaiting
	ErrAlreadyApplied    = store.ErrAlreadyApplied
	ErrNothingToRollBack = store.ErrNothingToRollBack
)

// StoreUnusable says whether an error is the store being out of reach rather
// than the request being wrong. Re-exported the same way the sentinels above
// are: the store is the only package that may recognise it, and a surface has
// to be able to ask without ever hearing of SQLite.
var StoreUnusable = store.Unusable

// Backups is the set of backups kept beside a store.
//
// It is reached without opening the store on purpose: a backup is most needed
// exactly when the store cannot be opened — because it is broken, or because
// it comes from a cairn that knows more. Needing a working store to reach the
// backups would put them out of reach when they matter.
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

// Migrations hands back every change this cairn knows, applied or not.
func (e *Engine) Migrations() ([]migration.Migration, error) { return e.store.Migrations() }

// Change hands back one change by its number.
func (e *Engine) Change(number int) (migration.Migration, error) { return e.store.Change(number) }

// Waiting hands back the changes the store has not had yet.
func (e *Engine) Waiting() ([]migration.Migration, error) { return e.store.Waiting() }

// Held counts what the store holds, tasks and objectives apart.
func (e *Engine) Held() (domain.Held, error) { return e.store.Held() }

// Apply makes one change, after setting a backup aside: changing the shape of a
// store is not something to discover afterwards.
func (e *Engine) Apply(path string, number int) (string, error) {
	if err := e.backupFirst(path); err != nil {
		return "", err
	}
	return e.store.Apply(number)
}

// ApplyAll makes every change the store is missing, after setting a backup aside.
func (e *Engine) ApplyAll(path string) ([]string, error) {
	if err := e.backupFirst(path); err != nil {
		return nil, err
	}
	return e.store.ApplyAll()
}

// RollBack unmakes the last change applied, after setting a backup aside: rolling
// back drops what the older shape cannot hold, and that is worth being able to
// take back.
func (e *Engine) RollBack(path string) (migration.Migration, error) {
	if err := e.backupFirst(path); err != nil {
		return migration.Migration{}, err
	}
	return e.store.RollBack()
}

// backupFirst sets a backup aside before the shape of the store is touched.
func (e *Engine) backupFirst(path string) error {
	if _, err := BackupsAt(path).Make(); err != nil {
		return fmt.Errorf("setting a backup aside before changing the store: %w", err)
	}
	return nil
}
