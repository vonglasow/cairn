package engine

import (
	"errors"
	"testing"

	"cairn/internal/upgrade"
)

func TestNewer_readsEachNumberAsANumber(t *testing.T) {
	for _, c := range []struct {
		tag, version string
		newer        bool
	}{
		{"v0.10.0", "v0.9.0", true},
		{"v0.9.1", "v0.9.0", true},
		{"v1.0.0", "v0.99.99", true},
		{"v0.10.0", "v0.10.0", false},
		{"v0.9.0", "v0.10.0", false},
		{"v0.10.0", "v0.10.1", false},
		{"v0.10.0", "v1.0.0", false},
	} {
		if got := newer(versionOf(c.tag), versionOf(c.version)); got != c.newer {
			t.Errorf("%s newer than %s: %v, want %v", c.tag, c.version, got, c.newer)
		}
	}
}

func TestVersionOf_takesAReleaseAndNothingElse(t *testing.T) {
	for said, release := range map[string]bool{
		"v0.10.0":              true,
		"v12.3.45":             true,
		"dev":                  false,
		"v0.10.0-3-g4c5a15f":   false,
		"v0.10.0-dirty":        false,
		"0.10.0":               false,
		"v0.10":                false,
		"v0.10.x":              false,
		"4c5a15f":              false,
		"v0.10.0.1":            false,
		"v-1.0.0":              false,
		"v0.10.0-3-g4c5-dirty": false,
	} {
		if _, ok := releaseOf(said); ok != release {
			t.Errorf("%q read as a release: %v, want %v", said, ok, release)
		}
	}
}

func TestUpgrade_leavesWhatSomebodyElseKeeps(t *testing.T) {
	was := upgrade.Latest
	upgrade.Latest = "http://127.0.0.1:0/never-asked"
	t.Cleanup(func() { upgrade.Latest = was })
	if _, err := Upgrade("v0.10.0-3-g4c5a15f", "/usr/local/bin/cairn", "darwin/arm64"); !errors.Is(err, ErrBuiltFromSource) {
		t.Errorf("a build from source: %v", err)
	}
	if _, err := Upgrade("v0.10.0", "/opt/homebrew/Cellar/cairn/0.10.0/bin/cairn", "darwin/arm64"); !errors.Is(err, ErrKeptByHomebrew) {
		t.Errorf("a cairn Homebrew keeps: %v", err)
	}
}

func TestUpgraded_saysWhatWasDone(t *testing.T) {
	if said := (Upgraded{From: "v0.9.0", To: "v0.10.0", Binary: "/bin/cairn"}).String(); said != "cairn went from v0.9.0 to v0.10.0, in /bin/cairn\n" {
		t.Errorf("an upgrade says %q", said)
	}
	if said := (Upgraded{From: "v0.10.0", To: "v0.10.0", Binary: "/bin/cairn"}).String(); said != "cairn v0.10.0 is the newest already, in /bin/cairn\n" {
		t.Errorf("no upgrade says %q", said)
	}
}

func versionOf(said string) [3]int {
	v, _ := releaseOf(said)
	return v
}
