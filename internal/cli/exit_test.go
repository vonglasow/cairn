package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	if got := codeFor(engine.Misread(errors.New("unknown flag: --colour"))); got != exitUsage {
		t.Errorf("codeFor = %d, want %d for something it could not parse", got, exitUsage)
	}
}

// Kept for a request understood and still refused, and for whatever nobody
// foresaw: a failure is "could not do it" unless something said it was not
// understood — a parse error that has to be guessed at is a guess.
func TestCodeFor_tellsAFailureNobodyForesawAsOne(t *testing.T) {
	if got := codeFor(errors.New("the disk is on fire")); got != exitRefused {
		t.Errorf("codeFor = %d, want %d for a failure nobody foresaw", got, exitRefused)
	}
}

func TestMain_tellsEveryRequestItDidNotUnderstand(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	for _, args := range [][]string{
		{"nope"},
		{"task", "nope"},
		{"task", "show"},
		{"task", "show", "1", "2"},
		{"task", "list", "--colour"},
		{"task", "show", "four"},
	} {
		if got := run(t, args...).Code; got != exitUsage {
			t.Errorf("%v: exit code = %d, want %d", args, got, exitUsage)
		}
	}
}

func TestMain_answersARefusalAsJSONWithItsKindAndReason(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	var got engine.Refusal
	if err := json.Unmarshal([]byte(run(t, "task", "show", "1", "--json").Stdout), &got); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	if got.Kind != engine.KindRefused || got.Reason != "no-such-task" {
		t.Errorf("refusal = %+v, want refused/no-such-task", got)
	}
}

// What is written after `--` is a value, whatever it reads like.
func TestAskedForJSON_readsNothingAfterTheDoubleDash(t *testing.T) {
	if askedForJSON([]string{"task", "new", "--", "--json"}) {
		t.Error("a title written after -- was read as asking for JSON")
	}
}

// failing is a stream nothing can be written to.
type failing struct{}

func (failing) Write([]byte) (int, error) { return 0, errors.New("the pipe is closed") }

// A refusal that cannot be written as JSON is still said, where a person
// reads: a machine reading an empty stream learns nothing either way.
func TestMain_saysARefusalItCouldNotWriteAsJSON(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	var stderr bytes.Buffer
	code := Main(context.Background(), []string{"task", "show", "1", "--json"}, strings.NewReader(""), failing{}, &stderr)
	if code == exitOK || !strings.Contains(stderr.String(), "no such task") {
		t.Errorf("exit %d, stderr %q: want the refusal said on standard error", code, stderr.String())
	}
}

// A deletion refused for want of --yes names what it would have deleted.
func TestMain_namesWhatAnUnconfirmedDeletionWouldTake(t *testing.T) {
	t.Setenv("CAIRN_STORE", storeFor(t))
	mustRun(t, "task", "new", "Write the spec")
	r := run(t, "task", "delete", "1")
	if r.Code != exitRefused || !strings.Contains(strings.ToLower(r.Stderr), "task 1") || !strings.Contains(r.Stderr, "--yes") {
		t.Errorf("exit %d, stderr %q: want the task named and --yes asked for", r.Code, r.Stderr)
	}
}
