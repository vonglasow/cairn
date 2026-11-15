// Package backup holds what a store holds, and a backup of it.
package backup

import (
	"errors"
	"fmt"
	"time"
)

// Held is what a store, or a backup of one, contains.
type Held struct {
	Tasks      int `json:"tasks"`
	Objectives int `json:"objectives"`
}

// HeldText writes what is held the way a person reads it.
func HeldText(h Held) string {
	return fmt.Sprintf("holds %d task(s) and %d objective(s)", h.Tasks, h.Objectives)
}

// Backup is a backup of the store, kept beside it and referred to by its number, and how many bytes its file takes.
type Backup struct {
	Number int
	Made   time.Time
	Path   string
	Size   int64
}

// FormatSize writes a number of bytes the way a person reads one: whole bytes and kilobytes, megabytes to a tenth.
func FormatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%d KB", (bytes+512)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}

// The reasons a backup cannot be found.
var (
	ErrNoSuchBackup = errors.New("there is no such backup")
	ErrNoBackup     = errors.New("there is no backup to restore")

	// ErrNotAStore: the file is there and readable, and holds no tasks table.
	ErrNotAStore = errors.New("the store cannot be read")

	// ErrDamagedBackup: the backup asked for is not a database SQLite can read whole.
	ErrDamagedBackup = errors.New("the backup cannot be read whole")
)

// Vault is the backups kept beside a store.
type Vault interface {
	Make() (Backup, error)
	All() ([]Backup, error)
	Find(number int) (Backup, error)
	Latest() (Backup, error)
	Delete(number int) error
	Restore(number int) error
	HoldsStore() (Held, error)
	Holds(one Backup) (Held, error)
}
