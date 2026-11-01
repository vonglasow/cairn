package domain

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// LabelColours are the colours a label can wear, by name: Catppuccin Mocha's fourteen accents, made to be read on a dark ground and all of one weight, so that no label shouts over another.
var LabelColours = []string{
	"rosewater", "flamingo", "pink", "mauve", "red", "maroon", "peach",
	"yellow", "green", "teal", "sky", "sapphire", "blue", "lavender",
}

// ErrNoSuchColour is returned when a colour named is none of the label colours.
var ErrNoSuchColour = errors.New("a label's colour is one of " + strings.Join(LabelColours, ", "))

// ParseColour reads a colour's name the one way it is written.
func ParseColour(given string) (string, error) {
	name := strings.ToLower(strings.TrimSpace(given))
	if !slices.Contains(LabelColours, name) {
		return "", fmt.Errorf("%q: %w", given, ErrNoSuchColour)
	}
	return name, nil
}

// ColourOf is the colour a label wears when none was chosen for it: one of the fourteen, read off its name, so the same label wears the same colour wherever it is drawn.
func ColourOf(label string) string {
	sum := uint32(2166136261)
	for _, b := range []byte(label) {
		sum = (sum ^ uint32(b)) * 16777619
	}
	return LabelColours[sum%uint32(len(LabelColours))]
}
