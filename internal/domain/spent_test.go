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

// Time spent is a tally of work done. Taking off more than was ever put on
// leaves nothing — a negative tally is not a smaller amount of work, it is a
// number that would travel all the way to the JSON and into any sum.
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
	if !moving.Moves() {
		t.Error("a signed duration moves what is there")
	}

	setting, err := ParseSpent("2h")
	if err != nil {
		t.Fatalf("ParseSpent: %v", err)
	}
	if setting.Moves() {
		t.Error("a plain duration sets the total, it does not move it")
	}
}

// Adding nothing is a thing somebody can ask for, and it changes nothing. The
// boundary matters: refusing zero would make `+0m` an error rather than a
// no-op.
func TestSpent_acceptsMovingByNothing(t *testing.T) {
	if got := spentAfter(t, 2*time.Hour, "+0m"); got != 2*time.Hour {
		t.Errorf("got %v, want 2h unchanged", got)
	}
}

// A sign already says which way it goes, so a second one inside the duration
// is somebody saying two things at once.
func TestSpent_refusesASignedDurationAfterASign(t *testing.T) {
	for _, given := range []string{"+-30m", "-+30m"} {
		if _, err := ParseSpent(given); !errors.Is(err, ErrNotADuration) {
			t.Errorf("ParseSpent(%q) err = %v, want ErrNotADuration", given, err)
		}
	}
}
