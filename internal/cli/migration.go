package cli

// The `migration` commands: what has been done to the shape of the store, what
// a change will do before it is made, and how to apply or roll one back. They
// open the store whatever shape it is in — changing that shape is their job.

import (
	"errors"
	"fmt"

	"cairn/internal/engine"

	"github.com/spf13/cobra"
)

// newMigrationCmd builds the `migration` group.
func newMigrationCmd() *cobra.Command {
	group := &cobra.Command{
		Use:   "migration",
		Short: "Read and undo what has been done to the shape of the store",
		Args:  noneAtAll(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("migration needs a command — usage: %s", cmd.UseLine())
		},
	}
	group.AddCommand(newMigrationListCmd(), newMigrationShowCmd(),
		newMigrationApplyCmd(), newMigrationRollbackCmd())
	return group
}

func newMigrationListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Name every change to the store, applied or waiting",
		Args:  noneAtAll(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			eng, _, err := openToChange()
			if err != nil {
				return err
			}
			defer func() { _ = eng.Close() }()

			all, err := eng.Migrations()
			if err != nil {
				return err
			}
			held, err := eng.Held()
			if err != nil {
				return err
			}
			return writeChanges(cmd, all, held)
		},
	}
}

func newMigrationShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <number>",
		Short: "Say what a change does, and what undoing it does",
		Args:  exactly(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := number(args[0])
			if err != nil {
				return err
			}
			eng, _, err := openToChange()
			if err != nil {
				return err
			}
			defer func() { _ = eng.Close() }()

			one, err := eng.Change(n)
			if err != nil {
				return err
			}
			return writeChange(cmd, one)
		},
	}
}

func newMigrationApplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "apply [number]",
		Short: "Make the changes the store is missing, or one of them",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			eng, path, err := openToChange()
			if err != nil {
				return err
			}
			defer func() { _ = eng.Close() }()

			made, err := applyWhat(eng, path, args)
			if errors.Is(err, engine.ErrNothingWaiting) {
				return writeNothingWaiting(cmd)
			}
			if err != nil {
				return err
			}
			return writeApplied(cmd, made)
		},
	}
}

// applyWhat makes what the arguments asked for: one change, or everything
// waiting when none was named.
func applyWhat(eng *engine.Engine, path string, args []string) ([]string, error) {
	if len(args) == 0 {
		return eng.ApplyAll(path)
	}
	n, err := number(args[0])
	if err != nil {
		return nil, err
	}
	one, err := eng.Apply(path, n)
	if err != nil {
		return nil, err
	}
	return []string{one}, nil
}

func newMigrationRollbackCmd() *cobra.Command {
	var confirmed bool
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Unmake the last change applied to the store",
		Args:  noneAtAll(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirmed {
				return fmt.Errorf("%w: add --yes to confirm", errNotConfirmed)
			}
			eng, path, err := openToChange()
			if err != nil {
				return err
			}
			defer func() { _ = eng.Close() }()

			rolled, err := eng.RollBack(path)
			if err != nil {
				return err
			}
			return writeRolledBack(cmd, rolled)
		},
	}
	cmd.Flags().BoolVar(&confirmed, "yes", false, "confirm rolling the change back")
	return cmd
}

// openToChange opens the store whatever shape it is in, for the commands whose
// job is to change that shape, and hands back where it lives so a backup can be
// set aside first.
func openToChange() (*engine.Engine, string, error) {
	path, err := storePath()
	if err != nil {
		return nil, "", err
	}
	eng, err := engine.OpenBehind(path)
	if err != nil {
		return nil, "", err
	}
	return eng, path, nil
}
