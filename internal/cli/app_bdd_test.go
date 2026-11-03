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
	runAcceptance(t, "app.feature", initApp(t), "")
}

func (w *world) registerApp(sc *godog.ScenarioContext) {
	var signed []string
	wasHere, wasSigning, wasRunning := macapp.Here, macapp.Signing, macapp.Running
	const running = "/usr/local/bin/cairn"
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		signed = nil
		macapp.Here = true
		macapp.Running = func() (string, error) { return running, nil }
		macapp.Signing = func(app string) error {
			signed = append(signed, app)
			return nil
		}
		return ctx, nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		macapp.Here, macapp.Signing, macapp.Running = wasHere, wasSigning, wasRunning
		return ctx, err
	})
	installed := func() string { return filepath.Join(w.Defaults, "Applications", "cairn.app") }
	sc.Step("^I ran `([^`]*)`$", func(line string) error { return w.Seed(line) })
	sc.Step(`^this is not a Mac$`, func() { macapp.Here = false })
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
	sc.Step(`^it says an application is installed on a Mac only$`, func() error {
		return test.WantContains(w.Out.Stderr, "only on a Mac", "the refusal")
	})
}
