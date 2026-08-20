// Package cli assembles the command tree. Every command is a thin shell: it
// parses arguments, calls one thing, and prints. Anything worth testing lives
// deeper, so the commands themselves stay boring.
package cli

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"runtime/debug"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

// Build stamps, injected by the linker (see the Makefile). They stay vars, not
// consts, precisely so -ldflags -X can reach them.
var (
	version = ""
	commit  = ""
	built   = ""
)

// shortCommitLen is how much of a SHA is enough to identify a commit while
// staying readable.
const shortCommitLen = 12

// BuildInfo is everything `version` reports. Three fields come from the linker,
// one from the VCS stamp Go embeds on its own, two from the runtime — so a
// binary built without any stamping still tells the truth about itself instead
// of pretending.
type BuildInfo struct {
	Version    string
	Commit     string
	Built      string
	CommitDate string
	Go         string
	Platform   string
}

// Build reads the build stamps, falling back through the VCS information Go
// embeds automatically, then to placeholders. `go run` gets neither stamps nor
// VCS data, and must still print something honest.
func Build() BuildInfo {
	b := BuildInfo{
		Version:    version,
		Commit:     commit,
		Built:      built,
		CommitDate: "",
		Go:         runtime.Version(),
		Platform:   runtime.GOOS + "/" + runtime.GOARCH,
	}

	// `go build` stamps the revision and its date without being asked; `go run`
	// does not. Use them to fill whatever the linker did not.
	if info, ok := debug.ReadBuildInfo(); ok {
		b = applyVCS(b, info.Settings)
	}

	// No conditional here on purpose: `if len > n` and `if len >= n` produce the
	// same string for a commit of exactly n characters, so the boundary would be
	// untestable and any mutant on it would survive forever.
	b.Commit = b.Commit[:min(len(b.Commit), shortCommitLen)]

	b.Version = orElse(b.Version, "dev")
	b.Commit = orElse(b.Commit, "none")
	b.Built = orElse(b.Built, "unknown")
	b.CommitDate = orElse(b.CommitDate, "unknown")
	return b
}

// applyVCS fills whatever the linker left empty with the revision and date Go
// embeds on its own. Split out of Build so it can be tested without needing a
// binary that was actually stamped.
func applyVCS(b BuildInfo, settings []debug.BuildSetting) BuildInfo {
	for _, s := range settings {
		switch s.Key {
		case "vcs.revision":
			// The linker wins: an explicit stamp is more precise than whatever
			// the working tree happened to be.
			if b.Commit == "" {
				b.Commit = s.Value
			}
		case "vcs.time":
			b.CommitDate = s.Value
		}
	}
	return b
}

// String renders the build info as aligned key/value lines, the name and
// version first.
func (b BuildInfo) String() string {
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, "cairn %s\n", b.Version)

	tw := tabwriter.NewWriter(&sb, 0, 2, 2, ' ', 0)
	_, _ = fmt.Fprintf(tw, "commit\t%s\n", b.Commit)
	_, _ = fmt.Fprintf(tw, "built\t%s\n", b.Built)
	_, _ = fmt.Fprintf(tw, "commit date\t%s\n", b.CommitDate)
	_, _ = fmt.Fprintf(tw, "go\t%s\n", b.Go)
	_, _ = fmt.Fprintf(tw, "platform\t%s\n", b.Platform)
	_ = tw.Flush()
	return sb.String()
}

func orElse(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

// NewRootCmd builds the whole command tree. It takes no globals so a test can
// run the real CLI in-process, against its own state.
func NewRootCmd() *cobra.Command {
	b := Build()
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
	// Asked for on any command, so `task add --json` and `task list --json` read
	// the same way. Declared on the root because the answer's shape is a
	// property of the whole conversation, not of one subcommand.
	root.PersistentFlags().Bool(jsonFlag, false, "answer as JSON, on standard output")
	root.AddCommand(newTaskCmd(), newBackupCmd(), newMigrationCmd())
	return root
}

// Main runs the command line and returns the code the process should exit with.
//
// It takes its arguments and its two streams rather than reading os.Args,
// os.Stdout and os.Stderr, so a test runs exactly what a user runs — fang's
// error rendering included — without spawning a binary and without a global to
// restore afterwards. `cmd/cairn` is then the one line that calls os.Exit.
func Main(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	root := NewRootCmd()
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
