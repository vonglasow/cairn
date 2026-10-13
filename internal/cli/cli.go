// Package cli assembles the command tree.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"cairn/internal/build"
	"cairn/internal/domain"
	"cairn/internal/engine"
)

var stamps = build.Read

// NewRootCmd builds the whole command tree.
func NewRootCmd() *cobra.Command {
	b := stamps()
	root := &cobra.Command{
		Use:           "cairn",
		Short:         "Cairn — the markers that show the path you are on",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       b.Version,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.SetVersionTemplate(b.String())
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		typed, ok := leadingMinus(err)
		if !ok {
			return engine.Misread(err)
		}
		if _, number := strconv.Atoi(typed); number == nil {
			return engine.Misread(fmt.Errorf("%q: a task number is expected", typed))
		}
		return engine.Misread(fmt.Errorf("%q: a value that starts with - is written after --, as in -- %s", typed, typed))
	})
	root.PersistentFlags().Bool(jsonFlag, false, "answer as JSON, on standard output")
	root.AddCommand(newActionGroups()...)
	root.AddCommand(newMCPCmd())
	root.AddCommand(newTuiCmd())
	root.AddCommand(newGuiCmd())
	markMisreads(root)
	return root
}

func markMisreads(c *cobra.Command) {
	if args := c.Args; args != nil {
		c.Args = func(cmd *cobra.Command, given []string) error {
			if err := args(cmd, given); err != nil {
				return engine.Misread(err)
			}
			return nil
		}
	}
	for _, sub := range c.Commands() {
		markMisreads(sub)
	}
}

func leadingMinus(err error) (string, bool) {
	var unknown *pflag.NotExistError
	if !errors.As(err, &unknown) || unknown.GetSpecifiedShortnames() == "" {
		return "", false
	}
	return "-" + unknown.GetSpecifiedShortnames(), true
}

func saidToAPerson(w io.Writer, styles fang.Styles, err error) {
	styles.ErrorText = styles.ErrorText.UnsetTransform()
	fang.DefaultErrorHandler(w, styles, shownError{err})
}

type shownError struct{ err error }

func (s shownError) Error() string { return domain.Shown(s.err.Error()) }
func (s shownError) Unwrap() error { return s.err }

func askedForJSON(args []string) bool {
	for _, arg := range args {
		if arg == "--" {
			return false
		}
		if arg == "--"+jsonFlag {
			return true
		}
		if said, ok := strings.CutPrefix(arg, "--"+jsonFlag+"="); ok {
			if yes, err := strconv.ParseBool(said); err != nil || yes {
				return true
			}
		}
	}
	return false
}

// Main runs the command line and returns the code the process should exit with.
func Main(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	root := NewRootCmd()
	root.SetIn(stdin)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs(args)

	asMachine := askedForJSON(args)
	flagsRead := true
	refusing := root.FlagErrorFunc()
	root.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		flagsRead = false
		return refusing(c, err)
	})

	err := fang.Execute(ctx, root, fang.WithErrorHandler(
		func(w io.Writer, styles fang.Styles, err error) {
			machine := asMachine
			if c, _, found := root.Find(args); found == nil && flagsRead && c.Flags().Parsed() {
				machine = wantsJSON(c)
			}
			if machine && writeRefusal(stdout, err) == nil {
				return
			}
			saidToAPerson(w, styles, err)
		}))
	return codeFor(err)
}
