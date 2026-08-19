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
