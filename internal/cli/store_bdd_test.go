package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	"cairn/test"

	"cairn/internal/config"
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
			code := Main(context.Background(), args, strings.NewReader(""), &out, &errs)
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

func (w *world) registerAtOnce(sc *godog.ScenarioContext) {
	sc.Step(`^the store cannot be written to$`, func() error {
		if err := os.Chmod(w.StorePath, 0o444); err != nil {
			return err
		}
		w.OnEnd(func() { _ = os.Chmod(w.StorePath, 0o600) })
		return nil
	})

	sc.Step(`^the environment says nothing about where anything lives$`, func() error {
		w.Setenv(config.StoreVariable, "")
		w.Setenv("HOME", "")
		if err := os.Unsetenv(config.StoreVariable); err != nil {
			return err
		}
		return os.Unsetenv("HOME")
	})

	sc.Step(`^(\d+) commands add a task at the same time$`, func(n int) error {
		lines := make([][]string, 0, n)
		for i := 1; i <= n; i++ {
			lines = append(lines, []string{"task", "new", fmt.Sprintf("At once %d", i)})
		}
		code, said := atOnce(lines)
		if code != exitOK {
			return fmt.Errorf("one of them exited %d: %s", code, said)
		}
		w.Out = test.Result{Code: exitOK}
		return nil
	})

	sc.Step(`^(\d+) commands read and write at the same time$`, func(n int) error {
		lines := make([][]string, 0, n)
		for i := 1; i <= n; i++ {
			if i%2 == 0 {
				lines = append(lines, []string{"task", "new", fmt.Sprintf("Written %d", i)})
				continue
			}
			lines = append(lines, []string{"task", "list"})
		}
		code, said := atOnce(lines)
		if code != exitOK {
			return fmt.Errorf("one of them exited %d: %s", code, said)
		}
		w.Out = test.Result{Code: exitOK}
		return nil
	})
}
