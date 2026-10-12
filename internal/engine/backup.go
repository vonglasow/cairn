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

// BackupNew backs up the store as it stands.
var BackupNew = Operation{
	Thing:   "backup",
	Verb:    "new",
	Short:   "Make a new backup of the store as it stands",
	Does:    "Make a new backup of the store as it stands.",
	Example: "cairn backup new",
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
	Short: "Restore a backup, the most recent one unless told otherwise",
	Does:  "Restore a backup, the most recent one unless told otherwise.",
	Example: `cairn backup restore --yes
cairn backup restore 3 --yes`,
	Level:   Full,
	Needs:   NoStore,
	Again:   DoesItAgain,
	Answers: Restored{},
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
		return Answer{Restored: &one.Number}, nil
	},
}

// Exported says which file the store was written to, and what it holds.
type Exported struct {
	Exported string       `json:"exported"`
	Holds    *domain.Held `json:"holds"`
}

// Imported says which file the store now holds what it held, and the backup
// it was taken in as.
type Imported struct {
	Imported string `json:"imported"`
	Backup   int    `json:"backup"`
}

// BackupExport writes the store to a file of a person's choosing, to carry it
// to another machine; the backups stay beside the store they back up.
var BackupExport = Operation{
	Thing:   "backup",
	Verb:    "export",
	Short:   "Write the store to a file of your choosing",
	Does:    "Write the store as it stands to a file of your choosing, to carry it elsewhere. A file already there is never written over.",
	Example: "cairn backup export ~/Desktop/cairn.db",
	Level:   Write,
	Needs:   NoStore,
	Again:   ChangesNothingMore,
	Answers: Exported{},
	Takes: []Argument{
		{Name: "to", Kind: Text, Positional: true, Required: true, Means: "the file to write the store to"},
	},
	Run: func(on On, in Input) (Answer, error) {
		to := in.Given["to"]
		held, err := on.Backups.Export(to)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Exported: &to, Holds: &held}, nil
	},
}

// BackupImport puts a store from a file in place of this one: taken in as a
// backup, then put back the way a backup is, what the store held backed up
// first. The file is left where it was.
var BackupImport = Operation{
	Thing:   "backup",
	Verb:    "import",
	Short:   "Put the store held in a file in place of this one",
	Does:    "Put the store held in a file in place of this one. It is taken in as a backup first, and a new backup of what the store held is made before it goes.",
	Example: "cairn backup import ~/Downloads/cairn.db --yes",
	Level:   Full,
	Needs:   NoStore,
	Again:   DoesItAgain,
	Answers: Imported{},
	Takes: []Argument{
		{Name: "from", Kind: Text, Positional: true, Required: true, Means: "the file holding the store to bring in"},
	},
	Run: func(on On, in Input) (Answer, error) {
		from := in.Given["from"]
		taken, err := on.Backups.Adopt(from)
		if err != nil {
			return Answer{}, err
		}
		if err := on.Backups.Restore(taken.Number); err != nil {
			return Answer{}, err
		}
		return Answer{Imported: &from, Restored: &taken.Number}, nil
	},
}
