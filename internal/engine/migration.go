package engine

import (
	"errors"

	"cairn/internal/migration"
)

// MigrationList names every change to the store, applied or waiting.
var MigrationList = Operation{
	Thing:   "migration",
	Verb:    "list",
	Short:   "Name every change to the store, applied or waiting",
	Does:    "Name every change to the store, applied or waiting.",
	Example: "cairn migration list",
	Level:   Read,
	Needs:   aStoreBehind,
	Answers: ledger{},
	Run: func(on workedOn, _ Input) (Answer, error) {
		all, err := on.Engine.migrations()
		if err != nil {
			return Answer{}, err
		}
		held, err := on.Engine.held()
		if err != nil {
			return Answer{}, err
		}
		return Answer{Changes: all, Holds: &held}, nil
	},
}

// migrationShow says what a change does, and what undoing it does.
var migrationShow = Operation{
	Thing:   "migration",
	Verb:    "show",
	Short:   "Say what a change does, and what undoing it does",
	Does:    "Say what a change does, and what undoing it does.",
	Example: "cairn migration show 4",
	Level:   Read,
	Needs:   aStoreBehind,
	Answers: migration.Shape{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a migration", Means: "the number of the change"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		n, _ := in.whole("number")
		one, err := on.Engine.change(n)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Change: &one}, nil
	},
}

// MigrationApply makes the changes the store is missing, or one of them.
var MigrationApply = Operation{
	Thing: "migration",
	Verb:  "apply",
	Short: "Make the changes the store is missing, or one of them",
	Does:  "Make the changes the store is missing, or one of them.",
	Example: `cairn migration apply
cairn migration apply 4`,
	Level:   write,
	Needs:   aStoreBehind,
	Again:   ChangesNothingMore,
	Answers: applied{},
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Of: "a migration", Means: "the number of the change"},
	},
	Run: func(on workedOn, in Input) (Answer, error) {
		n, named := in.whole("number")
		if !named {
			made, err := on.Engine.applyAll(on.Path)
			if errors.Is(err, errNothingWaiting) {
				return Answer{Applied: []string{}}, nil
			}
			if err != nil {
				return Answer{}, err
			}
			return Answer{Applied: made}, nil
		}
		one, err := on.Engine.apply(on.Path, n)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Applied: []string{one}}, nil
	},
}

// MigrationRollback unmakes the last change applied to the store.
var MigrationRollback = Operation{
	Thing:   "migration",
	Verb:    "rollback",
	Short:   "Unmake the last change applied to the store",
	Does:    "Unmake the last change applied to the store.",
	Example: "cairn migration rollback --yes",
	Level:   Full,
	Needs:   aStoreBehind,
	Again:   doesItAgain,
	Answers: migration.RolledBack{},
	Run: func(on workedOn, _ Input) (Answer, error) {
		rolled, err := on.Engine.rollBack(on.Path)
		if err != nil {
			return Answer{}, err
		}
		return Answer{RolledBack: &rolled}, nil
	},
}
