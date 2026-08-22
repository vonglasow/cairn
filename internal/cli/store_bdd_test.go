package cli

// Executable acceptance scenarios for bdd/store.feature: what happens when
// more than one cairn wants the same file at the same time.

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/cucumber/godog"
)

func initStore(t *testing.T) func(*godog.ScenarioContext) {
	return func(sc *godog.ScenarioContext) {
		w := newWorld(t)
		w.register(sc)
		w.registerAtOnce(sc)
	}
}

func TestAcceptance_Store(t *testing.T) {
	runAcceptance(t, "store.feature", initStore(t), "")
}

// atOnce runs commands together and hands back the worst exit code any of
// them returned. Main builds a fresh command tree per call and writes into
// the buffers it is given, so nothing is shared but the store on disk —
// which is the whole point of the scenario.
func atOnce(lines [][]string) (int, string) {
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		worst int
		said  string
	)
	for _, args := range lines {
		wg.Add(1)
		go func(args []string) {
			defer wg.Done()
			var out, errs bytes.Buffer
			code := Main(context.Background(), args, &out, &errs)
			mu.Lock()
			defer mu.Unlock()
			if code > worst {
				worst = code
				said = errs.String()
			}
		}(args)
	}
	wg.Wait()
	return worst, said
}

// registerAtOnce declares the steps only this feature uses.
func (w *world) registerAtOnce(sc *godog.ScenarioContext) {
	sc.Step(`^(\d+) commands add a task at the same time$`, func(n int) error {
		lines := make([][]string, 0, n)
		for i := 1; i <= n; i++ {
			lines = append(lines, []string{"task", "add", fmt.Sprintf("At once %d", i)})
		}
		code, said := atOnce(lines)
		if code != exitOK {
			return fmt.Errorf("one of them exited %d: %s", code, said)
		}
		w.out = result{code: exitOK}
		return nil
	})

	// Reads mixed in with writes: a reader holding the file is exactly what
	// used to turn a write into a refusal.
	sc.Step(`^(\d+) commands read and write at the same time$`, func(n int) error {
		lines := make([][]string, 0, n)
		for i := 1; i <= n; i++ {
			if i%2 == 0 {
				lines = append(lines, []string{"task", "add", fmt.Sprintf("Written %d", i)})
				continue
			}
			lines = append(lines, []string{"task", "list"})
		}
		code, said := atOnce(lines)
		if code != exitOK {
			return fmt.Errorf("one of them exited %d: %s", code, said)
		}
		w.out = result{code: exitOK}
		return nil
	})
}
