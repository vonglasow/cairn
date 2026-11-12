package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/cucumber/godog"
	"github.com/spf13/cobra"

	"cairn/internal/domain"
	"cairn/internal/engine"
	"cairn/test"
)

const recentHeading = "recent folders"

func (w *workspaces) registerRecent(sc *godog.ScenarioContext) {
	sc.Step("^I open the board with `([^`]*)`$", w.openTheBoardWith)
	sc.Step(`^the environment says XDG_STATE_HOME is "([^"]*)"$`, func(place string) error {
		w.Setenv(engine.StateVariable, w.expand(place))
		return nil
	})
	sc.Step(`^ten recent folders, "~/old/1" opened first and "~/old/10" last$`, w.tenRecentFolders)
	sc.Step(`^it lists the recent folder "([^"]*)"( as not found)?$`, w.listsTheRecentFolder)
	sc.Step(`^it does not list the recent folder "([^"]*)"$`, func(folder string) error {
		if line := w.recentLine(folder); line != "" {
			return fmt.Errorf("%q is listed among the recent folders: %q", folder, line)
		}
		return nil
	})
	sc.Step(`^it lists no recent folder$`, func() error {
		if strings.Contains(w.Out.Stdout, recentHeading) {
			return fmt.Errorf("it lists recent folders:\n%s", w.Out.Stdout)
		}
		return nil
	})
	sc.Step(`^the recent folders listed are (\d+), "([^"]*)" first and "([^"]*)" next$`, w.recentFoldersListed)
	sc.Step(`^the structured answer holds the recent folder "([^"]*)", opened today$`, w.holdsTheRecentFolder)
	sc.Step(`^the file "([^"]*)" names "([^"]*)"$`, func(file, store string) error {
		read, err := os.ReadFile(w.expand(file))
		if err != nil {
			return err
		}
		return test.WantContains(string(read), w.expand(store), file)
	})
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
	sc.Step(`^it lists "([^"]*)" at "([^"]*)" as not found$`, func(name, at string) error {
		fields, err := w.workspaceFields(name, at)
		if err != nil {
			return err
		}
		return test.WantContains(strings.Join(fields, " "), "not found", "the line of "+name)
	})
	sc.Step(`^it does not list "([^"]*)"$`, func(name string) error {
		for _, line := range strings.Split(w.Out.Stdout, "\n") {
			if fields := strings.Fields(line); len(fields) > 0 && fields[0] == name {
				return fmt.Errorf("%q is listed: %q", name, line)
			}
		}
		return nil
	})
	sc.Step(`^"([^"]*)" wears one of the label colours$`, func(name string) error {
		colour, err := w.colourListed(name)
		if err != nil {
			return err
		}
		if !slices.Contains(domain.LabelColours, colour) {
			return fmt.Errorf("%q wears %q, none of the label colours", name, colour)
		}
		return nil
	})
	sc.Step(`^"([^"]*)" wears the same colour when the workspaces are listed again$`, func(name string) error {
		was, err := w.colourListed(name)
		if err != nil {
			return err
		}
		w.Run("cairn workspace list")
		now, err := w.colourListed(name)
		if err != nil {
			return err
		}
		if now != was {
			return fmt.Errorf("%q wore %q, and %q listed again", name, was, now)
		}
		return nil
	})
	sc.Step(`^the structured answer holds "([^"]*)" wearing "([^"]*)"$`, func(name, colour string) error {
		doc, err := w.listedAsJSON()
		if err != nil {
			return err
		}
		for _, one := range doc.Workspaces {
			if one.Name == name && one.Colour == colour {
				return nil
			}
		}
		return fmt.Errorf("the answer does not hold %q wearing %q:\n%s", name, colour, w.Out.Stdout)
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

func (w *workspaces) tenRecentFolders() error {
	type kept struct {
		Store  string    `json:"store"`
		Opened time.Time `json:"opened"`
	}
	var folders []kept
	at := time.Now().Add(-time.Hour)
	for n := 10; n >= 1; n-- {
		folders = append(folders, kept{Store: w.expand(fmt.Sprintf("~/old/%d/cairn.db", n)), Opened: at.Add(time.Duration(n) * time.Minute)})
	}
	said, err := json.Marshal(map[string][]kept{"recent": folders})
	if err != nil {
		return err
	}
	path := w.expand("~/.local/state/cairn/recent.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, said, 0o600)
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

func (w *workspaces) recentFoldersListed(count int, first, next string) error {
	lines := w.recentLines()
	if len(lines) != count {
		return fmt.Errorf("%d recent folders are listed, want %d:\n%s", len(lines), count, w.Out.Stdout)
	}
	for i, folder := range []string{first, next} {
		if fields := strings.Fields(lines[i]); fields[0] != w.expand(folder) {
			return fmt.Errorf("recent folder %d is %q, want %q", i+1, fields[0], w.expand(folder))
		}
	}
	return nil
}

type listedJSON struct {
	Workspaces []struct {
		Name   string `json:"name"`
		Colour string `json:"colour"`
	} `json:"workspaces"`
	Recent []struct {
		Folder string    `json:"folder"`
		Opened time.Time `json:"opened"`
	} `json:"recent"`
}

func (w *workspaces) listedAsJSON() (listedJSON, error) {
	var doc listedJSON
	if err := json.Unmarshal([]byte(w.Out.Stdout), &doc); err != nil {
		return doc, fmt.Errorf("the answer is not JSON (%w):\n%s", err, w.Out.Stdout)
	}
	return doc, nil
}

func (w *workspaces) holdsTheRecentFolder(folder string) error {
	doc, err := w.listedAsJSON()
	if err != nil {
		return err
	}
	for _, one := range doc.Recent {
		if one.Folder == w.expand(folder) && one.Opened.Format(time.DateOnly) == time.Now().Format(time.DateOnly) {
			return nil
		}
	}
	return fmt.Errorf("the answer does not hold the recent folder %q, opened today:\n%s", w.expand(folder), w.Out.Stdout)
}

func (w *workspaces) workspaceFields(name, at string) ([]string, error) {
	for _, line := range strings.Split(w.Out.Stdout, "\n") {
		if fields := strings.Fields(line); len(fields) > 2 && fields[0] == name && fields[1] == w.expand(at) {
			return fields, nil
		}
	}
	return nil, fmt.Errorf("%q at %q is not listed:\n%s", name, w.expand(at), w.Out.Stdout)
}

func (w *workspaces) colourListed(name string) (string, error) {
	for _, line := range strings.Split(w.Out.Stdout, "\n") {
		if fields := strings.Fields(line); len(fields) > 2 && fields[0] == name {
			return fields[len(fields)-1], nil
		}
	}
	return "", fmt.Errorf("%q is not listed:\n%s", name, w.Out.Stdout)
}
