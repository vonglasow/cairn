// Package cli assembles the command tree. Every command is a thin shell: it
// parses arguments, calls one thing, and prints. Anything worth testing lives
// deeper, so the commands themselves stay boring.
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

// stamps is how this surface asks which build it is. It is a variable so a
// scenario can decide what the binary claims to be: the answer is read at the
// moment the tree is built, and a scenario drives Main, not the linker.
var stamps = build.Read

// NewRootCmd builds the whole command tree. It takes no globals so a test can
// run the real CLI in-process, against its own state.
func NewRootCmd() *cobra.Command {
	b := stamps()
	root := &cobra.Command{
		Use:   "cairn",
		Short: "Cairn — the markers that show the path you are on",
		// Without this, a root command holding only subcommands accepts any
		// stray argument, prints its help and exits zero — so a typo looks like
		// a success.
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		// Setting Version is what gives cobra its --version flag. The template
		// makes it print the full build block rather than cobra's one-liner,
		// from the same BuildInfo everything else reads.
		Version: b.Version,
		// Cobra skips Args validation on a command with nothing to run, so the
		// root needs a RunE for NoArgs to bite. Bare `cairn` prints its help.
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.SetVersionTemplate(b.String())
	// `task show -1` is somebody naming a number, not misusing a flag. Cobra
	// reads the leading minus as a shorthand and answers "unknown shorthand
	// flag: '1' in -1", which sends a person looking through flags they never
	// typed. Said plainly instead, and left as a usage error either way: the
	// number does not exist, whichever way it was meant.
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		typed, ok := leadingMinus(err)
		if !ok {
			return engine.Misread(err)
		}
		if _, number := strconv.Atoi(typed); number == nil {
			return engine.Misread(fmt.Errorf("%q: a task number is expected", typed))
		}
		// Anything else that starts with a minus — a length to take back —
		// is a value, and the command line reads values that look like flags
		// only after `--`. Said here, where the problem is: the description
		// of the argument is read by agents too, and they have no `--`.
		return engine.Misread(fmt.Errorf("%q: a value that starts with - is written after --, as in -- %s", typed, typed))
	})
	// Asked for on any command, so `task new --json` and `task list --json` read
	// the same way. Declared on the root because the answer's shape is a
	// property of the whole conversation, not of one subcommand.
	root.PersistentFlags().Bool(jsonFlag, false, "answer as JSON, on standard output")
	root.AddCommand(newActionGroups()...)
	root.AddCommand(newMCPCmd())
	root.AddCommand(newTuiCmd())
	root.AddCommand(newGuiCmd())
	markMisreads(root)
	return root
}

// markMisreads marks what every command's arguments refuse as not having been
// understood. Cobra says so in sentences of its own, and a failure is
// otherwise answered as one nobody foresaw — so the shell was told "could not
// do it" for a command it could not even read.
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

// leadingMinus pulls what somebody actually typed out of the parser's
// complaint about a shorthand that is really a value starting with a minus —
// `-30m` read as the flags 3, 0 and m. Read off the parser's own error rather
// than out of its sentence, which an upgrade is free to reword.
func leadingMinus(err error) (string, bool) {
	var unknown *pflag.NotExistError
	if !errors.As(err, &unknown) || unknown.GetSpecifiedShortnames() == "" {
		return "", false
	}
	return "-" + unknown.GetSpecifiedShortnames(), true
}

// saidToAPerson writes a refusal the way fang does, with two things taken
// back. fang's style capitalises the text, and a refusal opens on the value
// that was typed: "abc" was quoted back as "Abc", a tab as \T — something
// nobody wrote, and not what the JSON says. And what reaches the screen is
// shown, never obeyed: a value is quoted with its escapes written out, and
// whatever a message carries that a terminal would act on is replaced
// (domain.Shown) rather than handed to it.
func saidToAPerson(w io.Writer, styles fang.Styles, err error) {
	styles.ErrorText = styles.ErrorText.UnsetTransform()
	fang.DefaultErrorHandler(w, styles, shownError{err})
}

// shownError is an error as it is safe to print, still answering to what it
// wraps.
type shownError struct{ err error }

func (s shownError) Error() string { return domain.Shown(s.err.Error()) }
func (s shownError) Unwrap() error { return s.err }

// askedForJSON says whether the arguments ask for JSON, spelled any of the ways
// the flag itself accepts: `--json`, or `--json=` with a value read as a yes.
func askedForJSON(args []string) bool {
	for _, arg := range args {
		// What follows `--` is a value, whatever it reads like: a title
		// that happens to read "--json" asks for nothing.
		if arg == "--" {
			return false
		}
		if arg == "--"+jsonFlag {
			return true
		}
		// Any spelling but a plain no is a request: one the flag cannot read
		// still says the caller wanted a machine's answer.
		if said, ok := strings.CutPrefix(arg, "--"+jsonFlag+"="); ok {
			if yes, err := strconv.ParseBool(said); err != nil || yes {
				return true
			}
		}
	}
	return false
}

// Main runs the command line and returns the code the process should exit with.
//
// It takes its arguments and its three streams rather than reading os.Args,
// os.Stdin, os.Stdout and os.Stderr, so a test runs exactly what a user runs —
// fang's error rendering included — without spawning a binary and without a
// global to restore afterwards. `cmd/cairn` is then the one line that calls
// os.Exit.
func Main(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	root := NewRootCmd()
	root.SetIn(stdin)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs(args)

	// Asked for JSON, the whole answer goes to standard output as JSON, whether
	// it worked or not — a machine reads one stream, and a refusal it cannot
	// parse is a refusal it cannot act on. Read from the arguments rather than
	// from the parsed flags, because a command that failed to parse never got
	// as far as its flags.
	asMachine := askedForJSON(args)
	// Whether the flags could be read at all. When they could, what they say
	// decides; when they could not, the words are all there is to go on.
	flagsRead := true
	refusing := root.FlagErrorFunc()
	root.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		flagsRead = false
		return refusing(c, err)
	})

	// fang writes the error itself, to the command's error stream, and returns
	// it so the caller can still decide what to exit with.
	err := fang.Execute(ctx, root, fang.WithErrorHandler(
		func(w io.Writer, styles fang.Styles, err error) {
			// Once the flags were read, what they say decides: "--json"
			// given as the value of another flag asked for nothing. The words
			// are only looked at when the flags could not be read at all.
			machine := asMachine
			if c, _, found := root.Find(args); found == nil && flagsRead && c.Flags().Parsed() {
				machine = wantsJSON(c)
			}
			// A refusal that cannot be written as JSON is said where a
			// person reads instead: an empty stream tells a machine nothing.
			if machine && writeRefusal(stdout, err) == nil {
				return
			}
			saidToAPerson(w, styles, err)
		}))
	return codeFor(err)
}
