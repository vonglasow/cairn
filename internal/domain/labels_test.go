package domain

import (
	"errors"
	"strings"
	"testing"
)

func applied(t *testing.T, current []string, given ...string) []string {
	t.Helper()
	change, err := ParseLabelChange(given)
	if err != nil {
		t.Fatalf("ParseLabelChange(%v): %v", given, err)
	}
	got, err := change.Apply(current)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	return got
}

func TestLabelChange_replacesTheSetWhenLabelsAreGivenPlainly(t *testing.T) {
	if got := applied(t, []string{"spec", "team"}, "draft"); strings.Join(got, ",") != "draft" {
		t.Errorf("got %v, want [draft]: a plain label says what the set becomes", got)
	}
}

func TestLabelChange_addsALabelGivenWithAPlus(t *testing.T) {
	if got := applied(t, []string{"spec", "team"}, "+urgent"); strings.Join(got, ",") != "spec,team,urgent" {
		t.Errorf("got %v, want [spec team urgent]", got)
	}
}

func TestLabelChange_takesAwayALabelGivenWithAMinus(t *testing.T) {
	if got := applied(t, []string{"spec", "team"}, "-spec"); strings.Join(got, ",") != "team" {
		t.Errorf("got %v, want [team]", got)
	}
}

func TestLabelChange_addsAndTakesAwayInTheSameBreath(t *testing.T) {
	if got := applied(t, []string{"spec", "team"}, "+urgent", "-spec"); strings.Join(got, ",") != "team,urgent" {
		t.Errorf("got %v, want [team urgent]", got)
	}
}

func TestLabelChange_addingOneAlreadyThereChangesNothing(t *testing.T) {
	if got := applied(t, []string{"spec", "team"}, "+spec"); len(got) != 2 {
		t.Errorf("got %v, want two labels", got)
	}
}

func TestLabelChange_takingAwayOneThatIsNotThereChangesNothing(t *testing.T) {
	if got := applied(t, []string{"spec", "team"}, "-draft"); len(got) != 2 {
		t.Errorf("got %v, want two labels", got)
	}
}

func TestLabelChange_lowercasesWhatItAdds(t *testing.T) {
	if got := applied(t, nil, "+Urgent"); strings.Join(got, ",") != "urgent" {
		t.Errorf("got %v, want [urgent]", got)
	}
}

func TestLabelChange_lowercasesWhatItTakesAway(t *testing.T) {
	if got := applied(t, []string{"urgent"}, "-URGENT"); len(got) != 0 {
		t.Errorf("got %v, want none: case is not what tells two labels apart", got)
	}
}

func TestLabelChange_refusesToReplaceAndChangeAtOnce(t *testing.T) {
	if _, err := ParseLabelChange([]string{"+urgent", "draft"}); !errors.Is(err, ErrLabelsBothWays) {
		t.Errorf("err = %v, want ErrLabelsBothWays", err)
	}
}

func TestLabelChange_refusesASignWithNoName(t *testing.T) {
	for _, given := range []string{"+", "-", "+  ", "-  "} {
		if _, err := ParseLabelChange([]string{given}); !errors.Is(err, ErrLabelRequired) {
			t.Errorf("ParseLabelChange(%q) err = %v, want ErrLabelRequired", given, err)
		}
	}
}

func TestLabelChange_knowsWhetherItChangesOrReplaces(t *testing.T) {
	change, err := ParseLabelChange([]string{"+urgent"})
	if err != nil {
		t.Fatalf("ParseLabelChange: %v", err)
	}
	if !change.Adjusts() {
		t.Error("a signed label adjusts an existing set")
	}

	change, err = ParseLabelChange([]string{"urgent"})
	if err != nil {
		t.Fatalf("ParseLabelChange: %v", err)
	}
	if change.Adjusts() {
		t.Error("a plain label replaces the set, it does not adjust it")
	}
}

func TestLabelChange_clearsTheSetWhenGivenNothing(t *testing.T) {
	if got := applied(t, []string{"spec", "team"}, ""); len(got) != 0 {
		t.Errorf("got %v, want none", got)
	}
}

