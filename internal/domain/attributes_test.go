package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestParseState_acceptsTheFourColumns(t *testing.T) {
	for _, name := range []string{"backlog", "todo", "current", "done"} {
		got, err := ParseState(name)
		if err != nil {
			t.Errorf("ParseState(%q): %v", name, err)
			continue
		}
		if string(got) != name {
			t.Errorf("ParseState(%q) = %q", name, got)
		}
	}
}

func TestParseState_refusesAnythingElse(t *testing.T) {
	if _, err := ParseState("someday"); err == nil {
		t.Error("ParseState(someday) should fail: it is not one of the columns")
	}
}

func TestParseState_readsAnEmptyValueAsTheBacklog(t *testing.T) {
	got, err := ParseState("")
	if err != nil || got != Backlog {
		t.Errorf("ParseState(\"\") = %q, %v; want the backlog", got, err)
	}
}

func TestParsePriority_acceptsTheFourPriorities(t *testing.T) {
	for _, name := range []string{"low", "normal", "high", "critical"} {
		got, err := ParsePriority(name)
		if err != nil {
			t.Errorf("ParsePriority(%q): %v", name, err)
			continue
		}
		if string(got) != name {
			t.Errorf("ParsePriority(%q) = %q", name, got)
		}
	}
}

func TestParsePriority_refusesAnythingElse(t *testing.T) {
	if _, err := ParsePriority("urgent"); err == nil {
		t.Error("ParsePriority(urgent) should fail: it is not one of the priorities")
	}
}

func TestParsePriority_readsAnEmptyValueAsNormal(t *testing.T) {
	got, err := ParsePriority("")
	if err != nil || got != Normal {
		t.Errorf("ParsePriority(\"\") = %q, %v; want normal", got, err)
	}
}

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

func TestParseDeadline_acceptsADay(t *testing.T) {
	got, err := ParseDeadline("2026-09-01")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}
	if got.Format(dayLayout) != "2026-09-01" {
		t.Errorf("ParseDeadline = %v, want 2026-09-01", got)
	}
}

func TestParseDeadline_refusesWhatIsNotADate(t *testing.T) {
	if _, err := ParseDeadline("soon"); err == nil {
		t.Error("ParseDeadline(soon) should fail")
	}
}

func TestParseDeadline_refusesADayThatCannotExist(t *testing.T) {
	if _, err := ParseDeadline("2026-13-45"); err == nil {
		t.Error("ParseDeadline(2026-13-45) should fail")
	}
}

func TestParseDeadline_readsAnEmptyValueAsNoDeadline(t *testing.T) {
	got, err := ParseDeadline("")
	if err != nil || !got.IsZero() {
		t.Errorf("ParseDeadline(\"\") = %v, %v; want no deadline", got, err)
	}
}

func TestParseEstimate_normalisesWhatItIsGiven(t *testing.T) {
	got, err := ParseEstimate("90m")
	if err != nil {
		t.Fatalf("ParseEstimate: %v", err)
	}
	if got != 90*time.Minute {
		t.Errorf("ParseEstimate = %v, want 90m", got)
	}
}

func TestParseEstimate_refusesWhatIsNotADuration(t *testing.T) {
	if _, err := ParseEstimate("ages"); err == nil {
		t.Error("ParseEstimate(ages) should fail")
	}
}

func TestParseEstimate_refusesANegativeDuration(t *testing.T) {
	if _, err := ParseEstimate("-30m"); err == nil {
		t.Error("ParseEstimate(-30m) should fail: time does not run backwards")
	}
}

func TestParseEstimate_readsAnEmptyValueAsNoEstimate(t *testing.T) {
	got, err := ParseEstimate("")
	if err != nil || got != 0 {
		t.Errorf("ParseEstimate(\"\") = %v, %v; want no estimate", got, err)
	}
}

func TestFormatDuration_dropsTheEmptyTail(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"90m", "1h30m"},
		{"1h", "1h"},
		{"45m", "45m"},
		{"0", ""},
	} {
		d, err := ParseEstimate(tc.in)
		if err != nil {
			t.Fatalf("ParseEstimate(%q): %v", tc.in, err)
		}
		if got := FormatDuration(d); got != tc.want {
			t.Errorf("FormatDuration(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFormatDuration_writesSecondsWhenThereAreSome(t *testing.T) {
	if got := FormatDuration(90 * time.Second); got != "1m30s" {
		t.Errorf("FormatDuration(90s) = %q, want 1m30s", got)
	}
}

func TestFormatDay_writesTheDayItWasGiven(t *testing.T) {
	day, err := ParseDeadline("2026-09-01")
	if err != nil {
		t.Fatalf("ParseDeadline: %v", err)
	}
	if got := FormatDay(day); got != "2026-09-01" {
		t.Errorf("FormatDay = %q, want 2026-09-01", got)
	}
}

func TestFormatDay_writesNothingWhenThereIsNoDeadline(t *testing.T) {
	if got := FormatDay(time.Time{}); got != "" {
		t.Errorf("FormatDay(zero) = %q, want nothing", got)
	}
}

// cairn writes a duration down to the second and no finer, so anything that
// is not a whole number of seconds cannot be read back as it was given.
// Refusing it is the only answer that does not lose it in silence: 500ms was
// accepted, stored, and came back as nothing at all.
func TestParseEstimate_refusesWhatItCannotWriteBack(t *testing.T) {
	for _, amount := range []string{"500ms", "1m30s500ms", "1ns", "1500ms"} {
		if _, err := ParseEstimate(amount); !errors.Is(err, ErrFinerThanASecond) {
			t.Errorf("ParseEstimate(%q) err = %v, want ErrFinerThanASecond", amount, err)
		}
	}
}

func TestParseEstimate_keepsWholeSeconds(t *testing.T) {
	for _, amount := range []string{"30s", "1m30s", "90m", "1.5h", "0s"} {
		if _, err := ParseEstimate(amount); err != nil {
			t.Errorf("ParseEstimate(%q): %v", amount, err)
		}
	}
}

func TestParseEstimate_saysWhatItWasGiven(t *testing.T) {
	_, err := ParseEstimate("500ms")
	if got := err.Error(); !strings.Contains(got, `"500ms"`) {
		t.Errorf("err = %q, want it to quote what it was given", got)
	}
}
