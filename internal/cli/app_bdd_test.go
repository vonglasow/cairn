package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"cairn/internal/build"
	"cairn/internal/linuxapp"
	"cairn/internal/macapp"
	"cairn/test"
)

func initApp(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerApp(sc)
	}
}

func TestAcceptance_App(t *testing.T) {
	runOneAtATime(t, "app.feature", initApp(t), "")
}

func (w *world) registerApp(sc *godog.ScenarioContext) {
	var signed []string
	var data, tools string
	wasHere, wasLinux, wasSigning, wasRunning := macapp.Here, linuxapp.Here, macapp.Signing, build.Running
	const running = "/usr/local/bin/cairn"
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		signed = nil
		macapp.Here = true
		build.Running = func() (string, error) { return running, nil }
		macapp.Signing = func(app string) error {
			signed = append(signed, app)
			return nil
		}
		return ctx, nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		macapp.Here, linuxapp.Here, macapp.Signing, build.Running = wasHere, wasLinux, wasSigning, wasRunning
		return ctx, err
	})
	installed := func() string { return filepath.Join(w.Defaults, "Applications", "cairn.app") }
	sc.Step("^I ran `([^`]*)`$", func(line string) error { return w.Seed(line) })
	sc.Step(`^this is neither a Mac nor Linux$`, func() { macapp.Here, linuxapp.Here = false, false })
	sc.Given(`^this is a Linux desktop$`, func() error {
		macapp.Here, linuxapp.Here = false, true
		w.Setenv("XDG_DATA_HOME", "")
		data = filepath.Join(w.Defaults, ".local", "share")
		tools = filepath.Join(w.Defaults, "tools")
		w.SetProcessEnv("PATH", tools)
		if err := os.MkdirAll(tools, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(tools, "xdg-desktop-menu"), []byte("#!/bin/sh\necho \"$*\" >> '"+filepath.Join(tools, "refreshed")+"'\n"), 0o755)
	})
	entry := func() string { return filepath.Join(data, "applications", "cairn.desktop") }
	const theirs = "[Desktop Entry]\nName=mine\nExec=mine\n"
	sc.Given(`^a desktop entry of cairn's name written by somebody else$`, func() error {
		if err := os.MkdirAll(filepath.Dir(entry()), 0o755); err != nil {
			return err
		}
		return os.WriteFile(entry(), []byte(theirs), 0o644)
	})
	sc.Step(`^it says that entry is left alone$`, func() error {
		return test.WantContains(w.Out.Stderr, "left alone", "the refusal")
	})
	sc.Step(`^the entry is still the one somebody else wrote$`, func() error {
		if said, err := os.ReadFile(entry()); err != nil || string(said) != theirs {
			return fmt.Errorf("the entry now says %q, %w", said, err)
		}
		return nil
	})
	sc.Step(`^it says the application is in my Applications folder$`, func() error {
		return test.WantContains(w.Out.Stdout, installed(), "the answer")
	})
	sc.Step(`^the application runs the window of the cairn that wrote it$`, func() error {
		launcher, err := os.ReadFile(filepath.Join(installed(), "Contents", "MacOS", "cairn"))
		if err != nil {
			return err
		}
		if !strings.Contains(string(launcher), "exec '"+running+"' gui") {
			return fmt.Errorf("the application runs:\n%s\nwant %s gui", launcher, running)
		}
		info, err := os.Stat(filepath.Join(installed(), "Contents", "MacOS", "cairn"))
		if err != nil {
			return err
		}
		if info.Mode()&0o100 == 0 {
			return fmt.Errorf("the launcher cannot be run: %s", info.Mode())
		}
		return nil
	})
	sc.Step(`^the application carries cairn's icon and its version$`, func() error {
		plist, err := os.ReadFile(filepath.Join(installed(), "Contents", "Info.plist"))
		if err != nil {
			return err
		}
		for _, want := range []string{"<string>cairn</string>", "<string>" + build.Read().Version + "</string>", "<string>cairn.icns</string>"} {
			if !strings.Contains(string(plist), want) {
				return fmt.Errorf("Info.plist holds no %s:\n%s", want, plist)
			}
		}
		icon, err := os.ReadFile(filepath.Join(installed(), "Contents", "Resources", "cairn.icns"))
		if err != nil || !strings.HasPrefix(string(icon), "icns") {
			return fmt.Errorf("no icns icon: %w", err)
		}
		return nil
	})
	sc.Step(`^the application was signed for this machine$`, func() error {
		if len(signed) != 1 || signed[0] != installed() {
			return fmt.Errorf("signed %v, want %s once", signed, installed())
		}
		return nil
	})
	sc.Step(`^it says an application is installed on a Mac or a Linux desktop only$`, func() error {
		return test.WantContains(w.Out.Stderr, "only on a Mac or a Linux desktop", "the refusal")
	})
}
