package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

const ellipsis = "…"

// Clip cuts a string to a display width, marking it when something was left out.
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

// Wrap lays a title out over a fixed number of lines of a given width, cutting at a word where it can and marking the last line when the title went on.
func Wrap(s string, width, lines int) []string {
	out := make([]string, lines)
	if lines <= 0 {
		return out
	}
	words := strings.Fields(s)
	at := 0
	for _, word := range words {
		if at >= lines {
			break
		}
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
	for i := range out {
		out[i] = Clip(out[i], width)
	}
	if at >= lines {
		out[lines-1] = Clip(out[lines-1]+" "+ellipsis, width)
	}
	return out
}

// Spread puts one string against the left of a width and another against the right, with at least one space between them.
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

// Paragraphs lays a text out on lines of a width, every paragraph of it: a break somebody wrote is a line, and nothing is cut short.
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

// The two marks a bar is drawn with down the right of a text that scrolls: the track the whole height, and the thumb where the window is.
const (
	Track = "│"
	Thumb = "┃"
)

// InSight is what of a laid-out text is in a box of so many rows, starting at a line, each line padded to a width and carrying the bar down its right.
func InSight(lines []string, from, rows, width int) []string {
	from = min(max(0, from), max(0, len(lines)-rows))
	shown := lines[from:min(len(lines), from+rows)]
	if len(lines) <= rows {
		return marked(shown, width, 0, len(shown))
	}
	long := max(1, rows*rows/len(lines))
	at := from * (rows - long) / (len(lines) - rows)
	return marked(shown, width, at, long)
}

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
