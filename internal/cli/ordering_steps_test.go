package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/domain"
	"cairn/internal/task"
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

func hangingUnder(card string) string {
	_, under, found := strings.Cut(card, "sub-tasks")
	if !found {
		return ""
	}
	return under
}

func distinct(all []string) int {
	seen := map[string]bool{}
	for _, one := range all {
		seen[one] = true
	}
	return len(seen)
}

func columnsNamed(said string) []string {
	named := []string{}
	for _, state := range domain.States {
		if strings.Contains(said, string(state)) {
			named = append(named, string(state))
		}
	}
	return named
}

func (w *world) registerOrdering(sc *godog.ScenarioContext) {
	sc.Step(`^it lists the tasks in the order (.+)$`, func(list string) error {
		w.Run("cairn task list")
		return listedInOrder(rowsOf(w.Out.Stdout, false), list)
	})

	sc.Step(`^it lists the sub-tasks of task (\d+) in the order (.+)$`,
		func(n int, list string) error {
			w.Run(fmt.Sprintf("cairn task show %d", n))
			return listedInOrder(rowsOf(hangingUnder(w.Out.Stdout), true), list)
		})

	marked := func(title string) error {
		for _, line := range strings.Split(w.Out.Stdout, "\n") {
			if strings.Contains(line, title) && strings.Contains(line, task.InFocusMark) {
				return nil
			}
		}
		return fmt.Errorf("%q is not marked as in focus:\n%s", title, w.Out.Stdout)
	}
	sc.Step(`^it marks "([^"]*)" as the task in focus$`, marked)
	sc.Step(`^it marks "([^"]*)" as a task in focus$`, marked)

	sc.Step(`^it marks nothing else$`, func() error {
		return marksExactly(w.Out.Stdout, 1)
	})
	sc.Step(`^it marks (\d+) tasks? in focus$`, func(n int) error {
		return marksExactly(w.Out.Stdout, n)
	})

	sc.Step(`^it says which column each of them is in$`, func() error {
		if named := columnsNamed(w.Out.Stderr); len(named) < 2 {
			return fmt.Errorf("the reason names %v, want the column each of the two is in:\n%s",
				named, w.Out.Stderr)
		}
		return nil
	})

	sc.Step(`^it says whose sub-task it is$`, func() error {
		if named := numbersIn.FindAllString(w.Out.Stderr, -1); distinct(named) < 2 {
			return fmt.Errorf("the reason names %v, want both the sub-task and its parent:\n%s",
				named, w.Out.Stderr)
		}
		return nil
	})
}

func marksExactly(out string, want int) error {
	if got := strings.Count(out, task.InFocusMark); got != want {
		return fmt.Errorf("%d lines are marked as in focus, want %d:\n%s", got, want, out)
	}
	return nil
}
