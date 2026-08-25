// Package build says which build this is. It knows nobody: an identity that
// knew how it was displayed would drag a terminal into every test that reads
// it, and an identity owned by one surface would be an identity the other has
// to borrow.
//
// It sits here rather than in a surface because both of them have to answer the
// question. `cairn --version` prints it for a person; an MCP server hands it to
// a client at the greeting. Were it owned by the command line, the server would
// be reading another surface's private state to say what it is.
package build

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"text/tabwriter"
)

// The stamps, injected by the linker (see PKG in the Makefile). They stay vars,
// not consts, precisely so -ldflags -X can reach them.
var (
	version = ""
	commit  = ""
	built   = ""
)

// shortCommitLen is how much of a SHA is enough to identify a commit while
// staying readable.
const shortCommitLen = 12

// Info is everything a build reports about itself. Three fields come from the
// linker, one from the VCS stamp Go embeds on its own, two from the runtime —
// so a binary built without any stamping still tells the truth about itself
// instead of pretending.
type Info struct {
	Version    string
	Commit     string
	Built      string
	CommitDate string
	Go         string
	Platform   string
}

// Read answers what this build is, from the stamps the linker left behind.
func Read() Info { return Stamped(version, commit, built) }

// Stamped answers what a build carrying these three stamps would report,
// falling back through the VCS information Go embeds automatically, then to
// placeholders. `go run` gets neither stamps nor VCS data, and must still print
// something honest.
//
// Read calls it with what the linker left behind. It takes the three rather
// than reading them so that what a build says can be worked out without being
// that build.
func Stamped(version, commit, built string) Info {
	b := Info{
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
// embeds on its own. Split out of Stamped so it can be tested without needing a
// binary that was actually stamped.
func applyVCS(b Info, settings []debug.BuildSetting) Info {
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
func (b Info) String() string {
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
