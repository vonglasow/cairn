// Package test carries the acceptance harness the surfaces share: the world one scenario works against, and the runner that turns a feature file into Go tests.
package test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

// Result is what one run of a command line produced: the two streams kept apart, and the code the process would have exited with.
type Result struct {
	Stdout string
	Stderr string
	Code   int
}

// A Runner runs a command line in-process and answers what it produced.
type Runner func(t *testing.T, args ...string) Result

const ok = 0

// Suite describes the run of the scenarios carrying tag from the named feature file.
func Suite(t *testing.T, feature string, init func(*godog.ScenarioContext), tag string) godog.TestSuite {
	return godog.TestSuite{
		Name:                "cairn",
		ScenarioInitializer: init,
		Options: &godog.Options{
			Format:         "pretty",
			Paths:          []string{filepath.Join("..", "..", "bdd", feature)},
			Tags:           tag,
			Strict:         true,
			NoColors:       true,
			TestingT:       t,
			DefaultContext: context.Background(),
		},
	}
}

// SplitArgs turns the command line a scenario wrote into the arguments a shell would have handed over: words separated by spaces, unless quoted.
func SplitArgs(line string) []string {
	var (
		args    []string
		current strings.Builder
		quoted  bool
		started bool
	)
	for _, r := range line {
		if r == '"' {
			quoted = !quoted
			started = true
			continue
		}
		if r == ' ' && !quoted {
			if started {
				args = append(args, current.String())
				current.Reset()
				started = false
			}
			continue
		}
		current.WriteRune(r)
		started = true
	}
	if started {
		args = append(args, current.String())
	}
	return joinFlagValues(args)
}

func joinFlagValues(args []string) []string {
	joined := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		joined = append(joined, args[i])
		if !strings.HasPrefix(args[i], "--") || args[i] == "--" {
			continue
		}
		value := []string{}
		for i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
			value = append(value, args[i+1])
			i++
		}
		if len(value) > 0 {
			joined = append(joined, strings.Join(value, " "))
		}
	}
	return joined
}

// Flatten makes a rendered message comparable: a surface may capitalise words and wrap long lines to the terminal width, so a step comparing exactly would be asserting typography rather than cairn's answer.
func Flatten(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}

// WantContains looks for what inside where, past that typography.
func WantContains(where, what, label string) error {
	if !strings.Contains(Flatten(where), Flatten(what)) {
		return fmt.Errorf("%s does not mention %q:\n%s", label, what, where)
	}
	return nil
}

// CountsTasks fails unless a listing carries exactly n tasks.
func CountsTasks(out string, n int) error {
	got := strings.Count(out, "#")
	if got != n {
		return fmt.Errorf("%d tasks listed, want %d:\n%s", got, n, out)
	}
	return nil
}

// ListsTask fails unless a listing carries the task numbered n under title.
func ListsTask(out string, n int, title string) error {
	want := fmt.Sprintf("#%d  %s", n, title)
	if !strings.Contains(out, want) {
		return fmt.Errorf("the listing does not carry %q:\n%s", want, out)
	}
	return nil
}
