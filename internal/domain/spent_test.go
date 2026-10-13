package domain

import (
	"errors"
	"testing"
	"time"
)

func spentAfter(t *testing.T, current time.Duration, given string) time.Duration {
	t.Helper()
	change, err := ParseSpent(given)
	if err != nil {
		t.Fatalf("ParseSpent(%q): %v", given, err)
	}
	return change.Apply(current)
}

func TestSpent_writesTheTotalWhenGivenPlainly(t *testing.T) {
	if got := spentAfter(t, 30*time.Minute, "2h"); got != 2*time.Hour {
		t.Errorf("got %v, want 2h: a plain duration says what the total becomes", got)
	}
}

func TestSpent_addsWhatIsGivenWithAPlus(t *testing.T) {
	if got := spentAfter(t, 2*time.Hour, "+30m"); got != 150*time.Minute {
		t.Errorf("got %v, want 2h30m", got)
	}
}

func TestSpent_takesOffWhatIsGivenWithAMinus(t *testing.T) {
	if got := spentAfter(t, 2*time.Hour, "-15m"); got != 105*time.Minute {
		t.Errorf("got %v, want 1h45m", got)
	}
}

func TestSpent_stopsAtNothingRatherThanGoingBelow(t *testing.T) {
	if got := spentAfter(t, 2*time.Hour, "-3h"); got != 0 {
		t.Errorf("got %v, want none", got)
	}
}

func TestSpent_clearsWhenGivenNothing(t *testing.T) {
	if got := spentAfter(t, 2*time.Hour, ""); got != 0 {
		t.Errorf("got %v, want none", got)
	}
}

func TestSpent_refusesWhatIsNotADuration(t *testing.T) {
	for _, given := range []string{"ages", "+ages", "-ages", "+", "-"} {
		if _, err := ParseSpent(given); !errors.Is(err, ErrNotADuration) {
			t.Errorf("ParseSpent(%q) err = %v, want ErrNotADuration", given, err)
		}
	}
}

func TestSpent_refusesANegativeTotal(t *testing.T) {
	if _, err := ParseSpent("-"); !errors.Is(err, ErrNotADuration) {
		t.Errorf("err = %v, want ErrNotADuration", err)
	}
}

func TestSpent_knowsWhetherItMovesOrSets(t *testing.T) {
	moving, err := ParseSpent("+30m")
	if err != nil {
		t.Fatalf("ParseSpent: %v", err)
	}
	if got := moving.Apply(time.Hour); got != 90*time.Minute {
		t.Errorf("a signed duration moves what is there: an hour became %v", got)
	}

	setting, err := ParseSpent("2h")
	if err != nil {
		t.Fatalf("ParseSpent: %v", err)
	}
	if got := setting.Apply(time.Hour); got != 2*time.Hour {
		t.Errorf("a plain duration sets the total: an hour became %v", got)
	}
}

func TestSpent_acceptsMovingByNothing(t *testing.T) {
	if got := spentAfter(t, 2*time.Hour, "+0m"); got != 2*time.Hour {
		t.Errorf("got %v, want 2h unchanged", got)
	}
}

func TestSpent_refusesASignedDurationAfterASign(t *testing.T) {
	for _, given := range []string{"+-30m", "-+30m"} {
		if _, err := ParseSpent(given); !errors.Is(err, ErrNotADuration) {
			t.Errorf("ParseSpent(%q) err = %v, want ErrNotADuration", given, err)
		}
	}
}

func TestParseSpent_refusesWhatItCannotWriteBack(t *testing.T) {
	for _, amount := range []string{"500ms", "+500ms", "-500ms", "1m30s500ms"} {
		if _, err := ParseSpent(amount); !errors.Is(err, ErrFinerThanASecond) {
			t.Errorf("ParseSpent(%q) err = %v, want ErrFinerThanASecond", amount, err)
		}
	}
}

func TestParseSpent_keepsWholeSeconds(t *testing.T) {
	for _, amount := range []string{"30s", "+1m", "-2h", "90m"} {
		if _, err := ParseSpent(amount); err != nil {
			t.Errorf("ParseSpent(%q): %v", amount, err)
		}
	}
}
