package domain

import (
	"errors"
	"fmt"
	"strings"
)

// The reasons a piece of a text cannot be replaced.
var (
	ErrReplacedNowhere      = errors.New("the text to replace is not in the description")
	ErrReplacedMoreThanOnce = errors.New("the text to replace has to be there once; give more of it")
)

// ReplaceOnce puts what replaces a piece of a text in its place, and refuses unless the piece is there exactly once.
func ReplaceOnce(text, old, by string) (string, error) {
	found := 0
	if old != "" {
		found = strings.Count(text, old)
	}
	if found == 0 {
		return "", ErrReplacedNowhere
	}
	if found > 1 {
		return "", fmt.Errorf("the text to replace is there %d times: %w", found, ErrReplacedMoreThanOnce)
	}
	return strings.Replace(text, old, by, 1), nil
}
