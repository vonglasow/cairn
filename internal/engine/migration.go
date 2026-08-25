package engine

// The migration operations: what has been done to the shape of the store,
// what a change will do before it is made, and how to apply or roll one
// back, described once so both surfaces build the same command from the
// same facts. They open the store whatever shape it is in — changing that
// shape is their job.

import (
	"errors"
)

// MigrationList names every change to the store, applied or waiting.
var MigrationList = Operation{
	Thing: "migration",
	Verb:  "list",
	Short: "Name every change to the store, applied or waiting",
	Does:  "Name every change to the store, applied or waiting.",
	Level: Read,
	Needs: AStoreBehind,
	Run: func(on On, _ Input) (Answer, error) {
		all, err := on.Engine.Migrations()
		if err != nil {
			return Answer{}, err
		}
		held, err := on.Engine.Held()
		if err != nil {
			return Answer{}, err
		}
		return Answer{Changes: all, Holds: &held}, nil
	},
}

// MigrationShow says what a change does, and what undoing it does.
var MigrationShow = Operation{
	Thing: "migration",
	Verb:  "show",
	Short: "Say what a change does, and what undoing it does",
	Does:  "Say what a change does, and what undoing it does.",
	Level: Read,
	Needs: AStoreBehind,
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Required: true, Of: "a migration", Means: "the number of the change"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, _ := in.Whole("number")
		one, err := on.Engine.Change(n)
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
	Level: Write,
	Needs: AStoreBehind,
	Takes: []Argument{
		{Name: "number", Kind: Number, Positional: true, Of: "a migration", Means: "the number of the change"},
	},
	Run: func(on On, in Input) (Answer, error) {
		n, named := in.Whole("number")
		if !named {
			made, err := on.Engine.ApplyAll(on.Path)
			if errors.Is(err, ErrNothingWaiting) {
				return Answer{Applied: []string{}}, nil
			}
			if err != nil {
				return Answer{}, err
			}
			return Answer{Applied: made}, nil
		}
		one, err := on.Engine.Apply(on.Path, n)
		if err != nil {
			return Answer{}, err
		}
		return Answer{Applied: []string{one}}, nil
	},
}

// MigrationRollback unmakes the last change applied to the store.
var MigrationRollback = Operation{
	Thing: "migration",
	Verb:  "rollback",
	Short: "Unmake the last change applied to the store",
	Does:  "Unmake the last change applied to the store.",
	Level: Full,
	Needs: AStoreBehind,
	Run: func(on On, _ Input) (Answer, error) {
		rolled, err := on.Engine.RollBack(on.Path)
		if err != nil {
			return Answer{}, err
		}
		return Answer{RolledBack: &rolled}, nil
	},
}
