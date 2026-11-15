package label

import (
	"cairn/internal/engine/change"
	"cairn/internal/engine/item"
	"errors"
	"fmt"
	"slices"
	"strings"
	"unicode"
)

// What a label name given can be turned down for.
var (
	ErrLabelRequired    = errors.New("a label needs a name")
	ErrLabelNotOneLine  = errors.New("a label is one word or a few, on one line")
	ErrLabelHasNoFamily = errors.New("a label has no family: write two labels, one each side of the colon")
	ErrLabelHolds       = errors.New("a label holds only a to z, 0 to 9, dashes and dots")
)

// ParseLabels turns what was given into the set a task carries: lower case, no duplicate, alphabetical order.
func ParseLabels(given []string) ([]string, error) {
	given = change.Listed(given)
	kept := make([]string, 0, len(given))
	for _, label := range given {
		if label == "" {
			continue
		}
		name, err := labelNamed(label)
		if err != nil {
			return nil, err
		}
		kept = append(kept, name)
	}
	slices.Sort(kept)
	return slices.Compact(kept), nil
}

// Folded is a name read the way a label is written — lower case, no accent, words joined by a dash — so two names a person would read as one compare equal.
func Folded(name string) string {
	return dashed(unaccented(strings.ToLower(strings.TrimSpace(name))))
}

func labelNamed(raw string) (string, error) {
	if strings.ContainsFunc(strings.TrimSpace(raw), item.BreaksALine) {
		return "", fmt.Errorf("%q: %w", raw, ErrLabelNotOneLine)
	}
	name := Folded(raw)
	if name == "" {
		return "", ErrLabelRequired
	}
	if strings.ContainsRune(name, ':') {
		return "", fmt.Errorf("%q: %w", raw, ErrLabelHasNoFamily)
	}
	if strings.ContainsFunc(name, notInALabel) {
		return "", fmt.Errorf("%q: %w", raw, ErrLabelHolds)
	}
	return name, nil
}

var unaccented = strings.NewReplacer(
	"à", "a", "á", "a", "â", "a", "ã", "a", "ä", "a", "å", "a",
	"ç", "c",
	"è", "e", "é", "e", "ê", "e", "ë", "e",
	"ì", "i", "í", "i", "î", "i", "ï", "i",
	"ñ", "n",
	"ò", "o", "ó", "o", "ô", "o", "õ", "o", "ö", "o",
	"ù", "u", "ú", "u", "û", "u", "ü", "u",
	"ý", "y", "ÿ", "y",
).Replace

func dashed(word string) string {
	pieces := strings.FieldsFunc(word, func(r rune) bool {
		return r == '-' || r == '_' || r == '/' || unicode.IsSpace(r)
	})
	return strings.Join(pieces, "-")
}

func notInALabel(r rune) bool {
	return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' && r != '.'
}
