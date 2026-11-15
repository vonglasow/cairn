package label

import (
	"errors"
	"testing"
)

func TestParseLabels_lowercasesEveryLabel(t *testing.T) {
	got, err := ParseLabels([]string{"Spec"})
	if err != nil || len(got) != 1 || got[0] != "spec" {
		t.Errorf("ParseLabels = %v, %v; want [spec]", got, err)
	}
}

func TestParseLabels_sortsThemAlphabetically(t *testing.T) {
	got, err := ParseLabels([]string{"team", "spec"})
	if err != nil || len(got) != 2 || got[0] != "spec" || got[1] != "team" {
		t.Errorf("ParseLabels = %v, %v; want [spec team]", got, err)
	}
}

func TestParseLabels_countsTheSameLabelOnce(t *testing.T) {
	got, err := ParseLabels([]string{"Spec", "spec"})
	if err != nil || len(got) != 1 {
		t.Errorf("ParseLabels = %v, %v; want one label", got, err)
	}
}

func TestParseLabels_refusesALabelWithNoName(t *testing.T) {
	if _, err := ParseLabels([]string{"  "}); err == nil {
		t.Error("ParseLabels should refuse a label that is only spaces")
	}
}

func TestParseLabels_readsASingleEmptyValueAsNoLabelAtAll(t *testing.T) {
	got, err := ParseLabels([]string{""})
	if err != nil || len(got) != 0 {
		t.Errorf("ParseLabels = %v, %v; want no label", got, err)
	}
}

func TestParseLabels_readsOnPastALabelNamedAsNothing(t *testing.T) {
	if got, err := ParseLabels([]string{"", "spec"}); err != nil || len(got) != 1 || got[0] != "spec" {
		t.Errorf("ParseLabels = %v, %v; want [spec]", got, err)
	}
	change, err := ParseLabelChange([]string{"", "+spec"})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := change.Apply(nil); len(got) != 1 || got[0] != "spec" {
		t.Errorf("Apply = %v, want [spec]: the change after the empty piece was dropped", got)
	}
}

func TestFolded_readsANameWhateverItsCaseAndAccents(t *testing.T) {
	if Folded(" Grocéries ") != Folded("groceries") || Folded("Home automation") != "home-automation" {
		t.Errorf("folded: %q, %q", Folded(" Grocéries "), Folded("Home automation"))
	}
	if Folded("Groceries") == Folded("Grocery") {
		t.Errorf("two names folded the same")
	}
}

func TestParseLabels_refusesALabelThatIsNotOneLine(t *testing.T) {
	for _, label := range []string{"two\nlines", "esc\x1b[2J", "tab\tbed"} {
		if _, err := ParseLabels([]string{label}); !errors.Is(err, ErrLabelNotOneLine) {
			t.Errorf("ParseLabels(%q) = %v, want it refused", label, err)
		}
		if _, err := ParseLabelChange([]string{"+" + label}); !errors.Is(err, ErrLabelNotOneLine) {
			t.Errorf("ParseLabelChange(+%q) = %v, want it refused", label, err)
		}
	}
}
