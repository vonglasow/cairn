package cli

// Steps observed only by bdd/timesheet.feature.

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"cairn/internal/domain"
	"cairn/test"
)

// dayBlock is the lines a timesheet writes under one day, the day's own line
// excluded; nothing when the day is not written at all.
func (w *world) dayBlock(day string) []string {
	var block []string
	in := false
	for _, l := range strings.Split(w.Out.Stdout, "\n") {
		if strings.TrimSpace(l) == day {
			in = true
			continue
		}
		if in && l != "" && !strings.HasPrefix(l, " ") {
			break
		}
		if in && l != "" {
			block = append(block, strings.TrimSpace(l))
		}
	}
	return block
}

// registerTimesheet declares the steps only bdd/timesheet.feature uses.
func (w *world) registerTimesheet(sc *godog.ScenarioContext) {
	sc.Step("^I ran `([^`]*)`$", func(line string) error { return w.Seed(line) })

	// A task's line reads its number, then its time, then its title: the time
	// comes before the title so a title holding "1h" cannot be read as one.
	sc.Step(`^the day "([^"]*)" gives task (\d+) "([^"]*)"$`, func(day string, n int, want string) error {
		for _, l := range w.dayBlock(day) {
			fields := strings.Fields(l)
			if len(fields) >= 2 && fields[0] == fmt.Sprintf("#%d", n) {
				if fields[1] != want {
					return fmt.Errorf("the day %s gives task %d %q, want %q:\n%s", day, n, fields[1], want, w.Out.Stdout)
				}
				return nil
			}
		}
		return fmt.Errorf("the day %s says nothing of task %d:\n%s", day, n, w.Out.Stdout)
	})

	// A day nothing was done on is not written at all, and comes to nothing.
	sc.Step(`^the day "([^"]*)" comes to "([^"]*)"$`, func(day, want string) error {
		block := w.dayBlock(day)
		if len(block) == 0 {
			if want == "0m" {
				return nil
			}
			return fmt.Errorf("the timesheet says nothing of %s:\n%s", day, w.Out.Stdout)
		}
		for _, l := range block {
			if fields := strings.Fields(l); len(fields) == 2 && fields[0] == "total" {
				if fields[1] != want {
					return fmt.Errorf("the day %s comes to %q, want %q", day, fields[1], want)
				}
				return nil
			}
		}
		return fmt.Errorf("the day %s says no total:\n%s", day, w.Out.Stdout)
	})

	sc.Step(`^it says today$`, func() error {
		return test.WantContains(w.Out.Stdout, domain.FormatDay(time.Now()), "the timesheet")
	})

	sc.Step(`^it says (\S+) was added for "([^"]*)"$`, func(span, day string) error {
		for _, l := range strings.Split(w.Out.Stdout, "\n") {
			if strings.Contains(l, "time added") && strings.Contains(l, span) && strings.Contains(l, day) {
				return nil
			}
		}
		return fmt.Errorf("the history does not say %s was added for %s:\n%s", span, day, w.Out.Stdout)
	})

	sc.Step(`^the JSON gives the day "([^"]*)" task (\d+) at (\d+) minutes$`, func(day string, n, minutes int) error {
		var sheet struct {
			Days []struct {
				Day   string `json:"day"`
				Tasks []struct {
					Number  int `json:"number"`
					Minutes int `json:"minutes"`
				} `json:"tasks"`
			} `json:"days"`
		}
		if err := json.Unmarshal([]byte(w.Out.Stdout), &sheet); err != nil {
			return err
		}
		for _, d := range sheet.Days {
			for _, t := range d.Tasks {
				if d.Day == day && t.Number == n {
					if t.Minutes != minutes {
						return fmt.Errorf("the JSON gives %d minutes, want %d", t.Minutes, minutes)
					}
					return nil
				}
			}
		}
		return fmt.Errorf("the JSON says nothing of task %d on %s:\n%s", n, day, w.Out.Stdout)
	})
}
