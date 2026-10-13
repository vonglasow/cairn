package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"cairn/internal/domain"
)

func TestRefusalOf_namesWhatKindOfAnswerAFailureIs(t *testing.T) {
	for _, c := range []struct {
		err          error
		kind, reason string
	}{
		{fmt.Errorf("a task %w", domain.ErrTitleRequired), "refused", "title-required"},
		{fmt.Errorf("task 4: %w", ErrNoSuchTask), "refused", "no-such-task"},
		{fmt.Errorf("%w — run it", ErrStoreBehind), "store", "store-behind"},
		{fmt.Errorf("%w: disk", ErrStore), "store", "store-unreadable"},
		{Misread(errors.New("unknown flag: --colour")), "usage", "not-understood"},
		{errors.New("something nobody foresaw"), "failed", ""},
	} {
		got := RefusalOf(c.err)
		if got.Kind != c.kind || got.Reason != c.reason {
			t.Errorf("RefusalOf(%v) = %s/%s, want %s/%s", c.err, got.Kind, got.Reason, c.kind, c.reason)
		}
		if got.Error != c.err.Error() {
			t.Errorf("Error = %q, want the message as it was", got.Error)
		}
	}
}

func TestRefusalOf_readsWhatCheckRefusesAsNotUnderstood(t *testing.T) {
	for what, in := range map[string]Input{
		"a name it does not take": {Given: map[string]string{"colour": "red"}},
		"a number that is not":    {Given: map[string]string{"number": "four"}},
		"nothing it needs":        {},
	} {
		err := TaskShow.Check(in)
		if got := RefusalOf(err); got.Kind != "usage" {
			t.Errorf("%s: RefusalOf(%v).Kind = %q, want usage", what, err, got.Kind)
		}
	}
}

func TestRefusalOf_readsAReasonASurfaceNamed(t *testing.T) {
	notConfirmed := Refusing("not-confirmed", "needs --yes to confirm")
	got := RefusalOf(fmt.Errorf("task 4 %w", notConfirmed))
	if got.Kind != "refused" || got.Reason != "not-confirmed" {
		t.Errorf("RefusalOf = %s/%s, want refused/not-confirmed", got.Kind, got.Reason)
	}
	if notConfirmed.Error() != "needs --yes to confirm" {
		t.Errorf("Error() = %q, want the sentence it was given", notConfirmed.Error())
	}
}

func TestReasons_areEachWrittenOnceAndSpelledAlike(t *testing.T) {
	spelled := regexp.MustCompile(`^[a-z]+(-[a-z]+)*$`)
	seen := map[string]bool{}
	for _, one := range reasons {
		if !spelled.MatchString(one.reason) {
			t.Errorf("reason %q is not lower-case words joined by hyphens", one.reason)
		}
		if seen[one.reason] {
			t.Errorf("reason %q is given to two refusals", one.reason)
		}
		seen[one.reason] = true
	}
}

func TestRefusalOf_readsWhatTheReadersRefuseAsNotUnderstood(t *testing.T) {
	_, typed := TaskShow.Read([]string{"4", "5"})
	_, flag := TaskNew.Read([]string{"x", "--title"})
	_, yes := TaskMove.Read([]string{"4", "--first=maybe"})
	_, sent := TaskNew.FromJSON([]byte(`{"label": "one"}`))
	_, yesSent := TaskMove.FromJSON([]byte(`{"first": "yes"}`))
	_, broken := TaskNew.FromJSON([]byte(`{"title": `))
	for what, err := range map[string]error{
		"a word too many": typed, "a flag given nothing": flag, "a yes that is not": yes,
		"a list sent as a string": sent, "a yes sent as a string": yesSent, "JSON that is not": broken,
	} {
		if got := RefusalOf(err); got.Kind != KindUsage {
			t.Errorf("%s: RefusalOf(%v).Kind = %q, want usage", what, err, got.Kind)
		}
	}
}

func TestEveryFailureCairnForesawHasAReason(t *testing.T) {
	listed, err := os.ReadFile("refusal.go")
	if err != nil {
		t.Fatal(err)
	}
	named := regexp.MustCompile(`(?m)^(?:var )?\t?(Err[A-Za-z]+) += +(?:errors\.New|fmt\.Errorf)`)
	for _, where := range []string{"domain", "engine", "objective", "task", "store", "config"} {
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
				if !regexp.MustCompile(`[{.]` + name + `,`).Match(listed) {
					t.Errorf("%s declares %s and refusal.go gives it no reason", file, name)
				}
			}
		}
	}
}

func TestRefusalOf_givesOneReasonForAValueOutsideItsSet(t *testing.T) {
	path := storeAt(t, "Write the spec")
	for _, c := range []struct {
		op     Operation
		in     Input
		reason string
	}{
		{TaskList, given("state", "bogus"), "not-a-state"},
		{TaskEdit, given("number", "1", "state", "bogus"), "not-a-state"},
		{TaskList, given("priority", "urgent"), "not-a-priority"},
		{TaskEdit, given("number", "1", "priority", "urgent"), "not-a-priority"},
		{ObjectiveList, given("state", "bogus"), "not-an-objective-state"},
	} {
		_, err := Do(c.op, path, c.in)
		if got := RefusalOf(err).Reason; got != c.reason {
			t.Errorf("%s %v: reason %q, want %q (%v)", c.op.Line(), c.in.Given, got, c.reason, err)
		}
	}
}
