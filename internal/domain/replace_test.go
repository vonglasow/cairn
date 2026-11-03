package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestReplaceOnce_changesThePieceAndLeavesTheRest(t *testing.T) {
	got, err := ReplaceOnce("built on main at abc123, tested", "abc123", "def456")
	if err != nil || got != "built on main at def456, tested" {
		t.Errorf("ReplaceOnce = %q, %v", got, err)
	}
	if got, err := ReplaceOnce("built on main, tested", ", tested", ""); err != nil || got != "built on main" {
		t.Errorf("replaced by nothing = %q, %v", got, err)
	}
}

func TestReplaceOnce_refusesAPieceThatIsNotThere(t *testing.T) {
	for _, old := range []string{"develop", ""} {
		if _, err := ReplaceOnce("built on main", old, "x"); !errors.Is(err, ErrReplacedNowhere) {
			t.Errorf("%q: err = %v, want ErrReplacedNowhere", old, err)
		}
	}
}

func TestReplaceOnce_refusesAPieceThereMoreThanOnceSayingHowOften(t *testing.T) {
	_, err := ReplaceOnce("main into main into main", "main", "develop")
	if !errors.Is(err, ErrReplacedMoreThanOnce) || !strings.Contains(err.Error(), "there 3 times") {
		t.Errorf("err = %v, want ErrReplacedMoreThanOnce saying 3 times", err)
	}
}
