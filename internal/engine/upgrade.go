package engine

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"cairn/internal/upgrade"
)

// What upgrading leaves to whoever keeps the cairn that was run.
var (
	errBuiltFromSource = errors.New("this cairn was built from its source, not taken from a release: build the newer one the same way")
	errKeptByHomebrew  = errors.New("this cairn was installed by Homebrew: run brew upgrade cairn")
)

// upgraded says which cairn was run, which one is in its place now, and where that place is.
type upgraded struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Binary string `json:"binary"`
}

// String says what upgrading did, in a line.
func (u upgraded) String() string {
	if u.From == u.To {
		return "cairn " + u.From + " is the newest already, in " + u.Binary + "\n"
	}
	return "cairn went from " + u.From + " to " + u.To + ", in " + u.Binary + "\n\n" +
		"its plugin is at the tag " + u.To + "; " + installText(installFromTag(u.To))
}

// Upgrade puts the newest release of cairn, made for platform, in the place of binary, the cairn that was run at version; one already the newest is left as it is.
func Upgrade(version, binary, platform string) (upgraded, error) {
	running, ok := releaseOf(version)
	if !ok {
		return upgraded{}, fmt.Errorf("%s: %w", version, errBuiltFromSource)
	}
	if strings.Contains(binary, string(filepath.Separator)+"Cellar"+string(filepath.Separator)) {
		return upgraded{}, fmt.Errorf("%s: %w", binary, errKeptByHomebrew)
	}
	if err := upgrade.Writable(binary); err != nil {
		return upgraded{}, err
	}
	release, err := upgrade.Newest()
	if err != nil {
		return upgraded{}, err
	}
	newest, ok := releaseOf(release.Tag)
	if !ok {
		return upgraded{}, fmt.Errorf("%w: the newest is tagged %q", upgrade.ErrOutOfReach, release.Tag)
	}
	if !newer(newest, running) {
		return upgraded{From: version, To: version, Binary: binary}, nil
	}
	if err := release.Install(binary, platform); err != nil {
		return upgraded{}, err
	}
	return upgraded{From: version, To: release.Tag, Binary: binary}, nil
}

func releaseOf(said string) ([3]int, bool) {
	var version [3]int
	parts := strings.Split(strings.TrimPrefix(said, "v"), ".")
	if !strings.HasPrefix(said, "v") || len(parts) != len(version) {
		return version, false
	}
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 || strconv.Itoa(n) != part {
			return version, false
		}
		version[i] = n
	}
	return version, true
}

func newer(than, version [3]int) bool {
	return slices.Compare(than[:], version[:]) > 0
}
