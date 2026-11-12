// Package cli assembles the command tree.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	rdebug "runtime/debug"
	"slices"
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
func NewRootCmd() *cobra.Command { return newRootCmdFor(home(), stamps()) }

func newRootCmdFor(h engine.Home, b build.Info) *cobra.Command {
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
	root.PersistentFlags().Bool(debugFlag, false, "write down what this run does, beside the store, as CAIRN_DEBUG=1 does")
	root.PersistentFlags().String(workspaceFlag, "", "the workspace this command uses, whatever the environment says")
	root.PersistentFlags().String(storeFlag, "", "the store this command uses, whatever the environment says")
	root.AddCommand(newActionGroups(engine.PagesFrom(h))...)
	root.AddCommand(newMCPCmd())
	root.AddCommand(newTuiCmd())
	root.AddCommand(newGuiCmd())
	root.AddCommand(newFixturesCmd())
	root.AddCommand(newUpgradeCmd())
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

func askedFor(flag string, args []string) bool {
	for _, arg := range args {
		if arg == "--" {
			return false
		}
		if arg == "--"+flag {
			return true
		}
		if said, ok := strings.CutPrefix(arg, "--"+flag+"="); ok {
			if yes, err := strconv.ParseBool(said); err != nil || yes {
				return true
			}
		}
	}
	return false
}

func valueOf(flag string, args []string) string {
	for i, arg := range args {
		if arg == "--" {
			return ""
		}
		if arg == "--"+flag && i+1 < len(args) {
			return args[i+1]
		}
		if said, ok := strings.CutPrefix(arg, "--"+flag+"="); ok {
			return said
		}
	}
	return ""
}

// Main runs the command line and returns the code the process should exit with.
func Main(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return MainIn(ctx, home(), args, stdin, stdout, stderr)
}

// MainIn runs the command line the way Main does, in the home given rather than the process's.
func MainIn(ctx context.Context, h engine.Home, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	ctx = withHome(ctx, h)
	h.Workspace, h.Store = valueOf(workspaceFlag, args), valueOf(storeFlag, args)
	root := newRootCmdFor(h, stampsIn(ctx))
	root.SetIn(stdin)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs(bareJoined(args))

	asMachine := askedFor(jsonFlag, args)
	sayIgnored(h, stderr)
	flagsRead := true
	refusing := root.FlagErrorFunc()
	root.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		flagsRead = false
		return refusing(c, err)
	})

	if written := debugging(root, h, args, stderr); written != nil {
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

func sayIgnored(h engine.Home, stderr io.Writer) {
	for _, ignored := range engine.Ignored(h) {
		_, _ = fmt.Fprintf(stderr, "config: %s\n", ignored)
	}
}

type notedKey struct{}

type stampsKey struct{}

func withStamps(ctx context.Context, read func() build.Info) context.Context {
	return context.WithValue(ctx, stampsKey{}, read)
}

func stampsIn(ctx context.Context) build.Info {
	if read, given := ctx.Value(stampsKey{}).(func() build.Info); given {
		return read()
	}
	return stamps()
}

func notedIn(ctx context.Context) *debug.Log {
	written, _ := ctx.Value(notedKey{}).(*debug.Log)
	return written
}

func debugging(root *cobra.Command, h engine.Home, args []string, stderr io.Writer) *debug.Log {
	if !askedFor(debugFlag, args) && !engine.Debugging(h) {
		return nil
	}
	at, err := engine.Locate(h)
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

func bareJoined(args []string) []string {
	var bare []string
	for _, op := range engine.All() {
		for _, arg := range op.Takes {
			if arg.Bare != "" {
				bare = append(bare, "--"+arg.Name)
			}
		}
	}
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if spelled, ok := strings.CutPrefix(args[i], "--description"); ok && (spelled == "" || strings.HasPrefix(spelled, "=")) {
			args[i] = "--desc" + spelled
		}
		if slices.Contains(bare, args[i]) && i+1 < len(args) {
			if _, err := strconv.Atoi(args[i+1]); err == nil {
				out = append(out, args[i]+"="+args[i+1])
				i++
				continue
			}
		}
		out = append(out, args[i])
	}
	return out
}
