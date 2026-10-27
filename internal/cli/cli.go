// Package cli assembles the command tree.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	rdebug "runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"cairn/internal/build"
	"cairn/internal/debug"
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
	root.AddCommand(newActionGroups(engine.PagesFrom(home()))...)
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
	said, command, ok := engine.ToType(err)
	if !ok {
		fang.DefaultErrorHandler(w, styles, shownError{err})
		return
	}
	fang.DefaultErrorHandler(w, styles, shownError{typedApart{said, err}})
	_, _ = fmt.Fprintf(w, "  To go on, run:\n\n    %s\n\n", domain.Shown(command))
}

type typedApart struct {
	said string
	err  error
}

func (t typedApart) Error() string { return t.said }
func (t typedApart) Unwrap() error { return t.err }

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

	if written := debugging(root, args, stderr); written != nil {
		engine.NoteTo(written)
		ctx = context.WithValue(ctx, notedKey{}, written)
		defer func() {
			engine.NoteTo(nil)
			written.Leave(stderr)
		}()
		defer func() {
			if stopped := recover(); stopped != nil {
				written.Panicked(stopped, rdebug.Stack())
				panic(stopped)
			}
		}()
	}

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

type notedKey struct{}

func notedIn(ctx context.Context) *debug.Log {
	written, _ := ctx.Value(notedKey{}).(*debug.Log)
	return written
}

func debugging(root *cobra.Command, args []string, stderr io.Writer) *debug.Log {
	if !engine.Debugging(home()) {
		return nil
	}
	at, err := engine.Locate(home())
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "debug: cannot write down: %v\n", err)
		return nil
	}
	var command []string
	if c, _, err := root.Find(args); err == nil && c != root {
		command = strings.Fields(strings.TrimPrefix(c.CommandPath(), root.Name()+" "))
	}
	surface := strings.Join(command, " ")
	if surface == "" {
		surface = root.Name()
	}
	written, err := debug.Open(at.Debug(), surface, time.Now())
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "debug: cannot write down in %s: %v\n", at.Debug(), err)
		return nil
	}
	for _, fact := range engine.DebugHeader(at, written.Args(command, args)) {
		written.Fact(fact[0], fact[1])
	}
	return written
}
