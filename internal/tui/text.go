package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Cutting text to a width. A card is twenty characters wide at a hundred
// columns and ten at sixty, so nothing here may assume a character is a cell:
// what is counted is what the terminal draws, not how many runes went in.

// ellipsis says a title went on. One character, so it is paid for out of the
// title rather than out of the card.
const ellipsis = "…"

// Clip cuts a string to a display width, marking it when something was left
// out. A width of nothing at all answers nothing at all rather than an
// ellipsis: a mark with no text beside it says a card carries something
// unreadable, when in fact the card has no room for anything.
func Clip(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	var out strings.Builder
	room := width - lipgloss.Width(ellipsis)
	for _, r := range s {
		if lipgloss.Width(out.String()+string(r)) > room {
			break
		}
		out.WriteRune(r)
	}
	return strings.TrimRight(out.String(), " ") + ellipsis
}

// Wrap lays a title out over a fixed number of lines of a given width, cutting
// at a word where it can and marking the last line when the title went on.
//
// The number of lines is fixed rather than however many it takes, because a
// card is six rows whatever it holds — a title that took a third line would
// push a card's facts out of sight, and every card below it out of place.
func Wrap(s string, width, lines int) []string {
	out := make([]string, lines)
	// Only the number of lines is guarded. A width of nothing needs no guard of
	// its own: Clip below answers nothing for it, line by line.
	if lines <= 0 {
		return out
	}
	words := strings.Fields(s)
	at := 0
	for _, word := range words {
		if at >= lines {
			break
		}
		// Ifs rather than a tagless switch, for the reason given in view.go.
		if out[at] == "" {
			out[at] = word
			continue
		}
		if lipgloss.Width(out[at]+" "+word) <= width {
			out[at] += " " + word
			continue
		}
		at++
		if at < lines {
			out[at] = word
		}
	}
	// A single word longer than the line is not a wrapping problem, so the loop
	// above leaves it whole. Cutting happens here, on every line at once, which
	// also catches the last line of a title that had more to say.
	for i := range out {
		out[i] = Clip(out[i], width)
	}
	if at >= lines {
		out[lines-1] = Clip(out[lines-1]+" "+ellipsis, width)
	}
	return out
}

// Spread puts one string against the left of a width and another against the
// right, with at least one space between them. If they will not both fit, the
// right one wins: it is a mark, and a mark that is only sometimes there is a
// mark nobody can rely on.
func Spread(left, right string, width int) string {
	room := width - lipgloss.Width(right)
	if room < 1 {
		return Clip(right, width)
	}
	left = Clip(left, room-1)
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	return left + strings.Repeat(" ", gap) + right
}

// Join puts things side by side while they fit, and stops when they do not.
// What is dropped is what came last, so the order they are given in is the
// order they matter in.
//
// It is what keeps a card's facts inside its card and the bar inside the
// screen. Both would otherwise wrap, and a wrapped bar pushes the board up out
// of the window.
func Join(sep string, width int, facts ...string) string {
	var out string
	for _, fact := range facts {
		if fact == "" {
			continue
		}
		next := fact
		if out != "" {
			next = out + sep + fact
		}
		if lipgloss.Width(next) > width {
			break
		}
		out = next
	}
	return out
}

// Paragraphs lays a text out on lines of a width, every paragraph of it: a
// break somebody wrote is a line, and nothing is cut short. Wrap is for a card,
// which has two lines and a word to spare; this is for a text that scrolls.
func Paragraphs(s string, width int) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, paragraph := range strings.Split(s, "\n") {
		lines := Wrap(paragraph, width, max(1, len([]rune(paragraph))))
		kept := 0
		for i, line := range lines {
			if line != "" {
				kept = i + 1
			}
		}
		out = append(out, lines[:max(1, kept)]...)
	}
	return out
}

// The two marks a bar is drawn with down the right of a text that scrolls: the
// track the whole height, and the thumb where the window is.
const (
	Track = "│"
	Thumb = "┃"
)

// InSight is what of a laid-out text is in a box of so many rows, starting at
// a line, each line padded to a width and carrying the bar down its right. The
// bar is drawn whether or not there is anything to scroll, the rule the window
// goes by: a bar that comes and goes is a bar somebody reaches for and does
// not find.
func InSight(lines []string, from, rows, width int) []string {
	from = min(max(0, from), max(0, len(lines)-rows))
	shown := lines[from:min(len(lines), from+rows)]
	if len(lines) <= rows {
		return marked(shown, width, 0, len(shown))
	}
	long := max(1, rows*rows/len(lines))
	// Spread over the room the thumb leaves, end to end: the window at the
	// last line puts the thumb at the bottom of the track. Spread over the
	// whole track it never got there, and a bar that never reaches the bottom
	// says there is more to come when there is not.
	at := from * (rows - long) / (len(lines) - rows)
	return marked(shown, width, at, long)
}

// marked pads each line in sight to a width and puts the bar down its right:
// the thumb from a row, so many rows long, and the track everywhere else.
func marked(shown []string, width, at, long int) []string {
	out := make([]string, len(shown))
	for i, line := range shown {
		mark := Track
		if i >= at && i < at+long {
			mark = Thumb
		}
		out[i] = Clip(line, width) + strings.Repeat(" ", max(0, width-lipgloss.Width(Clip(line, width)))) + " " + mark
	}
	return out
}
