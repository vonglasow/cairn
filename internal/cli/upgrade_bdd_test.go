package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/cucumber/godog"

	"cairn/internal/build"
	"cairn/internal/upgrade"
	"cairn/test"
)

func initUpgrade(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerUpgrade(sc)
	}
}

func TestAcceptance_Upgrade(t *testing.T) {
	runOneAtATime(t, "upgrade.feature", initUpgrade(t), "")
}

const runBefore = "the cairn that was there"

func (w *world) registerUpgrade(sc *godog.ScenarioContext) {
	var server *httptest.Server
	var place, locked string
	var asked atomic.Int32
	wasLatest, wasRunning, wasStamps := upgrade.Latest, build.Running, stamps
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		server, place, locked = nil, "", ""
		asked.Store(0)
		return ctx, nil
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		upgrade.Latest, build.Running, stamps = wasLatest, wasRunning, wasStamps
		if server != nil {
			server.Close()
		}
		if locked != "" {
			_ = os.Chmod(locked, 0o755)
		}
		return ctx, err
	})
	released := func(r test.Release) {
		r.Asked = &asked
		server = r.Serve(w.T)
		upgrade.Latest = server.URL + "/latest"
	}
	here := runtime.GOOS + "/" + runtime.GOARCH
	sc.Step(`^cairn (v[0-9.]+) is released, with a binary for this system$`, func(tag string) {
		released(test.Release{Tag: tag, Platform: here, Binary: []byte("the cairn released as " + tag)})
	})
	sc.Step(`^cairn (v[0-9.]+) is released, its binary not matching the sum it gives$`, func(tag string) {
		released(test.Release{Tag: tag, Platform: here, Binary: []byte("the cairn released as " + tag), WrongSum: true})
	})
	sc.Step(`^cairn (v[0-9.]+) is released, with no binary for this system$`, func(tag string) {
		released(test.Release{Tag: tag, Platform: "plan9/mips", Binary: []byte("the cairn released as " + tag)})
	})
	sc.Step(`^the releases cannot be reached$`, func() {
		gone := httptest.NewServer(nil)
		gone.Close()
		upgrade.Latest = gone.URL + "/latest"
	})
	running := func(version string, folder ...string) error {
		place = filepath.Join(append([]string{w.Defaults}, append(folder, "cairn")...)...)
		if err := os.MkdirAll(filepath.Dir(place), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(place, []byte(runBefore), 0o755); err != nil {
			return err
		}
		stamps = func() build.Info { return build.Stamped(version, "", "", "") }
		build.Running = func() (string, error) { return place, nil }
		return nil
	}
	sc.Step(`^the cairn I run is (v[0-9.]+), from a release$`, func(version string) error {
		return running(version, "bin")
	})
	sc.Step(`^the cairn I run is (\S+), built from its source$`, func(version string) error {
		return running(version, "bin")
	})
	sc.Step(`^the cairn I run is (v[0-9.]+), installed by Homebrew$`, func(version string) error {
		return running(version, "homebrew", "Cellar", "cairn", "0.10.0", "bin")
	})
	sc.Step(`^the cairn I run is (v[0-9.]+), from a release, in a folder I cannot write$`, func(version string) error {
		if err := running(version, "bin"); err != nil {
			return err
		}
		locked = filepath.Dir(place)
		return os.Chmod(locked, 0o555)
	})
	sc.Step(`^it says cairn went from (\S+) to (\S+)$`, func(from, to string) error {
		return test.WantContains(w.Out.Stdout, "cairn went from "+from+" to "+to+", in "+place, "the answer")
	})
	sc.Step(`^it says (\S+) is the newest already$`, func(version string) error {
		return test.WantContains(w.Out.Stdout, "cairn "+version+" is the newest already, in "+place, "the answer")
	})
	sc.Step(`^it says in JSON it went from (\S+) to (\S+), in its place$`, func(from, to string) error {
		var said map[string]string
		if err := json.Unmarshal([]byte(w.Out.Stdout), &said); err != nil {
			return fmt.Errorf("stdout is not JSON (%w):\n%s", err, w.Out.Stdout)
		}
		if want := map[string]string{"from": from, "to": to, "binary": place}; !maps.Equal(said, want) {
			return fmt.Errorf("the answer is %v, want %v", said, want)
		}
		return nil
	})
	sc.Step(`^nothing was asked of the releases$`, func() error {
		if n := asked.Load(); n != 0 {
			return fmt.Errorf("the releases were asked %d times", n)
		}
		return nil
	})
	sc.Step(`^nothing but the cairn is in its folder$`, func() error {
		entries, err := os.ReadDir(filepath.Dir(place))
		if err != nil {
			return err
		}
		if len(entries) != 1 || entries[0].Name() != filepath.Base(place) {
			var names []string
			for _, e := range entries {
				names = append(names, e.Name())
			}
			return fmt.Errorf("its folder holds %v", names)
		}
		return nil
	})
	sc.Step(`^the cairn in its place is the one released as (\S+)$`, func(tag string) error {
		return inPlace(place, "the cairn released as "+tag)
	})
	sc.Step(`^the cairn in its place is the one that was there$`, func() error {
		return inPlace(place, runBefore)
	})
	sc.Step(`^it says the archive does not match its sum$`, func() error {
		return test.WantContains(w.Out.Stderr, "does not match its sum", "the refusal")
	})
	sc.Step(`^it says the release has no binary for this system$`, func() error {
		return test.WantContains(w.Out.Stderr, "no binary for this system", "the refusal")
	})
	sc.Step(`^it says the releases could not be read$`, func() error {
		return test.WantContains(w.Out.Stderr, "the releases could not be read", "the refusal")
	})
	sc.Step(`^it says a cairn built from its source is built again$`, func() error {
		return test.WantContains(w.Out.Stderr, "built from its source", "the refusal")
	})
	sc.Step(`^it says to run brew upgrade cairn$`, func() error {
		return test.WantContains(w.Out.Stderr, "brew upgrade cairn", "the refusal")
	})
	sc.Step(`^it says its place cannot be written$`, func() error {
		return test.WantContains(w.Out.Stderr, "cannot be written", "the refusal")
	})
}

func inPlace(place, want string) error {
	said, err := os.ReadFile(place)
	if err != nil {
		return err
	}
	if string(said) != want {
		return fmt.Errorf("%s holds %q, want %q", place, said, want)
	}
	info, err := os.Stat(place)
	if err != nil {
		return err
	}
	if info.Mode()&0o100 == 0 {
		return fmt.Errorf("%s cannot be run: %s", place, info.Mode())
	}
	return nil
}
