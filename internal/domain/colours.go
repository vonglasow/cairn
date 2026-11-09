package domain

import (
	"fmt"
	"slices"
	"strings"
)

// LabelColours are the colours a label can wear, by name: Catppuccin Mocha's fourteen accents, made to be read on a dark ground and all of one weight, so that no label shouts over another.
var LabelColours = []string{
	"rosewater", "flamingo", "pink", "mauve", "red", "maroon", "peach",
	"yellow", "green", "teal", "sky", "sapphire", "blue", "lavender",
}

// LabelInks are the label colours as Catppuccin Mocha inks them, red, green and blue, for every surface that draws one.
var LabelInks = map[string][3]uint8{
	"rosewater": {0xf5, 0xe0, 0xdc}, "flamingo": {0xf2, 0xcd, 0xcd},
	"pink": {0xf5, 0xc2, 0xe7}, "mauve": {0xcb, 0xa6, 0xf7},
	"red": {0xf3, 0x8b, 0xa8}, "maroon": {0xeb, 0xa0, 0xac},
	"peach": {0xfa, 0xb3, 0x87}, "yellow": {0xf9, 0xe2, 0xaf},
	"green": {0xa6, 0xe3, 0xa1}, "teal": {0x94, 0xe2, 0xd5},
	"sky": {0x89, 0xdc, 0xeb}, "sapphire": {0x74, 0xc7, 0xec},
	"blue": {0x89, 0xb4, 0xfa}, "lavender": {0xb4, 0xbe, 0xfe},
}

// ErrNoSuchColour is returned when a colour named is none of the label colours.
var ErrNoSuchColour = fmt.Errorf("a label's colour is one of %s", strings.Join(LabelColours, ", "))

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