func TestParseLabels_splitsOnCommas(t *testing.T) {
	got, err := ParseLabels([]string{"spec,team"})
	if err != nil {
		t.Fatalf("ParseLabels: %v", err)
	}
	if len(got) != 2 || got[0] != "spec" || got[1] != "team" {
		t.Errorf("got %v, want [spec team]", got)
	}
}

func TestParseLabels_ignoresEmptyPlacesInAList(t *testing.T) {
	got, err := ParseLabels([]string{"spec,,team,"})
	if err != nil {
		t.Fatalf("ParseLabels: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %v, want two labels", got)
	}
}

func TestParseLabelChange_splitsOnCommas(t *testing.T) {
	change, err := ParseLabelChange([]string{"+urgent,+draft"})
	if err != nil {
		t.Fatalf("ParseLabelChange: %v", err)
	}
	got, err := change.Apply([]string{"spec", "team"})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if len(got) != 4 {
		t.Errorf("got %v, want four labels", got)
	}
}

func TestParseLabelChange_refusesAListThatMixesTheTwoForms(t *testing.T) {
	if _, err := ParseLabelChange([]string{"+urgent,spec"}); !errors.Is(err, ErrLabelsBothWays) {
		t.Errorf("err = %v, want ErrLabelsBothWays", err)
	}
}

func TestParseLabels_writesALabelOneWay(t *testing.T) {
	for typed, kept := range map[string]string{
		"Sp26 20":           "sp26-20",
		"Modularité/Helm":   "modularite-helm",
		"dette__config":     "dette-config",
		"Release__26.1.6--": "release-26.1.6",
		"  ÉTÉ  2026 ":      "ete-2026",
	} {
		got, err := ParseLabels([]string{typed})
		if err != nil || len(got) != 1 || got[0] != kept {
			t.Errorf("ParseLabels(%q) = %v, %v; want [%s]", typed, got, err, kept)
		}
	}
}

func TestParseLabelChange_writesALabelJoiningOneWay(t *testing.T) {
	if got := applied(t, nil, "+Sp26 20"); len(got) != 1 || got[0] != "sp26-20" {
		t.Errorf("+Sp26 20 joins as %v, want [sp26-20]", got)
	}
}

func TestParseLabels_refusesAFamily(t *testing.T) {
	_, err := ParseLabels([]string{"ado:77594"})
	if !errors.Is(err, ErrLabelHasNoFamily) {
		t.Errorf("ParseLabels(ado:77594) = %v, want it refused for its family", err)
	}
	if err != nil && !strings.Contains(err.Error(), "two labels") {
		t.Errorf("the refusal %q does not say to write two labels", err)
	}
}

func TestParseLabels_refusesWhatNoLabelMayHold(t *testing.T) {
	for _, typed := range []string{"#urgent", "a+b", "ça,va"[0:3] + "!"} {
		if _, err := ParseLabels([]string{typed}); !errors.Is(err, ErrLabelHolds) {
			t.Errorf("ParseLabels(%q) = %v, want it refused for what it holds", typed, err)
		}
	}
}

func TestParseLabels_refusesALabelWrittenToNothing(t *testing.T) {
	if _, err := ParseLabels([]string{"--__--"}); !errors.Is(err, ErrLabelRequired) {
		t.Errorf("ParseLabels(--__--) = %v, want it refused for having no name", err)
	}
}

func TestLabelChange_leavesALabelFromBeforeTheRuleAlone(t *testing.T) {
	if got := applied(t, []string{"c++"}, "+spec"); len(got) != 2 || got[0] != "c++" || got[1] != "spec" {
		t.Errorf("+spec over [c++] gives %v, want [c++ spec]", got)
	}
}

func TestParseLabels_holdsTheEndsOfWhatALabelMayHold(t *testing.T) {
	if got, err := ParseLabels([]string{"az09"}); err != nil || len(got) != 1 || got[0] != "az09" {
		t.Errorf("ParseLabels(az09) = %v, %v; want [az09]", got, err)
	}
	for _, typed := range []string{"a{", "a`", "a@", "a;"} {
		if _, err := ParseLabels([]string{typed}); !errors.Is(err, ErrLabelHolds) {
			t.Errorf("ParseLabels(%q) = %v, want it refused for what it holds", typed, err)
		}
	}
}
