package test

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/cucumber/godog"
)

func TestSplitArgs_readsACommandLineTheWayAShellWould(t *testing.T) {
	for _, tc := range []struct {
		line string
		want []string
	}{
		{`cairn task new "Write the spec"`, []string{"cairn", "task", "new", "Write the spec"}},
		{`cairn task new ""`, []string{"cairn", "task", "new", ""}},
		{`cairn task new "   "`, []string{"cairn", "task", "new", "   "}},
		{`cairn task new Write the spec`, []string{"cairn", "task", "new", "Write", "the", "spec"}},
		{`cairn task new -- "--not-a-flag"`, []string{"cairn", "task", "new", "--", "--not-a-flag"}},
		{`cairn task new "a" "b"`, []string{"cairn", "task", "new", "a", "b"}},
		{`cairn task list`, []string{"cairn", "task", "list"}},
		{`cairn task new "Spec" --desc the one`, []string{"cairn", "task", "new", "Spec", "--desc", "the one"}},
		{`cairn task new "Spec" --priority high --state todo`, []string{"cairn", "task", "new", "Spec", "--priority", "high", "--state", "todo"}},
		// A flag carrying nothing keeps nothing: a value invented here would
		// reach the surface as an empty string it never had to refuse.
		{`cairn task list --json`, []string{"cairn", "task", "list", "--json"}},
		{`cairn task delete 1 --yes`, []string{"cairn", "task", "delete", "1", "--yes"}},
	} {
		if got := SplitArgs(tc.line); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("SplitArgs(%q) = %#v, want %#v", tc.line, got, tc.want)
		}
	}
}

func TestSplitArgs_answersNothingForAnEmptyLine(t *testing.T) {
	if got := SplitArgs(""); len(got) != 0 {
		t.Errorf("SplitArgs(\"\") = %#v, want nothing", got)
	}
}

func TestFlatten_dropsTheTypography(t *testing.T) {
	if got := Flatten("  A  Title\nOver Two Lines\t"); got != "a title over two lines" {
		t.Errorf("Flatten = %q", got)
	}
}

func TestWantContains_findsWhatIsThere(t *testing.T) {
	if err := WantContains("A Task Needs\n a Title", "a task needs a title", "the reason"); err != nil {
		t.Errorf("WantContains = %v, want it found past the typography", err)
	}
}

func TestWantContains_saysWhatIsMissingAndWhere(t *testing.T) {
	err := WantContains("something else", "a task needs a title", "the reason")
	if err == nil {
		t.Fatal("WantContains = nil, want it to report what is missing")
	}
	for _, want := range []string{"the reason", "a task needs a title", "something else"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("%q does not carry %q", err, want)
		}
	}
}

func TestCountsTasks_countsWhatIsListed(t *testing.T) {
	if err := CountsTasks("#1  One\n#2  Two\n", 2); err != nil {
		t.Errorf("CountsTasks = %v, want 2 counted", err)
	}
}

func TestCountsTasks_refusesADifferentCount(t *testing.T) {
	if err := CountsTasks("#1  One\n", 2); err == nil {
		t.Error("CountsTasks = nil, want it to report the count it found")
	}
}

func TestListsTask_findsATaskUnderItsNumber(t *testing.T) {
	if err := ListsTask("#1  Write the spec\n", 1, "Write the spec"); err != nil {
		t.Errorf("ListsTask = %v, want it found", err)
	}
}

func TestListsTask_refusesATaskThatIsNotThere(t *testing.T) {
	if err := ListsTask("#1  Write the spec\n", 2, "Write the spec"); err == nil {
		t.Error("ListsTask = nil, want it to report what is missing")
	}
}

func TestSuite_readsTheFeatureFromTheRoot(t *testing.T) {
	suite := Suite(t, "task.feature", func(*godog.ScenarioContext) {}, "@a-tag")
	if got := suite.Options.Paths; len(got) != 1 || got[0] != filepath.Join("..", "..", "bdd", "task.feature") {
		t.Errorf("Paths = %#v, want the feature two levels up", got)
	}
	if suite.Options.Tags != "@a-tag" {
		t.Errorf("Tags = %q, want the tag it was handed", suite.Options.Tags)
	}
	if !suite.Options.Strict {
		t.Error("Strict is off: an undefined step would pass unnoticed")
	}
}
