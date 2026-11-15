package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cucumber/godog"

	"cairn/internal/engine/task"
)

func (w *world) registerBlocking(sc *godog.ScenarioContext) {
	sc.Step(`^the JSON row of task (\d+) gives "([^"]*)" as (.+)$`, func(n int, field, want string) error {
		var listed struct {
			Tasks []map[string]json.RawMessage `json:"tasks"`
		}
		if err := json.Unmarshal([]byte(w.Out.Stdout), &listed); err != nil {
			return fmt.Errorf("%w:\n%s", err, w.Out.Stdout)
		}
		for _, row := range listed.Tasks {
			if string(row["number"]) != fmt.Sprint(n) {
				continue
			}
			if got := string(row[field]); got != want {
				return fmt.Errorf("task %d gives %s as %s, want %s", n, field, got, want)
			}
			return nil
		}
		return fmt.Errorf("no row of task %d:\n%s", n, w.Out.Stdout)
	})
	sc.Step(`^it marks "([^"]*)" as having nothing in its way$`, func(title string) error {
		for _, line := range strings.Split(w.Out.Stdout, "\n") {
			if strings.Contains(line, title) && strings.Contains(line, task.NothingBlocksMark) {
				return nil
			}
		}
		return fmt.Errorf("%q is not marked as having nothing in its way:\n%s", title, w.Out.Stdout)
	})

}
