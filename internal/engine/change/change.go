// Package change holds the grammar of a change: +x -x =x, more or less time, replace … by ….
package change

import (
	"strings"
)

const (
	joins    = '+'
	parts    = '-'
	replaces = "="
)

// Listed splits what was given at its commas, one piece each.
func Listed(given []string) []string {
	pieces := make([]string, 0, len(given))
	for _, one := range given {
		for _, piece := range strings.Split(one, ",") {
			if strings.TrimSpace(piece) == "" && piece != one {
				continue
			}
			pieces = append(pieces, piece)
		}
	}
	return pieces
}

// SplitSign takes the leading + or - off a value, if it carries one, and hands back the sign and the rest.
func SplitSign(value string) (rune, string) {
	if value == "" {
		return 0, value
	}
	first := rune(value[0])
	if first == joins || first == parts {
		return first, value[1:]
	}
	return 0, value
}
