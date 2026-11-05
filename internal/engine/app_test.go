package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cairn/internal/build"
	"cairn/internal/macapp"
)

func TestAppInstall_writesTheApplicationRunningThisCairn(t *testing.T) {
	home := t.TempDir()
	wasHere, wasSigning, wasRunning := macapp.Here, macapp.Signing, build.Running
	running := "/usr/local/bin/cairn"
	macapp.Here, macapp.Signing = true, func(string) error { return nil }
	build.Running = func() (string, error) { return running, nil }
	t.Cleanup(func() { macapp.Here, macapp.Signing, build.Running = wasHere, wasSigning, wasRunning })
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
	if shaped, ok := Of(AppInstall, answer).(AppInstalled); !ok || shaped.Installed != app || shaped.Runs != running {
		t.Errorf("a machine is told %#v", Of(AppInstall, answer))
	}
	macapp.Here = false
	if _, err := AppInstall.Run(On{Home: Home{Dir: home}}, Input{}); !errors.Is(err, macapp.ErrOnlyOnAMac) {
		t.Errorf("elsewhere: %v", err)
	}
	if _, err := AppInstall.Run(On{}, Input{}); err == nil {
		t.Error("with no home, it installed somewhere")
	}
}
