package gui

import (
	"cairn/internal/gui/geom"
	"image/color"
	"strings"

	"gioui.org/font"
	"gioui.org/text"
)

const capGap = 3

// Sentence is a line of words, the keys it names between backticks, each drawn on its cap; small and faint unless told otherwise.
type Sentence struct {
	Said  string
	Ink   color.NRGBA
	Size  int
	Along text.Alignment
}

// Plain is the sentence as it reads.
func (s Sentence) Plain() string { return strings.ReplaceAll(s.Said, "`", "") }

// SentenceWide is how wide a sentence is drawn at a size, its keys on their caps.
func SentenceWide(m Measure, said string, size int) int {
	wide := 0
	for i, run := range strings.Split(said, "`") {
		if i%2 == 1 {
			wide += KeyChipWide(run)
			continue
		}
		wide += ends(m, run, size)
	}
	return wide
}

// Draw writes the sentence along at, its keys on their caps, and stops where at ends.
func (s Sentence) Draw(k Kit, at geom.Box) geom.Box {
	ink, size := s.Ink, s.Size
	if ink == (color.NRGBA{}) {
		ink = k.skin.Colour.Ink.Faint
	}
	if size == 0 {
		size = small
	}
	x, end := at.X, at.X+at.W
	switch room := at.W - SentenceWide(k.measured(), s.Said, size); {
	case room <= 0:
	case s.Along == text.Middle:
		x += room / 2
	case s.Along == text.End:
		x += room
	}
	for i, run := range strings.Split(s.Said, "`") {
		if run == "" || x >= end {
			continue
		}
		if i%2 == 1 {
			if x+KeyChipWide(run) > end {
				x = end
				continue
			}
			x += KeyChip{Key: run}.Draw(k, geom.Box{X: x, Y: at.Y, H: at.H}).W
			continue
		}
		sayIn(k.gtx, k.th, ink, geom.Box{X: x, Y: at.Y, W: end - x, H: at.H}, run, size, font.Font{}, text.Start)
		x += ends(k.measured(), run, size)
	}
	return at
}
