package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"cairn/test"
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
	sc.Step(`^a configuration naming the workspace "([^"]*)" at "([^"]*)"$`, func(name, folder string) error {
		return w.WriteConfig(fmt.Sprintf("[workspaces]\n%s = %q\n", name, folder))
	})
	sc.Step(`^the folder "([^"]*)" is not there$`, func(folder string) error {
		return os.RemoveAll(w.expand(folder))
	})
	sc.Step(`^the folder "([^"]*)" is there$`, func(folder string) error {
		_, err := os.Stat(w.expand(folder))
		return err
	})
	sc.Step(`^the folder "([^"]*)" is not made$`, func(folder string) error {
		if _, err := os.Stat(w.expand(folder)); !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s: %w, want it not made", folder, err)
		}
		return nil
	})
	sc.Step(`^it refuses, and names the folder "([^"]*)"$`, func(folder string) error {
		if err := w.itRefuses(); err != nil {
			return err
		}
		return test.WantContains(w.Out.Stderr, w.expand(folder), "the reason")
	})
	sc.Step(`^it lists "([^"]*)" at "([^"]*)"( as the default)?$`, w.lists)
	sc.Step(`^the structured answer holds "([^"]*)" at "([^"]*)" as the default$`, w.holdsTheDefault)
	sc.Step(`^the configuration still holds, unchanged, everything it held$`, w.stillHolds)
	sc.Step(`^the configuration reads, whole:$`, func(whole *godog.DocString) error {
		read, err := os.ReadFile(w.configPath())
		if err != nil {
			return err
		}
		if want := whole.Content + "\n"; string(read) != want {
			return fmt.Errorf("the configuration reads:\n%s\nwant:\n%s", read, want)
		}
		return nil
	})
	sc.Step(`^it says the default was "([^"]*)"$`, func(name string) error {
		return test.WantContains(w.Out.Stdout, fmt.Sprintf("it was %q", name), "the answer")
	})
	sc.Step(`^it says the workspace "([^"]*)" answers, why "([^"]*)"$`, func(name, why string) error {
		if err := test.WantContains(w.Out.Stdout, "workspace  "+name+"\n", "the answer"); err != nil {
			return err
		}
		return test.WantContains(w.Out.Stdout, "why        "+why, "the answer")
	})
	sc.Step(`^it says no workspace names the store, why "([^"]*)"$`, func(why string) error {
		if err := test.WantContains(w.Out.Stdout, "no workspace names this store", "the answer"); err != nil {
			return err
		}
		return test.WantContains(w.Out.Stdout, "why        "+why, "the answer")
	})
	sc.Step(`^the structured answer says why is "([^"]*)", over the default "([^"]*)"$`, w.whyAsJSON)
	sc.Step(`^it lists "([^"]*)" at "([^"]*)", marked "([^"]*)"$`, func(name, at, mark string) error {
		for _, line := range strings.Split(w.Out.Stdout, "\n") {
			if fields := strings.Fields(line); len(fields) > 1 && fields[0] == name && fields[1] == w.expand(at) {
				return test.WantContains(line, "  "+mark+"  ", "the line of "+name)
			}
		}
		return fmt.Errorf("%q at %q is not listed:\n%s", name, w.expand(at), w.Out.Stdout)
	})
	sc.Step(`^the structured answer holds "([^"]*)" with its time "([^"]*)"$`, w.holdsItsTime)
	sc.Step(`^it says the (store|notes) (?:is|are) at "([^"]*)"$`, func(place, at string) error {
		return test.WantContains(w.Out.Stdout, fmt.Sprintf("%-9s  %s\n", place, w.expand(at)), "the answer")
	})
	sc.Step(`^it says "([^"]*)"$`, func(said string) error {
		return test.WantContains(w.Out.Stdout, said, "the answer")
	})
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

func (w *workspaces) holdsTheDefault(name, at string) error {
	var doc struct {
		Workspaces []struct {
			Name    string `json:"name"`
			Folder  string `json:"folder"`
			Default bool   `json:"default"`
		} `json:"workspaces"`
	}
	if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
		return fmt.Errorf("the answer is not JSON (%w):\n%s", err, w.Out.Stdout)
	}
	for _, one := range doc.Workspaces {
		if one.Name == name && one.Folder == w.expand(at) && one.Default {
			return nil
		}
	}
	return fmt.Errorf("the answer does not hold %q at %q as the default:\n%s", name, at, w.Out.Stdout)
}

func (w *workspaces) stillHolds() error {
	now, err := os.ReadFile(w.configPath())
	if err != nil {
		return err
	}
	rest := string(now)
	for _, line := range strings.SplitAfter(string(w.held), "\n") {
		at := strings.Index(rest, line)
		if at < 0 {
			return fmt.Errorf("the configuration lost %q; it held:\n%s\nand now holds:\n%s", line, w.held, now)
		}
		rest = rest[at+len(line):]
	}
	return nil
}

func (w *workspaces) whyAsJSON(rung, chosen string) error {
	var doc struct {
		Why struct {
			Rung string `json:"rung"`
			Over []struct {
				Rung  string `json:"rung"`
				Names string `json:"names"`
			} `json:"over"`
		} `json:"why"`
	}
	if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
		return fmt.Errorf("the answer is not JSON (%w):\n%s", err, w.Out.Stdout)
	}
	if doc.Why.Rung != rung || len(doc.Why.Over) != 1 || doc.Why.Over[0].Rung != "the default" || doc.Why.Over[0].Names != chosen {
		return fmt.Errorf("the answer says why %+v, want %q over the default %q:\n%s", doc.Why, rung, chosen, w.Out.Stdout)
	}
	return nil
}

func (w *workspaces) holdsItsTime(name, time string) error {
	var doc struct {
		Workspaces []struct {
			Name string `json:"name"`
			Time string `json:"time"`
		} `json:"workspaces"`
	}
	if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
		return fmt.Errorf("the answer is not JSON (%w):\n%s", err, w.Out.Stdout)
	}
	for _, one := range doc.Workspaces {
		if one.Name == name && one.Time == time {
			return nil
		}
	}
	return fmt.Errorf("the answer does not hold %q with its time %q:\n%s", name, time, w.Out.Stdout)
}
