package cli

import (
	"errors"
	"fmt"
	"testing"

	"cairn/internal/domain"
	"cairn/internal/engine"
)

func TestCodeFor_saysNothingWentWrong(t *testing.T) {
	if got := codeFor(nil); got != exitOK {
		t.Errorf("codeFor(nil) = %d, want %d", got, exitOK)
	}
}

func TestCodeFor_tellsAStoreItCannotUse(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", engine.ErrStore)
	if got := codeFor(err); got != exitStore {
		t.Errorf("codeFor = %d, want %d for a store it cannot use", got, exitStore)
	}
}

func TestCodeFor_tellsARequestItUnderstoodAndRefused(t *testing.T) {
	for name, err := range map[string]error{
		"no title":      domain.ErrTitleRequired,
		"no label":      domain.ErrLabelRequired,
		"not a column":  domain.ErrNotAState,
		"no such task":  engine.ErrNoSuchTask,
		"nothing named": engine.ErrNothingToChange,
	} {
		if got := codeFor(fmt.Errorf("wrapped: %w", err)); got != exitRefused {
			t.Errorf("codeFor(%s) = %d, want %d", name, got, exitRefused)
		}
	}
}

func TestCodeFor_tellsARequestItDidNotUnderstand(t *testing.T) {
	if got := codeFor(errors.New("unknown flag: --colour")); got != exitUsage {
		t.Errorf("codeFor = %d, want %d for something it could not parse", got, exitUsage)
	}
}
