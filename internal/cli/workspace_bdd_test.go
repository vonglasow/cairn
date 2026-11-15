package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

type workspaces struct {
	*world
	held     []byte
	settings map[string]string
}

func initWorkspace(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := &workspaces{world: newWorld(t)}
		w.register(sc)
		w.registerConfig(sc)
		w.registerContext(sc)
		w.registerWorkspace(sc)
		w.registerPlaces(sc)
		w.registerRecent(sc)
		w.registerSettings(sc)
		w.registerServer(sc)
	}
}

func TestAcceptance_Workspace(t *testing.T) {
	runAcceptance(t, "workspace.feature", initWorkspace(t), "~@later")
}

func (w *workspaces) registerWorkspace(sc *godog.ScenarioContext) {
	sc.StepContext().Before(func(ctx context.Context, step *godog.Step) (context.Context, error) {
		if strings.HasPrefix(step.Text, "I run") || strings.HasPrefix(step.Text, "I add") {
			w.held, _ = os.ReadFile(w.configPath())
		}
		return ctx, nil
	})
	sc.Step(`^the workspaces "([^"]*)" at "([^"]*)" and "([^"]*)" at "([^"]*)", (the default being "([^"]*)"|with no default)$`, w.twoWorkspaces)
	sc.Step(`^it lists "([^"]*)" at "([^"]*)"( as the default)?$`, w.lists)
}

func (w *workspaces) twoWorkspaces(first, at, second, there, _, chosen string) error {
	said := "[workspaces]\n"
	if chosen != "" {
		said += fmt.Sprintf("default = %q\n", chosen)
	}
	said += fmt.Sprintf("%s = %q\n%s = %q\n", first, at, second, there)
	for _, folder := range []string{at, there} {
		if err := os.MkdirAll(w.expand(folder), 0o700); err != nil {
			return err
		}
	}
	return w.WriteConfig(said)
}

func (w *workspaces) lists(name, at, asDefault string) error {
	for _, line := range strings.Split(w.Out.Stdout, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != name || fields[1] != w.expand(at) {
			continue
		}
		if marked := len(fields) > 2 && fields[2] == "default"; marked != (asDefault != "") {
			return fmt.Errorf("%q is listed %q, marked the default %v, want %v", name, line, marked, asDefault != "")
		}
		return nil
	}
	return fmt.Errorf("%q at %q is not listed:\n%s", name, w.expand(at), w.Out.Stdout)
}
