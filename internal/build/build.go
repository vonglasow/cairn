// Package build says which build this is.
package build

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"text/tabwriter"
)

var (
	version   = ""
	commit    = ""
	built     = ""
	committed = ""
)

const shortCommitLen = 12

// Info is everything a build reports about itself.
type Info struct {
	Version    string
	Commit     string
	Built      string
	CommitDate string
	Go         string
	Platform   string
}

// Read answers what this build is, from the stamps the linker left behind.
func Read() Info { return Stamped(version, commit, built, committed) }

// Stamped answers what a build carrying these four stamps would report, falling back through the VCS information Go embeds automatically, then to placeholders.
func Stamped(version, commit, built, committed string) Info {
	b := Info{
		Version:    version,
		Commit:     commit,
		Built:      built,
		CommitDate: committed,
		Go:         runtime.Version(),
		Platform:   runtime.GOOS + "/" + runtime.GOARCH,
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		b = applyVCS(b, info.Settings)
	}

	b.Commit = b.Commit[:min(len(b.Commit), shortCommitLen)]

	b.Version = orElse(b.Version, "dev")
	b.Commit = orElse(b.Commit, "none")
	b.Built = orElse(b.Built, "unknown")
	b.CommitDate = orElse(b.CommitDate, "unknown")
	return b
}

func applyVCS(b Info, settings []debug.BuildSetting) Info {
	for _, s := range settings {
		switch s.Key {
		case "vcs.revision":
			if b.Commit == "" {
				b.Commit = s.Value
			}
		case "vcs.time":
			if b.CommitDate == "" {
				b.CommitDate = s.Value
			}
		}
	}
	return b
}

// String renders the build info as aligned key/value lines, the name and version first.
func (b Info) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "cairn %s\n", b.Version)

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
