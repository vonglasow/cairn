package engine

import (
	"cairn/internal/engine/app"
	"errors"
	"os"
	"path/filepath"
	"slices"
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
	answer, err := appInstall.Run(workedOn{Home: Home{Dir: home}}, Input{})
	if err != nil {
		t.Fatal(err)
	}
	installed := filepath.Join(home, "Applications", "cairn.app")
	if answer.Installed == nil || *answer.Installed != installed {
		t.Fatalf("installed %v, want %s", answer.Installed, installed)
	}
	if launcher, _ := os.ReadFile(filepath.Join(installed, "Contents", "MacOS", "cairn")); !strings.Contains(string(launcher), running) {
		t.Errorf("the launcher runs %q, not %s", launcher, running)
	}
	if said := Of(appInstall, answer).(app.Installed); said.Installed != installed || said.Runs != running || !strings.Contains(said.OpenedFrom, "Spotlight") {
		t.Errorf("it answers %+v", said)
	}
	if shaped, ok := Of(appInstall, answer).(app.Installed); !ok || shaped.Installed != installed || shaped.Runs != running || !strings.Contains(shaped.OpenedFrom, "Spotlight") {
		t.Errorf("a machine is told %#v", Of(appInstall, answer))
	}
	if _, err := appInstall.Run(workedOn{}, Input{}); err == nil {
		t.Error("with no home, it installed somewhere")
	}
}

func TestAppInstall_writesTheDesktopEntryOnLinux(t *testing.T) {
	running := "/usr/bin/cairn"
	home := installing(t, false, true, running)
	data := filepath.Join(home, "data")
	answer, err := appInstall.Run(workedOn{Home: Home{Dir: home, Lookup: func(key string) (string, bool) { return data, key == "XDG_DATA_HOME" }}}, Input{})
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
	if said := Of(appInstall, answer).(app.Installed); said.Installed != entry || !strings.Contains(said.OpenedFrom, "applications menu") {
		t.Errorf("it answers %+v", said)
	}
}

func TestAppInstall_refusesWhereThereIsNoApplicationToInstall(t *testing.T) {
	home := installing(t, false, false, "/usr/local/bin/cairn")
	if _, err := appInstall.Run(workedOn{Home: Home{Dir: home}}, Input{}); !errors.Is(err, app.ErrNoApplicationHere) {
		t.Errorf("elsewhere: %v", err)
	}
}

func TestAppInstall_refusesACairnThatWillNotStay(t *testing.T) {
	for _, mac := range []bool{true, false} {
		home := installing(t, mac, !mac, filepath.Join(os.TempDir(), "go-build123", "b001", "exe", "cairn"))
		if _, err := appInstall.Run(workedOn{Home: Home{Dir: home}}, Input{}); !errors.Is(err, app.ErrPassing) {
			t.Errorf("a go run build, mac %v: %v", mac, err)
		}
	}
}

func TestUpgrade_isAskedByAPersonAsTheCairnThatRuns(t *testing.T) {
	if slices.ContainsFunc(All(), func(o Operation) bool { return o.Line() == "upgrade" }) ||
		!slices.ContainsFunc(ForAPersonOnly(), func(o Operation) bool { return o.Line() == "upgrade" }) {
		t.Error("upgrade is to be asked by a person only, never over MCP")
	}
	installing(t, false, false, "/opt/homebrew/Cellar/cairn/0.10.0/bin/cairn")
	if _, err := Upgrade.Run(workedOn{Home: Home{Version: "v0.10.0"}}, Input{}); !errors.Is(err, app.ErrKeptByHomebrew) {
		t.Errorf("err = %v, want the cairn Homebrew keeps left to it", err)
	}
	if _, err := Upgrade.Run(workedOn{Home: Home{Version: "dev"}}, Input{}); !errors.Is(err, app.ErrBuiltFromSource) {
		t.Errorf("err = %v, want a cairn built from its source refused", err)
	}
}
