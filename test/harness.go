// Package test carries the acceptance harness the surfaces share: the world one
// scenario works against, and the runner that turns a feature file into Go
// tests.
//
// It knows no surface. A caller hands it a Runner — run this command line and
// tell me what came back — and the harness never learns who answers. That is
// what lets a surface keep its own scenarios in its own package: were this
// package to import a surface, that surface's scenarios would import a package
// which imports them back, and Go refuses the cycle.
//
// It imports testing outside a _test.go file, on purpose, because Go compiles
// a package's _test.go files only for that package: there is no other way to
// hand the same world to two of them. What holds the line instead is narrower
// and truer than "nothing imports testing" — nothing the binary is made of may
// reach here, and `make harness-check` is what proves it.
package test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

// Result is what one run of a command line produced: the two streams kept
// apart, and the code the process would have exited with.
//
// The streams are separate on purpose. A result goes to standard output and a
// reason goes to standard error; merged into one buffer, a message on the wrong
// stream reads exactly like a message on the right one, and no test could ever
// tell them apart.
type Result struct {
	Stdout string
	Stderr string
	Code   int
}

// A Runner runs a command line in-process and answers what it produced. It is
// how a world reaches a surface without being told which one it is.
type Runner func(t *testing.T, args ...string) Result

// ok is the code a process reports when nothing went wrong. Which of the others
// means what is a surface's own table, and this harness may not read one; that
// zero means "it worked" is a convention no surface owns.
const ok = 0

// Suite describes the run of the scenarios carrying tag from the named feature
// file.
//
// Every feature wires itself through here rather than backing up the options:
// one description, one place to fix. A feature adds its own initializer — steps
// live in the package that observes them — and nothing else. Running it is left
// to the caller, so that whose test fails is the surface's own business.
func Suite(t *testing.T, feature string, init func(*godog.ScenarioContext), tag string) godog.TestSuite {
	return godog.TestSuite{
		Name:                "cairn",
		ScenarioInitializer: init,
		Options: &godog.Options{
			Format: "pretty",
			// Relative to the package directory: `go test` runs there, and every
			// package holding scenarios sits two levels under the root.
			Paths:          []string{filepath.Join("..", "..", "bdd", feature)},
			Tags:           tag,
			Strict:         true,
			NoColors:       true,
			TestingT:       t,
			DefaultContext: context.Background(),
		},
	}
}

// SplitArgs turns the command line a scenario wrote into the arguments a shell
// would have handed over: words separated by spaces, unless quoted. It is
// deliberately small — a scenario writes commands a person types, not shell
// programs.
func SplitArgs(line string) []string {
	var (
		args    []string
		current strings.Builder
		quoted  bool
		started bool
	)
	// Written with ifs rather than a switch on purpose: a tagless switch puts
	// its conditions outside any covered block, so a mutation of one of them
	// reports as unreachable and the gate cannot judge it.
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

// joinFlagValues puts back together a flag value that carried a space. A
// scenario's Examples table cannot quote its cells, so `--desc the one` arrives
// as three words where a shell would have been handed two. Everything between a
// flag and the next one is that flag's value.
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

// Flatten makes a rendered message comparable: a surface may capitalise words
// and wrap long lines to the terminal width, so a step comparing exactly would
// be asserting typography rather than cairn's answer.
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
