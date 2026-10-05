package gui

import (
	"math"
)

// Where text goes inside a box: the lines it wraps into, the caret, what is
// held, and the bar a box scrolls by.

// Thumb is where that bar's grip sits: how tall it is says how much of the
// history is showing, and where it sits says which part.
//
// A history that fits gets a grip filling the whole track — the bar is still
// drawn, because a bar that comes and goes is a bar somebody reaches for and
// does not find.
func Thumb(track Box, lines, rows, at int) Box {
	high := thumbHigh(track, lines, rows)
	room := track.H - high
	if room < 1 {
		return track
	}
	down := room * min(at, lines-rows) / (lines - rows)
	return Box{X: track.X, Y: track.Y + down, W: track.W, H: high}
}

// thumbHigh is how tall that grip is: in proportion to how much of the history
// is showing, never smaller than can be taken hold of, and the whole track
// when there is nothing to scroll.
//
// One height and not two. Thumb and Reached each worked it out, and worked it
// out differently — a grip drawn one size and dragged as though it were
// another is a grip that jumps under the hand.
func thumbHigh(track Box, lines, rows int) int {
	if lines <= rows {
		return track.H
	}
	return max(leastThumb, track.H*rows/lines)
}

// Reached is which line of the history a point on the track means, so that
// taking hold of the grip and pulling it moves the history with it.
func Reached(track Box, lines, rows int, p Point) int {
	high := thumbHigh(track, lines, rows)
	room := track.H - high
	if room < 1 {
		return 0
	}
	// Past that guard there is more history than room, so the division below
	// has something to divide by.
	return (p.Y - track.Y - high/2) * (lines - rows) / room
}

// caretWide is how thick the caret is drawn. Two, because one disappears on a
// screen with two pixels to the dp.
const caretWide = 2

// LineAt is where one line of a wrapped text is drawn, given which line is the
// first one showing.
func LineAt(inside Box, i, from int) Box {
	return Box{X: inside.X, Y: inside.Y + (i-from)*lineHigh, W: inside.W, H: lineHigh}
}

// Lasting is the line after the last one drawn: the last there is, or the last
// there is room for.
func Lasting(lines, from, rows int) int { return min(lines, from+rows) }

// HeldOn says which characters of one line a held stretch of text covers, and
// whether that line carries any of it. Counted in characters and not in
// pixels: how wide a character is, is a question for the font, and it is asked
// where the caret asks it — a run measured at so many pixels a glyph sat a
// word away from the caret at its own end.
func HeldOn(lines Lines, from, to, i int) (a, b int, on bool) {
	begins, ends := lines.From[i], len([]rune(lines.Said[i]))+lines.From[i]
	a, b = max(from, begins), min(to, ends)
	if a >= b {
		return 0, 0, false
	}
	return a - begins, b - begins, true
}

// CaretAt is where the caret stands: on the line it fell on, after however
// much room what comes before it takes.
func CaretAt(inside Box, line, from, wide int) Box {
	at := LineAt(inside, line, from)
	return Box{X: at.X + wide, Y: at.Y, W: caretWide, H: at.H}
}

// Lines is a text broken into the lines it is drawn on, and where each of them
// began in the text — which is what puts the caret on the line it fell on,
// since the caret is counted in the whole text and not in one line of it.
type Lines struct {
	Said []string
	From []int
}

// Wrapped breaks a text into the lines it takes in a given room, at a space
// when there is one to break at, and inside a word when the word is longer
// than the room.
//
// Broken rather than cut: a description that runs out of its box is a
// description nobody can read back, and cutting it says nothing about what was
// lost. A word longer than the room is broken inside itself, because the only
// other answer is a line that runs out of the box — the very thing this
// prevents.
func Wrapped(said string, room int) Lines {
	kept := []rune(said)
	// No room is answered whole rather than one character at a time, and
	// nothing is answered as one empty line — a text of no length would
	// otherwise come back as no lines at all, and the caret has to fall
	// somewhere.
	if room < 1 || len(kept) == 0 {
		return Lines{Said: []string{said}, From: []int{0}}
	}
	lines := Lines{}
	from := 0
	for from < len(kept) {
		end := min(len(kept), from+room)
		// A line break is a line, before the width is ever consulted. Answered
		// only by the width, a description written in paragraphs came back as
		// one run of words with the breaks still in it, drawn as squares.
		if at := next(kept, from, end); at >= 0 {
			lines.Said = append(lines.Said, string(kept[from:at]))
			lines.From = append(lines.From, from)
			from = at + 1
			continue
		}
		if end < len(kept) {
			// Back to the last space there is room for — but never as far as
			// the line's own beginning, which would give an empty line and a
			// text that never gets any shorter.
			for i := end - 1; i > from; i-- {
				if kept[i] == ' ' {
					end = i
					break
				}
			}
		}
		lines.Said = append(lines.Said, string(kept[from:end]))
		lines.From = append(lines.From, from)
		from = end
		// The space a line was broken at belongs to neither line.
		if from < len(kept) && kept[from] == ' ' {
			from++
		}
	}
	// A text ending in a break has one more line, empty, for the caret to
	// stand on. Without it the caret stayed drawn at the end of the line
	// above and only came down once something had been typed there, so ⇧⏎
	// looked as though it had done nothing at all.
	if kept[len(kept)-1] == '\n' {
		lines.Said = append(lines.Said, "")
		lines.From = append(lines.From, len(kept))
	}
	return lines
}

// next is where the first line break between two places is, or that there is
// none.
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

// Window is the first of the lines shown when there are more of them than
// there is room for: the last screenful ending on the caret's line, because
// what is being typed is what has to stay in sight.
//
// It was also clamped against the lines there are, and the mutation gate
// showed that clamp could never win — the caret's line is always inside the
// text, so the run ending on it always ends inside the text too. Arithmetic
// nothing can reach is arithmetic nothing can judge.
func (l Lines) Window(rows, line int) int {
	return max(0, line-rows+1)
}

// room is how many characters a box can show, reckoned the way a title is.
func room(b Box) int { return (b.W - padding - padding) / perGlyph }

// Follow is the scroll a field is given when what is in sight is wherever its
// caret is, rather than wherever a wheel left it. A value no scroll can be: it
// was -1, and one character from 1, which is a line a wheel can leave a text
// at — so a slip there would have read a scrolled field as one following its
// caret, and no test could have reached a constant to say so.
const Follow = math.MinInt

// StartsAt is the first line of a text in sight: where the wheel left it, or,
// for a field that follows its caret, whatever window holds the caret's line.
func StartsAt(lines Lines, rows, caretLine, scrolled int) int {
	if scrolled == Follow {
		return lines.Window(rows, caretLine)
	}
	return scrolled
}
