package cli

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/cucumber/godog"
	"github.com/spf13/cobra"
)

const recentHeading = "recent folders"

func (w *workspaces) registerRecent(sc *godog.ScenarioContext) {
	sc.Step("^I open the board with `([^`]*)`$", w.openTheBoardWith)
	sc.Step(`^it lists the recent folder "([^"]*)"( as not found)?$`, w.listsTheRecentFolder)
	sc.Step(`^it lists "([^"]*)" at "([^"]*)" wearing "([^"]*)"$`, func(name, at, colour string) error {
		fields, err := w.workspaceFields(name, at)
		if err != nil {
			return err
		}
		if fields[len(fields)-1] != colour {
			return fmt.Errorf("%q wears %q, not %q", name, fields[len(fields)-1], colour)
		}
		return nil
	})
	sc.Step(`^it does not list "([^"]*)"$`, func(name string) error {
		for _, line := range strings.Split(w.Out.Stdout, "\n") {
			if fields := strings.Fields(line); len(fields) > 0 && fields[0] == name {
				return fmt.Errorf("%q is listed: %q", name, line)
			}
		}
		return nil
	})
}

func (w *workspaces) openTheBoardWith(line string) error {
	w.terminal = func(*cobra.Command, tea.Model) error { return nil }
	w.Run(line)
	if w.Out.Code != exitOK {
		return fmt.Errorf("the board did not open: exit %d (%s)", w.Out.Code, w.Out.Stderr)
	}
	return nil
}

func (w *workspaces) recentLines() []string {
	_, after, found := strings.Cut(w.Out.Stdout, recentHeading)
	if !found {
		return nil
	}
	var lines []string
	for _, line := range strings.Split(after, "\n")[1:] {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func (w *workspaces) recentLine(folder string) string {
	for _, line := range w.recentLines() {
		if fields := strings.Fields(line); len(fields) > 0 && fields[0] == w.expand(folder) {
			return line
		}
	}
	return ""
}

func (w *workspaces) listsTheRecentFolder(folder, notFound string) error {
	line := w.recentLine(folder)
	if line == "" {
		return fmt.Errorf("%q is not listed among the recent folders:\n%s", w.expand(folder), w.Out.Stdout)
	}
	if said := strings.Contains(line, "not found"); said != (notFound != "") {
		return fmt.Errorf("%q is listed %q, said not found %v, want %v", folder, line, said, notFound != "")
	}
	return nil
}

func (w *workspaces) workspaceFields(name, at string) ([]string, error) {
	for _, line := range strings.Split(w.Out.Stdout, "\n") {
		if fields := strings.Fields(line); len(fields) > 2 && fields[0] == name && fields[1] == w.expand(at) {
			return fields, nil
		}
	}
	return nil, fmt.Errorf("%q at %q is not listed:\n%s", name, w.expand(at), w.Out.Stdout)
}
