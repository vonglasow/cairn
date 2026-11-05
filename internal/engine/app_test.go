package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/build"
	"cairn/internal/linuxapp"
	"cairn/internal/macapp"
)

func installing(t *testing.T, mac, linux bool, running string) string {
	t.Helper()
	wasHere, wasLinux, wasSigning, wasRunning := macapp.Here, linuxapp.Here, macapp.Signing, build.Running
	macapp.Here, linuxapp.Here, macapp.Signing = mac, linux, func(string) error { return nil }
	build.Running = func() (string, error) { return running, nil }
	t.Cleanup(func() {
		macapp.Here, linuxapp.Here, macapp.Signing, build.Running = wasHere, wasLinux, wasSigning, wasRunning
	})
	t.Setenv("PATH", t.TempDir())
	return t.TempDir()
}

func TestAppInstall_writesTheApplicationRunningThisCairn(t *testing.T) {
	running := "/usr/local/bin/cairn"
	home := installing(t, true, false, running)
	answer, err := AppInstall.Run(On{Home: Home{Dir: home}}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	app := filepath.Join(home, "Applications", "cairn.app")
	if answer.Installed == nil || *answer.Installed != app {
		t.Fatalf("installed %v, want %s", answer.Installed, app)
	}
	if launcher, _ := os.ReadFile(filepath.Join(app, "Contents", "MacOS", "cairn")); !strings.Contains(string(launcher), running) {
		t.Errorf("the launcher runs %q, not %s", launcher, running)
	}
	said := RenderText(AppInstall, answer)
	if !strings.Contains(said, app) || !strings.Contains(said, running) || !strings.Contains(said, "Spotlight") {
		t.Errorf("it says %q", said)
	}
	if shaped, ok := Of(AppInstall, answer).(AppInstalled); !ok || shaped.Installed != app || shaped.Runs != running || !strings.Contains(shaped.OpenedFrom, "Spotlight") {
		t.Errorf("a machine is told %#v", Of(AppInstall, answer))
	}
	if _, err := AppInstall.Run(On{}, Input{}); err == nil {
		t.Error("with no home, it installed somewhere")
	}
}

func TestAppInstall_writesTheDesktopEntryOnLinux(t *testing.T) {
	running := "/usr/bin/cairn"
	home := installing(t, false, true, running)
	data := filepath.Join(home, "data")
	answer, err := AppInstall.Run(On{Home: Home{Dir: home, Lookup: func(key string) (string, bool) { return data, key == "XDG_DATA_HOME" }}}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(data, "applications", "cairn.desktop")
	if answer.Installed == nil || *answer.Installed != entry || *answer.Runs != running {
		t.Fatalf("installed %v running %v, want %s", answer.Installed, answer.Runs, entry)
	}
	if said, _ := os.ReadFile(entry); !strings.Contains(string(said), "Exec="+running+" gui") {
		t.Errorf("the entry says %q", said)
	}
	if said := RenderText(AppInstall, answer); !strings.Contains(said, entry) || !strings.Contains(said, "applications menu") {
		t.Errorf("it says %q", said)
	}
}

func TestAppInstall_refusesWhereThereIsNoApplicationToInstall(t *testing.T) {
	home := installing(t, false, false, "/usr/local/bin/cairn")
	if _, err := AppInstall.Run(On{Home: Home{Dir: home}}, Input{}); !errors.Is(err, ErrNoApplicationHere) {
		t.Errorf("elsewhere: %v", err)
	}
}

func TestAppInstall_refusesACairnThatWillNotStay(t *testing.T) {
	for _, mac := range []bool{true, false} {
		home := installing(t, mac, !mac, filepath.Join(os.TempDir(), "go-build123", "b001", "exe", "cairn"))
		if _, err := AppInstall.Run(On{Home: Home{Dir: home}}, Input{}); !errors.Is(err, ErrPassing) {
			t.Errorf("a go run build, mac %v: %v", mac, err)
		}
	}
}
