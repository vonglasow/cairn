package engine

// The backup operations: a backup of the store is made, listed, read,
// deleted and put back, described once so both surfaces build the same
// command from the same facts. They open no store — a backup is most needed
// exactly when the store cannot be opened.

import (
	"cairn/internal/domain"
)

// resolveBackup is the backup a restore is aimed at: the one named, or the
// most recent when none was.
func resolveBackup(vault *Backups, in Input) (domain.Backup, error) {
	n, named := in.Whole("number")
	if !named {
		return vault.Latest()
	}
	return vault.Find(n)
}

// BackupAdd backs up the store as it stands.
var BackupAdd = Operation{
	Thing:   "backup",
	Verb:    "add",
	Short:   "Back up the store as it stands",
	Does:    "Back up the store as it stands.",
	Example: "cairn backup add",
	Level:   Write,
	Needs:   NoStore,
	Again:   DoesItAgain,
	Answers: BackupShape{},
	Run: func(on On, _ Input) (Answer, error) {
		// Counted before backing up, not after: what is announced is what was
		// taken.
		held, err := on.Backups.HoldsStore()
		if err != nil {
			return Answer{}, err
		}
		made, err := on.Backups.Make()
		if err != nil {
			return Answer{}, err
		}
		return Answer{Backup: &made, Holds: &held}, nil
	},
}

// BackupList names every backup.
var BackupList = Operation{
	Thing:   "backup",
	Verb:    "list",
	Short:   "List every backup",
	Does:    "List every backup.",
	Example: "cairn backup list",
	Level:   Read,
	Needs:   NoStore,
	Answers: Kept{},
	Run: func(on On, _ Input) (Answer, error) {
		all, err := on.Backups.All()
		if err != nil {
			return Answer{}, err
		}
		if all == nil {
			// No backups folder was ever created, and no backup is a set
			// with none in it — not an absent answer indistinguishable from
			// an operation that never touched this field at all.
			all = []domain.Backup{}
		}
		return Answer{Backups: all}, nil
	},
}

// BackupShow says when a backup was made and what it holds.
var BackupShow = Operation{
	Thing:   "backup",
	Verb:    "show",
	Short:   "Say when a backup was made and what it holds",
	Does:    "Say when a backup was made and what it holds.",
	Example: "cairn backup show 3",
	Level:   Read,
	Needs:   NoStore,
	Answers: BackupShape{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a backup", Means: "the number of the backup"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		one, err := on.Backups.Find(n)
		if err != nil {
			return Answer{}, err
		}
		held, err := on.Backups.Holds(one)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Backup: &one, Holds: &held}, nil
	},
}

// BackupDelete removes a backup.
var BackupDelete = Operation{
	Thing:   "backup",
	Verb:    "delete",
	Short:   "Delete a backup",
	Does:    "Delete a backup.",
	Example: "cairn backup delete 3 --yes",
	Level:   Full,
	Needs:   NoStore,
	Again:   ChangesNothingMore,
	Answers: Deleted{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a backup", Means: "the number of the backup"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		if err := on.Backups.Delete(n); err != nil {
			return Answer{}, err
		}
		return Answer{Deleted: &n}, nil
	},
}

// BackupRestore puts a backup back, the most recent one unless told otherwise.
var BackupRestore = Operation{
	Thing: "backup",
	Verb:  "restore",
	Short: "Put a backup back, the most recent one unless told otherwise",
	Does:  "Put a backup back, the most recent one unless told otherwise.",
	Example: `cairn backup restore --yes
cairn backup restore 3 --yes`,
	Level:   Full,
	Needs:   NoStore,
	Again:   DoesItAgain,
	Answers: PutBack{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Of: "a backup", Means: "the number of the backup"},
	},
	Run: func(on On, in Input) (Answer, error) {
		one, err := resolveBackup(on.Backups, in)
		if err != nil {
			return Answer{}, err
		}
		if err := on.Backups.Restore(one.Number); err != nil {
			return Answer{}, err
		}
		return Answer{PutBack: &one.Number}, nil
	},
}
