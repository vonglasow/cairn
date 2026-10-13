package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/task"
)

func (w *world) registerBlocking(sc *godog.ScenarioContext) {
	sc.Step(`^it marks "([^"]*)" as having nothing in its way$`, func(title string) error {
		for _, line := range strings.Split(w.Out.Stdout, "\n") {
			if strings.Contains(line, title) && strings.Contains(line, task.NothingBlocksMark) {
				return nil
			}
		}
		return fmt.Errorf("%q is not marked as having nothing in its way:\n%s", title, w.Out.Stdout)
	})

	sc.Step(`^it marks nothing as having nothing in its way$`, func() error {
		if marked := strings.Count(w.Out.Stdout, task.NothingBlocksMark); marked != 0 {
			return fmt.Errorf("%d lines say nothing is in the way, want none:\n%s",
				marked, w.Out.Stdout)
		}
		return nil
	})

	sc.Step(`^it says nothing blocks it$`, func() error {
		if got := w.field("blocked by"); got != "" {
			return fmt.Errorf("what blocks it is %q, want nothing", got)
		}
		return nil
	})

	sc.Step("^it says task (\\d+) blocks it and \"([^\"]*)\" does too$",
		func(n int, words string) error {
			var card struct {
				BlockedBy []int  `json:"blocked_by"`
				Because   string `json:"because"`
			}
			if err := json.Unmarshal([]byte(w.Out.Stdout), &card); err != nil {
				return fmt.Errorf("the card is not JSON (%w):\n%s", err, w.Out.Stdout)
			}
			if len(card.BlockedBy) != 1 || card.BlockedBy[0] != n {
				return fmt.Errorf("blocked_by = %v, want just task %d", card.BlockedBy, n)
			}
			if card.Because != words {
				return fmt.Errorf("because = %q, want %q", card.Because, words)
			}
			return nil
		})
}
