package calendar

import (
	"errors"
	"strings"
	"testing"
	"time"
)

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

func TestParseMoment_readsBackWhatFormatMomentWrote(t *testing.T) {
	at := time.Date(2026, 10, 1, 9, 0, 0, 0, time.Local)
	back, err := ParseMoment(FormatMoment(at))
	if err != nil || !back.Equal(at) {
		t.Errorf("ParseMoment = %v, %v, want %v", back, err, at)
	}
	if _, err := ParseMoment("yesterday"); err == nil {
		t.Error("a word was read as a moment")
	}
}

func TestARefusalOfADaySaysWhatWouldHaveBeenTaken(t *testing.T) {
	for _, c := range []struct {
		what string
		err  error
		says []string
	}{
		{"deadline", refusedOf(ParseDeadline("nope")), []string{"2006-01-02"}},
		{"estimate", refusedOf(ParseEstimate("nope")), []string{"1h30m", "2h"}},
	} {
		t.Run(c.what, func(t *testing.T) {
			if c.err == nil {
				t.Fatal("nothing was refused")
			}
			for _, word := range c.says {
				if !strings.Contains(c.err.Error(), word) {
					t.Errorf("%q does not say %q is taken", c.err, word)
				}
			}
		})
	}
}

func refusedOf[T any](_ T, err error) error { return err }
