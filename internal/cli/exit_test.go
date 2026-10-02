package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

func TestCodeFor_saysNothingWentWrong(t *testing.T) {
	if got := codeFor(nil); got != exitOK {
		t.Errorf("codeFor(nil) = %d, want %d", got, exitOK)
	}
}

func TestCodeFor_tellsAStoreItCannotUse(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", engine.ErrStore)
	if got := codeFor(err); got != exitStore {
		t.Errorf("codeFor = %d, want %d for a store it cannot use", got, exitStore)
	}
}

func TestCodeFor_tellsARequestItUnderstoodAndRefused(t *testing.T) {
	for name, err := range map[string]error{
		"no title":      domain.ErrTitleRequired,
		"no label":      domain.ErrLabelRequired,
		"not a column":  domain.ErrNotAState,
		"no such task":  engine.ErrNoSuchTask,
		"nothing named": engine.ErrNothingToChange,
	} {
		if got := codeFor(fmt.Errorf("wrapped: %w", err)); got != exitRefused {
			t.Errorf("codeFor(%s) = %d, want %d", name, got, exitRefused)
		}
	}
}

func TestCodeFor_tellsARequestItDidNotUnderstand(t *testing.T) {
	if got := codeFor(errors.New("unknown flag: --colour")); got != exitUsage {
		t.Errorf("codeFor = %d, want %d for something it could not parse", got, exitUsage)
	}
}

// everyRefusalIsListed walks the source of every package that can refuse a
// request and checks each error it declares is either in the refusals list or
// named here as something else. A refusal left out of the list leaves the shell
// with code 2 — "I did not understand you" — for a request that was understood
// perfectly well and simply could not be granted.
var notARefusal = map[string]string{
	"ErrStore":       "the store itself is unusable: code 3",
	"ErrStoreBehind": "the store itself is unusable: code 3",
	"ErrNewerStore":  "the store itself is unusable: code 3",
	"ErrNotAStore":   "the store itself is unusable: code 3",
}

func TestEveryRefusalHasItsCode(t *testing.T) {
	listed, err := os.ReadFile("exit.go")
	if err != nil {
		t.Fatal(err)
	}
	named := regexp.MustCompile(`(?m)^(?:var )?\t?(Err[A-Za-z]+) += +(?:errors\.New|fmt\.Errorf)`)
	for _, where := range []string{"domain", "engine", "objective", "task", "store", "config", "mcp", "tui"} {
		files, err := filepath.Glob(filepath.Join("..", where, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			if strings.HasSuffix(file, "_test.go") {
				continue
			}
			source, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			for _, found := range named.FindAllStringSubmatch(string(source), -1) {
				name := found[1]
				if _, fine := notARefusal[name]; fine {
					continue
				}
				if !strings.Contains(string(listed), "."+name+",") {
					t.Errorf("%s declares %s and exit.go does not say what code it is", file, name)
				}
			}
		}
	}
}
