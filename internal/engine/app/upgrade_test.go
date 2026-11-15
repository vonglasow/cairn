package app

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"cairn/internal/upgrade"
)

func TestUpgraded_saysWhatWasDone(t *testing.T) {
	said := (Upgraded{From: "v0.9.0", To: "v0.10.0", Binary: "/bin/cairn"}).String()
	if !strings.HasPrefix(said, "cairn went from v0.9.0 to v0.10.0, in /bin/cairn\n") ||
		!strings.Contains(said, "/plugin marketplace add vonglasow/cairn@v0.10.0\n") {
		t.Errorf("an upgrade says %q", said)
	}
	if said := (Upgraded{From: "v0.10.0", To: "v0.10.0", Binary: "/bin/cairn"}).String(); said != "cairn v0.10.0 is the newest already, in /bin/cairn\n" {
		t.Errorf("no upgrade says %q", said)
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
