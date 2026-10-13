package gui

import (
	"math"
)

// Thumb is where that bar's grip sits: how tall it is says how much of the history is showing, and where it sits says which part.
func Thumb(track Box, lines, rows, at int) Box {
	high := thumbHigh(track, lines, rows)
	room := track.H - high
	if room < 1 {
		return track
	}
	down := room * min(at, lines-rows) / (lines - rows)
	return Box{X: track.X, Y: track.Y + down, W: track.W, H: high}
}

func thumbHigh(track Box, lines, rows int) int {
	if lines <= rows {
		return track.H
	}
	return max(leastThumb, track.H*rows/lines)
}

// Reached is which line of the history a point on the track means, so that taking hold of the grip and pulling it moves the history with it.
func Reached(track Box, lines, rows int, p Point) int {
	high := thumbHigh(track, lines, rows)
	room := track.H - high
	if room < 1 {
		return 0
	}
	return (p.Y - track.Y - high/2) * (lines - rows) / room
}

const caretWide = 2

// LineAt is where one line of a wrapped text is drawn, given which line is the first one showing.
func LineAt(inside Box, i, from int) Box {
	return Box{X: inside.X, Y: inside.Y + (i-from)*lineHigh, W: inside.W, H: lineHigh}
}

// Lasting is the line after the last one drawn: the last there is, or the last there is room for.
func Lasting(lines, from, rows int) int { return min(lines, from+rows) }

// HeldOn says which characters of one line a held stretch of text covers, and whether that line carries any of it.
func HeldOn(lines Lines, from, to, i int) (a, b int, on bool) {
	begins, ends := lines.From[i], len([]rune(lines.Said[i]))+lines.From[i]
	a, b = max(from, begins), min(to, ends)
	if a >= b {
		return 0, 0, false
	}
	return a - begins, b - begins, true
}

// CaretAt is where the caret stands: on the line it fell on, after however much room what comes before it takes.
func CaretAt(inside Box, line, from, wide int) Box {
	at := LineAt(inside, line, from)
	return Box{X: at.X + wide, Y: at.Y, W: caretWide, H: at.H}
}

// Lines is a text broken into the lines it is drawn on, and where each of them began in the text — which is what puts the caret on the line it fell on, since the caret is counted in the whole text and not in one line of it.
type Lines struct {
	Said []string
	From []int
}

// Wrapped breaks a text into the lines it takes in a given room, counted in characters: what Measure.Wrapped does in a face where every character is one wide.
func Wrapped(said string, room int) Lines {
	return Measure(characters).Wrapped(said, room, 0)
}

func characters(line string, _ int) []int {
	edges := make([]int, len([]rune(line))+1)
	for i := range edges {
		edges[i] = i
	}
	return edges
}

// Wrapped breaks a text into the lines it takes in a given width, measured in the face it is written in, at a space when there is one to break at, and inside a word when the word is wider than the room.
func (m Measure) Wrapped(said string, wide, size int) Lines {
	kept := []rune(said)
	if wide < 1 || len(kept) == 0 {
		return Lines{Said: []string{said}, From: []int{0}}
	}
	lines := Lines{}
	from := 0
	for from < len(kept) {
		stop := next(kept, from, len(kept))
		if stop < 0 {
			stop = len(kept)
		}
		edges := m(string(kept[from:stop]), size)
		if edges[len(edges)-1] <= wide {
			lines.Said = append(lines.Said, string(kept[from:stop]))
			lines.From = append(lines.From, from)
			from = stop + 1
			continue
		}
		end := from + fitting(edges, wide)
		for i := end - 1; i > from; i-- {
			if kept[i] == ' ' {
				end = i
				break
			}
		}
		lines.Said = append(lines.Said, string(kept[from:end]))
		lines.From = append(lines.From, from)
		from = end
		if from < len(kept) && kept[from] == ' ' {
			from++
		}
	}
	if kept[len(kept)-1] == '\n' {
		lines.Said = append(lines.Said, "")
		lines.From = append(lines.From, len(kept))
	}
	return lines
}

func fitting(edges []int, wide int) int {
	fit := 1
	for i := 2; i < len(edges); i++ {
		if edges[i] <= wide {
			fit = i
		}
	}
	return fit
}

func next(kept []rune, from, end int) int {
	for i := from; i < end; i++ {
		if kept[i] == '\n' {
			return i
		}
	}
	return -1
}

// At says which line the caret fell on and how far into that line it is.
func (l Lines) At(caret int) (line, column int) {
	for i, from := range l.From {
		if caret >= from {
			line = i
		}
	}
	column = min(len([]rune(l.Said[line])), max(0, caret-l.From[line]))
	return line, column
}

// Window is the first of the lines shown when there are more of them than there is room for: the last screenful ending on the caret's line, because what is being typed is what has to stay in sight.
func (l Lines) Window(rows, line int) int {
	return max(0, line-rows+1)
}

func room(b Box) int { return (b.W - padding - padding) / perGlyph }

// Follow is the scroll a field is given when what is in sight is wherever its caret is, rather than wherever a wheel left it.
const Follow = math.MinInt

// StartsAt is the first line of a text in sight: where the wheel left it, or, for a field that follows its caret, whatever window holds the caret's line.
func StartsAt(lines Lines, rows, caretLine, scrolled int) int {
	if scrolled == Follow {
		return lines.Window(rows, caretLine)
	}
	return scrolled
}
