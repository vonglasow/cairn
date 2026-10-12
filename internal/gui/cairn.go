package gui

// The window opens on a cairn being built: five stones falling one after the
// other onto the one below. A board that appears all at once is a tool; one
// that opens on the thing it is named after is this tool. It lasts a breath,
// starts on the first frame a screen draws — never in a scenario, which draws
// none — and the first key or click ends it and still does what it does.
// Drawn rather than shaded: Gio has no shader of a program's own, and five
// ellipses say it. See docs/gui.md.

import "time"

// IntroLasts is how long the cairn takes to build.
func IntroLasts() time.Duration { return 1400 * time.Millisecond }

// stones is each stone's width and height, the base first.
var stones = [5][2]int{{160, 44}, {130, 38}, {100, 32}, {72, 26}, {46, 20}}

// Born is the window told the moment of its first frame, which is when the
// cairn starts building. Told once: a later frame is not a birth.
func (w Window) Born(at time.Time) Window {
	if w.born.IsZero() {
		w.born, w.now = at, at
	}
	return w
}

// Intro is how far the cairn is built, from nought to one, and whether it is
// being built at all: not before the first frame, not once built, not once
// anything was pressed.
func (w Window) Intro() (float64, bool) {
	if w.born.IsZero() || w.introOver {
		return 0, false
	}
	progress := float64(w.now.Sub(w.born)) / float64(IntroLasts())
	if progress >= 1 {
		return 1, false
	}
	return max(0, progress), true
}

// interrupted ends the cairn: somebody has something to do.
func (w Window) interrupted() Window {
	w.introOver = true
	return w
}

// StonesAt is where each stone is, a share of the way through the build: each
// falls in its turn from above the window onto the one below it, slowing as
// it lands, and the stack stands on the middle of the window.
func StonesAt(width, height int, progress float64) [len(stones)]Box {
	var at [len(stones)]Box
	bottom := height/2 + 80
	for i, size := range stones {
		bottom -= size[1]
		rest := Box{X: (width - size[0]) / 2, Y: bottom, W: size[0], H: size[1]}
		// Each stone falls over two fifths of the build, a stone's turn after
		// the one below it.
		start := float64(i) * 0.15
		share := min(1, max(0, (progress-start)/0.4))
		fallen := 1 - (1-share)*(1-share)
		top := -size[1]
		rest.Y = top + int(float64(rest.Y-top)*fallen)
		at[i] = rest
	}
	return at
}
