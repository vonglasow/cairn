package cli

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/cucumber/godog"
)

func (w *world) registerLinking(sc *godog.ScenarioContext) {
	sc.Step(`^it carries the link to task (\d+)$`, func(n int) error {
		var said struct {
			Linked []int `json:"linked"`
		}
		if err := json.Unmarshal([]byte(w.Out.Stdout), &said); err != nil {
			return fmt.Errorf("reading the links: %w\n%s", err, w.Out.Stdout)
		}
		if !slices.Contains(said.Linked, n) {
			return fmt.Errorf("the task is linked to %v, not to task %d", said.Linked, n)
		}
		return nil
	})
}
