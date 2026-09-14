// Package cli assembles the command tree. Every command is a thin shell: it
// parses arguments, calls one thing, and prints. Anything worth testing lives
// deeper, so the commands themselves stay boring.
package cli

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"slices"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"

	"cairn/internal/build"
)

// stamps is how this surface asks which build it is. It is a variable so a
// scenario can decide what the binary claims to be: the answer is read at the
// moment the tree is built, and a scenario drives Main, not the linker.
var stamps = build.Read

// NewRootCmd builds the whole command tree. It takes no globals so a test can
// run the real CLI in-process, against its own state.
// NewRootCmd builds the command tree cairn answers with.
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
		if typed, ok := negativeNumber(err); ok {
			return fmt.Errorf("%q: a task number is expected", typed)
		}
		return err
	})
	// Asked for on any command, so `task add --json` and `task list --json` read
	// the same way. Declared on the root because the answer's shape is a
	// property of the whole conversation, not of one subcommand.
	root.PersistentFlags().Bool(jsonFlag, false, "answer as JSON, on standard output")
	root.AddCommand(newActionGroups()...)
	root.AddCommand(newMCPCmd())
	root.AddCommand(newBoardCmd())
	root.AddCommand(newGuiCmd())
	return root
}

// numberLikeFlag matches cobra's complaint about a shorthand that is really a
// negative number: "unknown shorthand flag: '1' in -12".
var numberLikeFlag = regexp.MustCompile(`unknown shorthand flag: '\d' in (-\d+)`)

// negativeNumber pulls what somebody actually typed out of that complaint.
func negativeNumber(err error) (string, bool) {
	found := numberLikeFlag.FindStringSubmatch(err.Error())
	if found == nil {
		return "", false
	}
	return found[1], true
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
	asMachine := slices.Contains(args, "--"+jsonFlag)

	// fang writes the error itself, to the command's error stream, and returns
	// it so the caller can still decide what to exit with.
	err := fang.Execute(ctx, root, fang.WithErrorHandler(
		func(w io.Writer, styles fang.Styles, err error) {
			if asMachine {
				_ = writeRefusal(stdout, err)
				return
			}
			fang.DefaultErrorHandler(w, styles, err)
		}))
	return codeFor(err)
}
