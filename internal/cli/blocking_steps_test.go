package cli

// Executable acceptance scenarios for what blocks a task. The steps every
// feature shares live in world_test.go; only what belongs to blocking is
// declared here.

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/task"
)

// registerBlocking declares the steps about what stops a task and about the
// mark that says nothing does any more.
func (w *world) registerBlocking(sc *godog.ScenarioContext) {
	sc.Step(`^it marks "([^"]*)" as having nothing in its way$`, func(title string) error {
		for _, line := range strings.Split(w.Out.Stdout, "\n") {
			if strings.Contains(line, title) && strings.Contains(line, task.NothingBlocksMark) {
				return nil
			}
		}
		return fmt.Errorf("%q is not marked as having nothing in its way:\n%s", title, w.Out.Stdout)
	})

	// The other half of the mark. Marking the right task proves nothing on its
	// own: the answer would read the same with every task marked, and the cases
	// where cairn must not claim anything are most of them.
	sc.Step(`^it marks nothing as having nothing in its way$`, func() error {
		if marked := strings.Count(w.Out.Stdout, task.NothingBlocksMark); marked != 0 {
			return fmt.Errorf("%d lines say nothing is in the way, want none:\n%s",
				marked, w.Out.Stdout)
		}
		return nil
	})

	// The card lays out every attribute a task carries, empty ones included, so
	// nothing blocking it is the row reading empty rather than the row being
	// gone. Asserted on the row and not on its absence: a card that stopped
	// laying out the row at all would leave a reader wondering.
	sc.Step(`^it says nothing blocks it$`, func() error {
		if got := w.field("blocked by"); got != "" {
			return fmt.Errorf("what blocks it is %q, want nothing", got)
		}
		return nil
	})

	// What crosses to a machine, read out of the JSON rather than off the card:
	// the two readings are answered by the same engine, and a scenario that
	// checked the card twice would prove the wire nothing.
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
