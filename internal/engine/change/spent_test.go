package change

import (
	"cairn/internal/engine/calendar"
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

func TestSpent_clearsWhenGivenNothing(t *testing.T) {
	if got := spentAfter(t, 2*time.Hour, ""); got != 0 {
		t.Errorf("got %v, want none", got)
	}
}

func TestSpent_refusesASignedDurationAfterASign(t *testing.T) {
	for _, given := range []string{"+-30m", "-+30m"} {
		if _, err := ParseSpent(given); !errors.Is(err, calendar.ErrNotADuration) {
			t.Errorf("ParseSpent(%q) err = %v, want ErrNotADuration", given, err)
		}
	}
}

func TestParseSpent_refusesWhatItCannotWriteBack(t *testing.T) {
	for _, amount := range []string{"500ms", "+500ms", "-500ms", "1m30s500ms"} {
		if _, err := ParseSpent(amount); !errors.Is(err, calendar.ErrFinerThanASecond) {
			t.Errorf("ParseSpent(%q) err = %v, want ErrFinerThanASecond", amount, err)
		}
	}
}
