package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
)

var quoted = regexp.MustCompile(`"([^"]*)"`)

func titlesNamed(list string) []string {
	found := quoted.FindAllStringSubmatch(list, -1)
	titles := make([]string, 0, len(found))
	for _, one := range found {
		titles = append(titles, one[1])
	}
	return titles
}

func rowsOf(listing string, nested bool) []string {
	rows := []string{}
	for _, line := range strings.Split(listing, "\n") {
		if !strings.Contains(line, "#") {
			continue
		}
		if strings.HasPrefix(line, " ") != nested {
			continue
		}
		rows = append(rows, line)
	}
	return rows
}

func listedInOrder(rows []string, list string) error {
	titles := titlesNamed(list)
	if len(rows) != len(titles) {
		return fmt.Errorf("%d listed, want the %d named %v:\n%s",
			len(rows), len(titles), titles, strings.Join(rows, "\n"))
	}
	for i, title := range titles {
		if !strings.Contains(rows[i], title) {
			return fmt.Errorf("row %d is %q, want %q:\n%s",
				i+1, strings.TrimSpace(rows[i]), title, strings.Join(rows, "\n"))
		}
	}
	return nil
}

func (w *world) registerOrdering(sc *godog.ScenarioContext) {
	sc.Step(`^it says on standard error "([^"]*)"$`, w.itSaysOnStandardError)
	sc.Step(`^it lists the tasks in the order (.+)$`, func(list string) error {
		w.Run("cairn task list")
		return listedInOrder(rowsOf(w.Out.Stdout, false), list)
	})

}
