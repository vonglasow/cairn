package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
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

func TestUpgraded_answersFromToAndBinaryInJSON(t *testing.T) {
	said, err := json.Marshal(Upgraded{From: "v0.10.0", To: "v0.10.0", Binary: "/bin/cairn"})
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"from":"v0.10.0","to":"v0.10.0","binary":"/bin/cairn"}`; string(said) != want {
		t.Errorf("in JSON %s, want %s", said, want)
	}
}

func TestUpgrade_triesItsPlaceBeforeAskingTheReleases(t *testing.T) {
	var asked atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { asked.Add(1) }))
	t.Cleanup(server.Close)
	was := upgrade.Latest
	upgrade.Latest = server.URL + "/latest"
	t.Cleanup(func() { upgrade.Latest = was })
	locked := t.TempDir()
	if err := os.Chmod(locked, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })
	if _, err := Upgrade("v0.10.0", filepath.Join(locked, "cairn"), "darwin/arm64"); !errors.Is(err, upgrade.ErrNotWritable) {
		t.Errorf("a place that cannot be written: %v", err)
	}
	if n := asked.Load(); n != 0 {
		t.Errorf("the releases were asked %d times", n)
	}
}

func releasesSaying(t *testing.T, tag string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `{"tag_name":%q,"assets":[]}`, tag)
	}))
	t.Cleanup(server.Close)
	was := upgrade.Latest
	upgrade.Latest = server.URL + "/latest"
	t.Cleanup(func() { upgrade.Latest = was })
}

func TestUpgrade_leavesTheNewestAsItIs(t *testing.T) {
	releasesSaying(t, "v0.10.0")
	binary := filepath.Join(t.TempDir(), "cairn")
	got, err := Upgrade("v0.10.0", binary, "darwin/arm64")
	if err != nil || got != (Upgraded{From: "v0.10.0", To: "v0.10.0", Binary: binary}) {
		t.Errorf("the newest already: %+v, %v", got, err)
	}
}

func TestUpgrade_saysAReleaseWithNothingForThisSystem(t *testing.T) {
	releasesSaying(t, "v0.11.0")
	if _, err := Upgrade("v0.10.0", filepath.Join(t.TempDir(), "cairn"), "darwin/arm64"); !errors.Is(err, upgrade.ErrNoBinaryHere) {
		t.Errorf("a release with no binary here: %v", err)
	}
}
