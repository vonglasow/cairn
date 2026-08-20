package cli

// The `backup` commands: a backup of the store is made, listed, read, deleted and
// put back like anything else cairn holds. They reach the backups without
// opening the store, because a backup is most needed when the store cannot be
// opened at all.

import (
	"fmt"
	"os"

	"cairn/internal/domain"
	"cairn/internal/engine"

	"github.com/spf13/cobra"
)

// storePath is where the store lives, according to the environment and the
// default place. Read here so the backup commands can work without opening the
// store — a backup is most needed when the store cannot be opened.
func storePath() (string, error) {
	path, err := engine.StoreAddress(os.LookupEnv, os.UserHomeDir)
	if err != nil {
		return "", fmt.Errorf("%w: %w", engine.ErrStore, err)
	}
	return path, nil
}

// backups names the backups kept beside the store.
func backups() (*engine.Backups, error) {
	path, err := storePath()
	if err != nil {
		return nil, err
	}
	return engine.BackupsAt(path), nil
}

// newBackupCmd builds the `backup` group: a backup is a thing in its own right,
// so it is added, listed, read, deleted — the same verbs as anything else.
func newBackupCmd() *cobra.Command {
	group := &cobra.Command{
		Use:   "backup",
		Short: "Make, list, read and put back backups of the store",
		Args:  noneAtAll(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("backup needs a command — usage: %s", cmd.UseLine())
		},
	}
	group.AddCommand(newBackupAddCmd(), newBackupListCmd(), newBackupShowCmd(),
		newBackupDeleteCmd(), newBackupRestoreCmd())
	return group
}

func newBackupAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Back up the store as it stands",
		Args:  noneAtAll(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			vault, err := backups()
			if err != nil {
				return err
			}
			// Counted before backing up, not after: what is announced is what was
			// taken. A file that answers nothing is not a store, and backing up it
			// would answer "backup made, 0 tasks" about somebody else's file.
			held, err := vault.HoldsStore()
			if err != nil {
				return err
			}
			made, err := vault.Make()
			if err != nil {
				return err
			}
			return writeBackup(cmd, made, held)
		},
	}
}

func newBackupListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List every backup",
		Args:  noneAtAll(),
		RunE: func(cmd *cobra.Command, _ []string) error {
			vault, err := backups()
			if err != nil {
				return err
			}
			all, err := vault.All()
			if err != nil {
				return err
			}
			return writeBackups(cmd, all)
		},
	}
}

func newBackupShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <number>",
		Short: "Say when a backup was made and what it holds",
		Args:  exactly(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := number(args[0])
			if err != nil {
				return err
			}
			vault, err := backups()
			if err != nil {
				return err
			}
			one, err := vault.Find(n)
			if err != nil {
				return err
			}
			held, err := vault.Holds(one)
			if err != nil {
				return err
			}
			return writeBackup(cmd, one, held)
		},
	}
}

func newBackupDeleteCmd() *cobra.Command {
	var confirmed bool
	cmd := &cobra.Command{
		Use:   "delete <number>",
		Short: "Delete a backup",
		Args:  exactly(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := number(args[0])
			if err != nil {
				return err
			}
			if !confirmed {
				return fmt.Errorf("%w: add --yes to confirm", errNotConfirmed)
			}
			vault, err := backups()
			if err != nil {
				return err
			}
			if err := vault.Delete(n); err != nil {
				return err
			}
			return writeDone(cmd, "deleted", n)
		},
	}
	cmd.Flags().BoolVar(&confirmed, "yes", false, "confirm the deletion")
	return cmd
}

func newBackupRestoreCmd() *cobra.Command {
	var confirmed bool
	cmd := &cobra.Command{
		Use:   "restore [number]",
		Short: "Put a backup back, the most recent one unless told otherwise",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vault, err := backups()
			if err != nil {
				return err
			}
			one, err := whichBackup(vault, args)
			if err != nil {
				return err
			}
			if !confirmed {
				return fmt.Errorf("%w: add --yes to confirm", errNotConfirmed)
			}
			if err := vault.Restore(one.Number); err != nil {
				return err
			}
			return writeDone(cmd, "put back", one.Number)
		},
	}
	cmd.Flags().BoolVar(&confirmed, "yes", false, "confirm putting the backup back")
	return cmd
}

// whichBackup is the backup a restore was aimed at: the one named, or the most
// recent when none was.
func whichBackup(vault *engine.Backups, args []string) (domain.Backup, error) {
	if len(args) == 0 {
		return vault.Latest()
	}
	n, err := number(args[0])
	if err != nil {
		return domain.Backup{}, err
	}
	return vault.Find(n)
}
