package cli

// Executable acceptance scenarios for bdd/ordering.feature. The steps every
// feature shares live in world_test.go; only what belongs to the order and the
// focus is declared here.

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/domain"
	"cairn/internal/task"
)

// quoted is how a step names a list of titles: each one in its own quotes.
var quoted = regexp.MustCompile(`"([^"]*)"`)

// titlesNamed reads the titles a step listed as `"First", "Second", "Third"`.
//
// Read by their quotes and not by the commas between them: a title may carry a
// comma, and splitting on one turned "First, renamed" into two titles — which
// made a scenario fail for a reason that had nothing to do with cairn.
func titlesNamed(list string) []string {
	found := quoted.FindAllStringSubmatch(list, -1)
	titles := make([]string, 0, len(found))
	for _, one := range found {
		titles = append(titles, one[1])
	}
	return titles
}

// rowsOf reads the lines of a listing that carry a task, keeping only those at
// the level asked about: a tree indents a sub-task under its parent, so
// counting every line would make a task's sisters and its sub-tasks one crowd.
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

// listedInOrder fails unless the rows given are exactly the titles named, in
// the order named. Row by row rather than by looking each title up: a title
// found somewhere in the listing says nothing about which line it is on, and
// counting alone would pass on any order at all.
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

// hangingUnder is the part of a task's card that lists what hangs under it.
// Read out of the card rather than from the whole listing, so the parent's own
// line is not counted among its sub-tasks.
func hangingUnder(card string) string {
	_, under, found := strings.Cut(card, "sub-tasks")
	if !found {
		return ""
	}
	return under
}

// distinct counts how many different strings a list holds.
func distinct(all []string) int {
	seen := map[string]bool{}
	for _, one := range all {
		seen[one] = true
	}
	return len(seen)
}

// columnsNamed reads which of the columns a message names.
func columnsNamed(said string) []string {
	named := []string{}
	for _, state := range domain.States {
		if strings.Contains(said, string(state)) {
			named = append(named, string(state))
		}
	}
	return named
}

// registerOrdering declares the steps about the one order there is and about
// which task is being worked on.
func (w *world) registerOrdering(sc *godog.ScenarioContext) {
	// The listing is asked for rather than read out of whatever the command
	// under test answered. Not every command answers with one — adding a task
	// answers with its card — and a step that used the last answer when it
	// happened to be a listing would mean two different things depending on
	// which command ran, which is how a scenario passes for the wrong reason.
	sc.Step(`^it lists the tasks in the order (.+)$`, func(list string) error {
		w.Run("cairn task list")
		return listedInOrder(rowsOf(w.Out.Stdout, false), list)
	})

	// The sub-tasks are asked for rather than read out of what the command
	// under test answered: adding a sub-task answers with the card of the
	// sub-task, which says nothing about the order of its sisters.
	sc.Step(`^it lists the sub-tasks of task (\d+) in the order (.+)$`,
		func(n int, list string) error {
			w.Run(fmt.Sprintf("cairn task show %d", n))
			return listedInOrder(rowsOf(hangingUnder(w.Out.Stdout), true), list)
		})

	sc.Step(`^it marks "([^"]*)" as the task in focus$`, func(title string) error {
		for _, line := range strings.Split(w.Out.Stdout, "\n") {
			if strings.Contains(line, title) && strings.Contains(line, task.InFocusMark) {
				return nil
			}
		}
		return fmt.Errorf("%q is not marked as the task in focus:\n%s", title, w.Out.Stdout)
	})

	// This is the step that makes the single holder observable. Showing one
	// task marked proves nothing on its own: the answer would read the same
	// with every task marked.
	sc.Step(`^it marks nothing else$`, func() error {
		if marked := strings.Count(w.Out.Stdout, task.InFocusMark); marked != 1 {
			return fmt.Errorf("%d lines are marked as in focus, want exactly one:\n%s",
				marked, w.Out.Stdout)
		}
		return nil
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
